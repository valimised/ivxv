package ee.ivxv.key.tool;

import ee.ivxv.common.M;
import ee.ivxv.common.cli.Arg;
import ee.ivxv.common.cli.Args;
import ee.ivxv.common.cli.Tool;
import ee.ivxv.common.crypto.CorrectnessUtil.CiphertextCorrectness;
import ee.ivxv.common.crypto.rnd.NativeRnd;
import ee.ivxv.common.model.AnonymousBallotBox;
import ee.ivxv.common.model.CandidateList;
import ee.ivxv.common.model.District;
import ee.ivxv.common.model.DistrictList;
import ee.ivxv.common.model.IBallotBox;
import ee.ivxv.common.service.bbox.impl.BboxHelperImpl;
import ee.ivxv.common.service.i18n.MessageException;
import ee.ivxv.common.service.report.Reporter;
import ee.ivxv.common.service.smartcard.Card;
import ee.ivxv.common.service.smartcard.Cards;
import ee.ivxv.common.service.smartcard.IndexedBlob;
import ee.ivxv.common.util.I18nConsole;
import ee.ivxv.common.util.ToolHelper;
import ee.ivxv.key.KeyContext;
import ee.ivxv.key.Msg;
import ee.ivxv.key.model.Vote;
import ee.ivxv.key.protocol.DecryptionProtocol;
import ee.ivxv.key.protocol.ProtocolException;
import ee.ivxv.key.protocol.SigningProtocol;
import ee.ivxv.key.protocol.ThresholdParameters;
import ee.ivxv.key.protocol.decryption.recover.RecoverDecryption;
import ee.ivxv.key.protocol.signing.shoup.ShoupSigning;
import ee.ivxv.key.tool.DecryptTool.DecryptArgs;
import ee.ivxv.key.util.ElectionResult;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.HashSet;
import java.util.Set;
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

/**
 * DecryptTool is a tool for decrypting the encrypted ballots.
 */
public class DecryptTool implements Tool.Runner<DecryptArgs> {

    static final Logger log = LoggerFactory.getLogger(DecryptTool.class);

    private final KeyContext ctx;
    private final I18nConsole console;
    private final ToolHelper tool;

    public DecryptTool(KeyContext ctx) {
        this.ctx = ctx;
        this.console = new I18nConsole(ctx.i.console, ctx.i.i18n);
        tool = new ToolHelper(console, ctx.container, new BboxHelperImpl(ctx.conf, ctx.container));
    }

    @Override
    public boolean run(DecryptArgs args) throws Exception {
        tool.checkBbChecksum(args.abb.value(), args.abbChecksum.value());
        AnonymousBallotBox abb = tool.readJsonAbb(args.abb.value(), IBallotBox.Type.ANONYMIZED);
        if (args.questionCount.value() != abb.getNumberOfQuestions()) {
            throw new MessageException(Msg.e_abb_invalid_question_count, args.questionCount.value(),
                    abb.getNumberOfQuestions());
        }
        DistrictList districts = tool.readJsonDistricts(args.districts.value());
        CandidateList candidates = tool.readJsonCandidates(args.candidates.value(), districts);

        console.println();
        console.println(Msg.m_abb_dist_verifying);
        verifyAbb(abb, districts);
        console.println(Msg.m_abb_dist_ok);

        console.println();
        if (args.doProvable.value()) {
            console.println(Msg.m_with_proof);
        } else {
            console.println(Msg.m_without_proof);
        }
        console.println(Msg.m_protocol_init);
        DecryptionProtocol dec = null;
        SigningProtocol signer = null;
        if (args.recover.isSet()) {
            ThresholdParameters tparams = new ThresholdParameters(args.dn.value(), args.dm.value());
            byte[] aid = new byte[] {0x01};
            byte[] decShareName = new byte[] {0x44, 0x45, 0x43};
            byte[] signShareName = new byte[] {0x53, 0x49, 0x47, 0x4E};
            Cards cards = ctx.card.createCards();
            if (!ctx.card.isPluggableService()) {
                for (int i = 0; i < tparams.getParties(); i++) {
                    cards.addCard(String.valueOf(i));
                }
            }
            Set<IndexedBlob> decBlobs = new HashSet<>();
            Set<IndexedBlob> signBlobs = new HashSet<>();
            for (int i = 0; i < tparams.getThreshold(); i++) {
                Card card;
                if (ctx.card.isPluggableService()) {
                    card = ctx.card.createCard("-1");
                    cards.initUnprocessedCard(card);
                } else {
                    card = cards.getCard(i);
                }
                IndexedBlob ib = card.getIndexedBlob(aid, decShareName);
                if (ib.getIndex() < 1 || ib.getIndex() > tparams.getParties()) {
                    throw new ProtocolException("Indexed blob index mismatch");
                }
                decBlobs.add(ib);

                ib = card.getIndexedBlob(aid, signShareName);
                if (ib.getIndex() < 1 || ib.getIndex() > tparams.getParties()) {
                    throw new ProtocolException("Indexed blob index mismatch");
                }
                signBlobs.add(ib);
            }

            dec = new RecoverDecryption(decBlobs, tparams, args.doProvable.value());
            signer = new ShoupSigning(signBlobs, tparams, new NativeRnd());
        }
        console.println(Msg.m_protocol_init_ok);

        console.println();
        console.println(Msg.m_dec_start);
        Path out = args.outputPath.value();
        ElectionResult result = processVotes(abb, dec, candidates, districts,
                args.doProvable.value(), args.checkDecodable.value(), ctx.args.threads.value());
        console.println(Msg.m_dec_done);

        console.println();
        console.println(M.m_out_start, out);
        Files.createDirectory(out);

        console.println(Msg.m_out_tally);
        result.outputTally(out, signer);

        if (args.doProvable.value()) {
            console.println(Msg.m_out_proof);
            result.outputProof(out);
        }

        console.println(Msg.m_out_invalid);
        result.outputInvalid(out);

        console.println(M.m_out_done);

        return true;
    }

    private ElectionResult processVotes(AnonymousBallotBox abb, DecryptionProtocol dec,
            CandidateList candidates, DistrictList districts, boolean withProof,
            boolean checkDecodable, int threadCount) throws Exception {
        ElectionResult result =
                new ElectionResult(abb.getElection(), candidates, districts, withProof);
        // WorkerFactory consumer = new WorkerFactory(getDecConsumer(dec, result));

        ExecutorService ioExecutor = Executors.newFixedThreadPool(3);
        CompletionService<Void> ioCompService = new ExecutorCompletionService<>(ioExecutor);

        ExecutorService decExecutor;
        if (threadCount == 0) {
            decExecutor = Executors.newCachedThreadPool();
        } else {
            decExecutor = new ThreadPoolExecutor(threadCount, threadCount, 0L,
                    TimeUnit.MILLISECONDS, new ArrayBlockingQueue<>(threadCount * 2));
        }

        WorkManager manager = new WorkManager(abb, getDecConsumer(dec, result, checkDecodable),
                decExecutor, result);
        ioCompService.submit(manager);
        ioCompService
                .submit(result.getResultWorker(abb.getNumberOfBallots(), console, ctx.reporter));

        try {
            for (int done = 0; done < 2; done++) {
                ioCompService.take().get();
            }
        } finally {
            ioExecutor.shutdown();
            decExecutor.shutdown();
        }

        return result;
    }

    private Consumer<Vote> getDecConsumer(DecryptionProtocol dec, ElectionResult result,
            boolean checkDecodable) {
        return (vote) -> {
            byte[] msg = vote.getVote();
            boolean isCorrect = true;
            if (checkDecodable) {
                try {
                    if (dec.checkCorrectness(msg) != CiphertextCorrectness.VALID) {
                        log.warn("Non-decodable ciphertext");
                        isCorrect = false;
                    }
                } catch (ProtocolException e) {
                    log.warn("Could not check correctness", e);
                    isCorrect = false;
                }
            }
            try {
                if (isCorrect) {
                    vote.setProof(dec.decryptMessage(msg));
                }
            } catch (Exception e) {
                log.warn("Couldn't decrypt vote", e);
                // console.println(M.e_decryption_error, vote);
            }
            result.addVote(vote);
        };
    }

    private void verifyAbb(AnonymousBallotBox abb, DistrictList districts) {
        abb.getDistricts().forEach((d, sMap) -> {
            District dist = districts.getDistricts().get(d);
            if (dist == null) {
                throw new MessageException(Msg.e_illegal_vote_district, d);
            }
            sMap.keySet().forEach(s -> {
                if (!dist.getStations().contains(s)) {
                    throw new MessageException(Msg.e_illegal_vote_station, s);
                }
            });
        });
    }

    public static class DecryptArgs extends Args {
        Arg<String> identifier = Arg.aString(Msg.arg_identifier);
        Arg<Path> abb = Arg.aPath(Msg.d_anonballotbox, true, false);
        Arg<Path> abbChecksum = Arg.aPath(Msg.d_anonballotbox_checksum, true, false);
        Arg<Integer> questionCount = Arg.anInt(Msg.d_questioncount).setDefault(1);
        Arg<Path> candidates = Arg.aPath(Msg.d_candidates, true, false);
        Arg<Path> districts = Arg.aPath(Msg.d_districts, true, false);
        Arg<Path> outputPath = Arg.aPath(Msg.arg_out, false, null);
        Arg<Boolean> doProvable = Arg.aFlag(Msg.d_provable).setDefault(true);
        Arg<Boolean> checkDecodable = Arg.aFlag(Msg.d_check_decodable).setDefault(false);

        // protocols

        Arg<Integer> dm = Arg.anInt(Msg.arg_threshold);
        Arg<Integer> dn = Arg.anInt(Msg.arg_parties);
        Arg<Args> recover = new Arg.Tree(Msg.d_recover, dm, dn).setOptional();

        Arg.Tree protocol = new Arg.Tree(Msg.d_protocol, recover).setExclusive();

        public DecryptArgs() {
            super();
            args.add(identifier);
            args.add(abb);
            args.add(abbChecksum);
            args.add(questionCount);
            args.add(candidates);
            args.add(districts);
            args.add(outputPath);
            args.add(doProvable);
            args.add(checkDecodable);
            args.add(protocol);
        }
    }

    private class WorkManager implements Callable<Void> {

        private final AnonymousBallotBox abb;
        private final Consumer<Vote> consumer;
        private final ExecutorService decExecutor;
        private final ElectionResult result;

        WorkManager(AnonymousBallotBox abb, Consumer<Vote> factory, ExecutorService decExecutor,
                ElectionResult result) {
            this.abb = abb;
            this.consumer = factory;
            this.decExecutor = decExecutor;
            this.result = result;
        }

        @Override
        public Void call() throws Exception {
            abb.getDistricts().forEach((d, sMap) -> sMap.forEach((s, qMap) -> {
                qMap.forEach((q, cList) -> cList.forEach(c -> {
                    boolean taskAdded = false;
                    do {
                        try {
                            decExecutor.execute(() -> consumer.accept(new Vote(d, s, q, c)));
                            taskAdded = true;
                        } catch (RejectedExecutionException e) {
                            try {
                                Thread.sleep(20);
                            } catch (InterruptedException e1) {
                                log.warn("Unexpected interruption", e1);
                            }
                        }
                    } while (!taskAdded);
                }));
            }));
            decExecutor.shutdown();
            decExecutor.awaitTermination(1, TimeUnit.DAYS);
            result.setEot();
            return null;
        }
    }
}
