package ee.ivxv.audit.shuffle;

import ee.ivxv.common.math.GroupElement;
import ee.ivxv.common.math.MathException;
import ee.ivxv.common.math.ModPGroupElement;
import ee.ivxv.common.math.ProductGroup;
import ee.ivxv.common.math.ProductGroupElement;
import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.Callable;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import java.util.concurrent.FutureTask;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class ThreadedVerifier extends Verifier {
    static final Logger log = LoggerFactory.getLogger(ThreadedVerifier.class);

    private int nothreads;
    private ExecutorService executor;

    public ThreadedVerifier(ShuffleProof proof, int nothreads) {
        super(proof);
        this.nothreads = nothreads;
        this.executor = Executors.newFixedThreadPool(nothreads);
    }

    public GroupElement compute_A_threaded(BigInteger[] e)
            throws MathException, InterruptedException, ExecutionException {
        return compute_A(get_proof().get_PermutationCommitment().get_u(), e, nothreads, executor);
    }

    public GroupElement compute_F_threaded(BigInteger[] e)
            throws MathException, InterruptedException, ExecutionException {
        return compute_F(get_proof().get_ciphertexts(), e, nothreads, executor);
    }

    public boolean verify_A_threaded(BigInteger v, GroupElement A, GroupElement[] h)
            throws MathException, InterruptedException, ExecutionException {
        return verify_A(v, A, get_proof().get_PoSCommitment().get_A_prim(),
                get_proof().get_ProtocolInformation().get_parsed_generator(), h,
                get_proof().get_PoSReply().get_kA(), get_proof().get_PoSReply().get_kE(), nothreads,
                executor);
    }

    public boolean verify_B_threaded(BigInteger v, GroupElement[] h)
            throws MathException, InterruptedException, ExecutionException {
        return verify_B(v, get_proof().get_PoSCommitment().get_B(),
                get_proof().get_PoSCommitment().get_B_prim(),
                get_proof().get_ProtocolInformation().get_parsed_generator(),
                get_proof().get_PoSReply().get_kB(), get_proof().get_PoSReply().get_kE(), h,
                nothreads, executor);
    }

    public boolean verify_F_threaded(BigInteger v, GroupElement F) throws MathException, InterruptedException, ExecutionException {
        return verify_F(v, F, get_proof().get_PoSCommitment().get_F_prim(),
                get_proof().get_publickey(), get_proof().get_PoSReply().get_kE(),
                get_proof().get_PoSReply().get_kF(), get_proof().get_shuffled_ciphertexts(),
                nothreads, executor);
    }

    public boolean verify_all() throws ShuffleException, MathException {
        byte[] rho = compute_rho();
        GroupElement[] h = compute_h(rho);
        byte[] s = compute_RO_seed(rho, h);
        BigInteger[] e = compute_e(s);
        BigInteger v = compute_v(rho, s);
        GroupElement C = compute_C(h);
        GroupElement D = compute_D(h, e);
        if (!verify_C(v, C)) {
            throw new ShuffleException("C failed");
        }
        if (!verify_D(v, D)) {
            throw new ShuffleException("D failed");
        }

        try {
            GroupElement A = compute_A_threaded(e);
            GroupElement F = compute_F_threaded(e);
            if (!verify_A_threaded(v, A, h)) {
                throw new ShuffleException("A failed");
            }
            if (!verify_B_threaded(v, h)) {
                throw new ShuffleException("B failed");
            }
            if (!verify_F_threaded(v, F)) {
                throw new ShuffleException("F failed");
            }
        } catch (InterruptedException | ExecutionException e1) {
            throw new ShuffleException(e1);
        } finally {
            executor.shutdown();
        }
        return true;
    }

    private static GroupElement compute_A(GroupElement[] u, BigInteger[] e, int nothreads,
            ExecutorService executor)
            throws MathException, InterruptedException, ExecutionException {
        log.debug("Compute A");
        List<Future<GroupElement>> futures = new ArrayList<>();
        for (int i = 0; i < nothreads; i++) {
            FutureTask<GroupElement> ft =
                    new FutureTask<>(get_compute_A_worker(u, e, i, nothreads));
            executor.submit(ft);
            futures.add(ft);
        }
        log.debug("Started all compute A workers");
        GroupElement res = u[0].getGroup().getIdentity();
        log.debug("Collecting compute A worker results");
        for (Future<GroupElement> ft : futures) {
            res = res.op(ft.get());
            log.debug("Collected compute A worker result");
        }
        log.debug("Collected all compute A worker results");
        return res;
    }

    private static Callable<GroupElement> get_compute_A_worker(GroupElement[] u, BigInteger[] e,
            int threadid, int nothreads) {
        return () -> {
            log.debug("Compute A worker started");
            GroupElement res = u[0].getGroup().getIdentity();
            for (int i = 0; i < u.length; i++) {
                if (i % nothreads != threadid)
                    continue;
                GroupElement exped = u[i].scale(e[i]);
                res = res.op(exped);
            }
            log.debug("Compute A worker finished");
            return res;
        };
    }

    public static GroupElement compute_F(GroupElement[] w, BigInteger[] e, int nothreads,
            ExecutorService executor)
            throws MathException, InterruptedException, ExecutionException {
        log.debug("Compute F");
        List<Future<GroupElement>> futures = new ArrayList<>();
        for (int i = 0; i < nothreads; i++) {
            FutureTask<GroupElement> ft =
                    new FutureTask<>(get_compute_F_worker(w, e, i, nothreads));
            executor.submit(ft);
            futures.add(ft);
        }
        log.debug("Started all compute F workers");
        GroupElement res = w[0].getGroup().getIdentity();
        log.debug("Collecting compute F worker results");
        for (Future<GroupElement> ft : futures) {
            res = res.op(ft.get());
            log.debug("Collected compute F worker result");
        }
        log.debug("Collected all compute F worker results");
        return res;
    }

    private static Callable<GroupElement> get_compute_F_worker(GroupElement[] w, BigInteger[] e,
            int threadid, int nothreads) throws MathException {
        return () -> {
            log.debug("Compute F worker started");
            GroupElement res = w[0].getGroup().getIdentity();
            for (int i = 0; i < w.length; i++) {
                if (i % nothreads != threadid)
                    continue;
                GroupElement exped = w[i].scale(e[i]);
                res = res.op(exped);
            }
            log.debug("Compute F worker finished");
            return res;
        };
    }

    private static boolean verify_A(BigInteger v, GroupElement A, GroupElement A_prim,
            GroupElement g, GroupElement[] h, BigInteger k_A, BigInteger[] k_E, int nothreads,
            ExecutorService executor)
            throws MathException, InterruptedException, ExecutionException {
        log.debug("Verify A");
        List<Future<GroupElement>> futures = new ArrayList<>();
        for (int i = 0; i < nothreads; i++) {
            FutureTask<GroupElement> ft =
                    new FutureTask<>(get_verify_A_worker(h, k_E, i, nothreads));
            executor.submit(ft);
            futures.add(ft);
        }
        log.debug("Started verify A workers");
        GroupElement left = A.scale(v).op(A_prim);
        GroupElement right = h[0].getGroup().getIdentity();
        log.debug("Collecting verify A worker results");
        for (Future<GroupElement> ft : futures) {
            right = right.op(ft.get());
            log.debug("Collected verify A worker result");
        }
        log.debug("Collected all verify A worker results");
        right = right.op(g.scale(k_A));
        return left.equals(right);
    }

    private static Callable<GroupElement> get_verify_A_worker(GroupElement[] h, BigInteger[] k_E,
            int threadid, int nothreads) throws MathException {
        return () -> {
            log.debug("Verify A worker started");
            GroupElement right = h[0].getGroup().getIdentity();
            for (int i = 0; i < h.length; i++) {
                if (i % nothreads != threadid)
                    continue;
                right = right.op(h[i].scale(k_E[i]));
            }
            log.debug("Verify A worker finished");
            return right;
        };
    }

    private static boolean verify_B(BigInteger v, GroupElement[] B, GroupElement[] B_prim,
            GroupElement g, BigInteger[] k_B, BigInteger[] k_E, GroupElement[] h, int nothreads,
            ExecutorService executor)
            throws MathException, InterruptedException, ExecutionException {
        log.debug("Verify B");
        GroupElement left = B[0].scale(v).op(B_prim[0]);
        GroupElement right = h[0].scale(k_E[0]).op(g.scale(k_B[0]));
        if (!left.equals(right)) {
            return false;
        }
        log.debug("Verified first B value");
        List<Future<Boolean>> futures = new ArrayList<>();
        for (int i = 0; i < nothreads; i++) {
            FutureTask<Boolean> ft = new FutureTask<>(
                    get_verify_B_worker(v, B, B_prim, g, k_B, k_E, h, i, nothreads));
            executor.submit(ft);
            futures.add(ft);
        }
        log.debug("Started verify B workers");
        log.debug("Collecting verify B worker results");
        for (Future<Boolean> ft : futures) {
            if (!ft.get())
                return false;
            log.debug("Collected verify B worker result");
        }
        log.debug("Collected all verify B worker results");
        return true;
    }

    private static Callable<Boolean> get_verify_B_worker(BigInteger v, GroupElement[] B,
            GroupElement[] B_prim, GroupElement g, BigInteger[] k_B, BigInteger[] k_E,
            GroupElement[] h, int threadid, int nothreads) {
        return () -> {
            log.debug("Verify B worker started");
            GroupElement left, right;
            for (int i = 1; i < B.length; i++) {
                if (i % nothreads != threadid)
                    continue;
                left = B[i].scale(v).op(B_prim[i]);
                right = B[i - 1].scale(k_E[i]);
                right = right.op(g.scale(k_B[i]));
                if (!left.equals(right))
                    return false;
            }
            log.debug("Verify B worker finished");
            return true;
        };
    }

    private static boolean verify_F(BigInteger v, GroupElement F, GroupElement F_prim,
            GroupElement pk, BigInteger[] k_E, ProductGroupElement k_F, GroupElement[] w_prim,
            int nothreads, ExecutorService executor) throws MathException, InterruptedException, ExecutionException {
        log.debug("Verify F");
        List<Future<GroupElement>> futures = new ArrayList<>();
        for (int i = 0; i < nothreads; i++) {
            FutureTask<GroupElement> ft = new FutureTask<>(
                    get_verify_F_worker(k_E, w_prim, i, nothreads));
            executor.submit(ft);
            futures.add(ft);
        }
        log.debug("Started verify F workers");
        BigInteger[] factors = new BigInteger[k_F.getElements().length];
        for (int i = 0; i < factors.length; i++) {
            factors[i] = ((ModPGroupElement) k_F.getElements()[i]).getValue().negate();
        }
        GroupElement left = F.scale(v).op(F_prim);
        GroupElement right = w_prim[0].getGroup().getIdentity();
        log.debug("Collecting verify F worker results");
        for (Future<GroupElement> ft : futures) {
            right = right.op(ft.get());
            log.debug("Collected verify F worker result");
        }
        log.debug("Collected all verify F worker results");
        ProductGroupElement pkl = (ProductGroupElement) ((ProductGroupElement) pk).getElements()[0];
        ProductGroupElement pkr = (ProductGroupElement) ((ProductGroupElement) pk).getElements()[1];
        ProductGroupElement tmpl = pkl.scale(factors);
        ProductGroupElement tmpr = pkr.scale(factors);
        ProductGroupElement tmp = new ProductGroupElement((ProductGroup) pk.getGroup(), tmpl, tmpr);
        right = right.op(tmp);
        return left.equals(right);
    }

    private static Callable<GroupElement> get_verify_F_worker(BigInteger[] k_E,
            GroupElement[] w_prim, int threadid, int nothreads) {
        return () -> {
            GroupElement right = w_prim[0].getGroup().getIdentity();
            for (int i = 0; i < w_prim.length; i++) {
                if (i % nothreads != threadid)
                    continue;
                right = right.op(w_prim[i].scale(k_E[i]));
            }
            return right;
        };
    }
}
