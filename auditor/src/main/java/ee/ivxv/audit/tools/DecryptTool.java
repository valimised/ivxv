package ee.ivxv.audit.tools;

import ee.ivxv.audit.AuditContext;
import ee.ivxv.audit.Msg;
import ee.ivxv.audit.tools.DecryptTool.DecryptArgs;
import ee.ivxv.audit.util.InvalidDecProofs;
import ee.ivxv.common.M;
import ee.ivxv.common.cli.Arg;
import ee.ivxv.common.cli.Args;
import ee.ivxv.common.cli.Tool;
import ee.ivxv.common.crypto.Plaintext;
import ee.ivxv.common.crypto.elgamal.ElGamalCiphertext;
import ee.ivxv.common.crypto.elgamal.ElGamalDecryptionProof;
import ee.ivxv.common.crypto.elgamal.ElGamalPublicKey;
import ee.ivxv.common.math.MathException;
import ee.ivxv.common.model.Proof;
import ee.ivxv.common.service.bbox.impl.BboxHelperImpl;
import ee.ivxv.common.service.console.Progress;
import ee.ivxv.common.util.I18nConsole;
import ee.ivxv.common.util.ToolHelper;
import ee.ivxv.common.util.Util;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.concurrent.ArrayBlockingQueue;
import java.util.concurrent.Callable;
import java.util.concurrent.CompletionService;
import java.util.concurrent.ExecutorCompletionService;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.RejectedExecutionException;
import java.util.concurrent.ThreadPoolExecutor;
import java.util.concurrent.TimeUnit;
import java.util.function.Consumer;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class DecryptTool implements Tool.Runner<DecryptArgs> {
    private final Logger log = LoggerFactory.getLogger(DecryptTool.class);

    private final AuditContext ctx;
    private final I18nConsole console;
    private final ToolHelper tool;

    public DecryptTool(AuditContext ctx) {
        this.ctx = ctx;
        this.console = new I18nConsole(ctx.i.console, ctx.i.i18n);
        tool = new ToolHelper(console, ctx.container, new BboxHelperImpl(ctx.conf, ctx.container));
    }

    @Override
    public boolean run(DecryptArgs args) throws Exception {
        console.println();
        console.println(Msg.m_pub_loading, args.pubPath.value());
        String keyString = new String(Files.readAllBytes(args.pubPath.value()), Util.CHARSET);
        ElGamalPublicKey pub = new ElGamalPublicKey(Util.decodePublicKey(keyString));
        console.println(Msg.m_pub_loaded);

        Proof proofs = tool.readJsonProofs(args.inputPath.value());

        console.println();
        console.println(Msg.m_verify_start);
        InvalidDecProofs invalid = verifyDecryption(proofs, pub, ctx.args.threads.value());
        console.println(Msg.m_verify_finish);
        console.println(Msg.m_failurecount, invalid.getCount());

        if (invalid.getCount() != 0) {
            console.println(M.m_out_start, args.outputPath.value());
            Files.createDirectory(args.outputPath.value());
            invalid.outputInvalidProofs(args.outputPath.value());
            console.println(M.m_out_done, args.outputPath.value());
        }

        return true;
    }

    private InvalidDecProofs verifyDecryption(Proof input, ElGamalPublicKey pub, int threadCount)
            throws Exception {
        InvalidDecProofs idp = new InvalidDecProofs(input.getElection());
        ExecutorService ioExecutor = Executors.newFixedThreadPool(2);
        CompletionService<Void> CompService = new ExecutorCompletionService<>(ioExecutor);

        ExecutorService verifyExecutor;
        if (threadCount == 0) {
            verifyExecutor = Executors.newCachedThreadPool();
        } else {
            verifyExecutor = new ThreadPoolExecutor(threadCount, threadCount, 0L,
                    TimeUnit.MILLISECONDS, new ArrayBlockingQueue<>(threadCount * 2));
        }

        WorkManager manager =
                new WorkManager(input, getVerifyConsumer(pub, idp), verifyExecutor, idp);
        CompService.submit(manager);
        CompService.submit(idp.getResultWorker());

        try {
            for (int done = 0; done < 2; done++) {
                CompService.take().get();
            }
        } finally {
            ioExecutor.shutdown();
            verifyExecutor.shutdown();
        }
        return idp;
    }

    private Consumer<Proof.ProofJson> getVerifyConsumer(ElGamalPublicKey pub,
            InvalidDecProofs out) {
        return (proofJson) -> {
            Plaintext pt = new Plaintext(proofJson.getMessage());
            ElGamalCiphertext ct =
                    new ElGamalCiphertext(pub.getParameters(), proofJson.getCiphertext());
            ElGamalDecryptionProof proof =
                    new ElGamalDecryptionProof(ct, pt, pub, proofJson.getProof());
            try {
                boolean res = proof.verifyProof();
                if (!res) {
                    log.warn("Proof verification failed: {}", proof);
                    out.addInvalidProof(proof);
                }
            } catch (MathException e) {
                log.warn("Proof verification exception: {}, {}", proof, e);
                out.addInvalidProof(proof);
            }
        };
    }

    public static class DecryptArgs extends Args {
        Arg<Path> inputPath = Arg.aPath(Msg.arg_input);
        Arg<Path> pubPath = Arg.aPath(Msg.arg_pub);
        Arg<Path> outputPath = Arg.aPath(Msg.arg_out, false, true);

        public DecryptArgs() {
            super();
            args.add(inputPath);
            args.add(pubPath);
            args.add(outputPath);
        }
    }

    private class WorkManager implements Callable<Void> {
        private final Proof in;
        private final Consumer<Proof.ProofJson> consumer;
        private final ExecutorService verifyExecutor;
        private final InvalidDecProofs idp;

        WorkManager(Proof in, Consumer<Proof.ProofJson> consumer, ExecutorService verifyExecutor,
                InvalidDecProofs idp) {
            this.in = in;
            this.consumer = consumer;
            this.verifyExecutor = verifyExecutor;
            this.idp = idp;
        }

        @Override
        public Void call() throws Exception {
            Progress progress = console.startProgress(in.getCount());
            in.getProofs().forEach(proof -> {
                boolean taskAdded = false;
                do {
                    try {
                        verifyExecutor.execute(() -> consumer.accept(proof));
                        taskAdded = true;
                    } catch (RejectedExecutionException e) {
                        try {
                            Thread.sleep(20);
                        } catch (InterruptedException e1) {
                            log.warn("Unexpected interruption", e1);
                        }
                    }
                } while (!taskAdded);
                progress.increase(1);
            });
            verifyExecutor.shutdown();
            verifyExecutor.awaitTermination(1, TimeUnit.DAYS);
            idp.setEot();
            progress.finish();
            return null;
        }
    }
}
