package ee.ivxv.audit.shuffle;

import ee.ivxv.audit.shuffle.ByteTree.Leaf;
import ee.ivxv.audit.shuffle.ByteTree.Node;
import ee.ivxv.audit.shuffle.ShuffleConsole.ShuffleStep;
import ee.ivxv.common.math.ECGroupElement;
import ee.ivxv.common.math.Group;
import ee.ivxv.common.math.GroupElement;
import ee.ivxv.common.math.MathException;
import ee.ivxv.common.math.MathUtil;
import ee.ivxv.common.math.ModPGroup;
import ee.ivxv.common.math.ModPGroupElement;
import ee.ivxv.common.math.ProductGroup;
import ee.ivxv.common.math.ProductGroupElement;
import ee.ivxv.common.service.console.Progress;
import ee.ivxv.common.util.Util;
import java.io.IOException;
import java.math.BigInteger;
import java.security.MessageDigest;

/**
 * Verificatum proof of a shuffle verifier.
 * <p>
 * See the Verificatum manual for implementing independent verifier for the explanation of the
 * variables used in the verifier.
 */
public class Verifier {

    protected final ShuffleProof proof;
    protected final ShuffleConsole console;

    /**
     * Initialize the verifier using proof of a shuffle.
     * 
     * @param proof
     */
    public Verifier(ShuffleProof proof) {
        this.console = null;
        this.proof = proof;
    }

    public Verifier(ShuffleConsole console, ShuffleProof proof) {
        this.console = console;
        this.proof = proof;
    }

    public ShuffleProof get_proof() {
        return proof;
    }

    public byte[] compute_rho() {
        ProtocolInformation p = proof.get_ProtocolInformation();
        return compute_rho(p.get_sid(), p.get_auxsid(), p.get_version(), p.get_statdist(),
                p.get_vbitlenro(), p.get_ebitlenro(), p.get_prg(), p.get_pgroup(), p.get_rohash());
    }

    public byte[] compute_RO_seed(byte[] rho, GroupElement[] h) {
        return compute_RO_seed(rho, get_proof().get_ProtocolInformation().get_parsed_generator(), h,
                get_proof().get_PermutationCommitment().get_u(), get_proof().get_publickey(),
                get_proof().get_ciphertexts(), get_proof().get_shuffled_ciphertexts(),
                get_proof().get_ProtocolInformation().get_rohash(),
                get_proof().get_ProtocolInformation().get_prg());
    }

    public BigInteger[] compute_e(byte[] s) {
        return compute_e(s, get_proof().get_ProtocolInformation().get_ebitlenro(),
                get_proof().get_ciphertexts().length,
                get_proof().get_ProtocolInformation().get_prg());
    }

    public GroupElement[] compute_h(byte[] rho) {
        return compute_h(get_proof().get_ProtocolInformation().get_parsed_pgroup(), rho,
                get_proof().get_ProtocolInformation().get_statdist(),
                get_proof().get_ciphertexts().length,
                get_proof().get_ProtocolInformation().get_rohash(),
                get_proof().get_ProtocolInformation().get_prg());
    }

    public BigInteger compute_v(byte[] rho, byte[] s) {
        return compute_v(rho, s, get_proof().get_PoSCommitment().get_A_prim(),
                get_proof().get_PoSCommitment().get_B(),
                get_proof().get_PoSCommitment().get_B_prim(),
                get_proof().get_PoSCommitment().get_C_prim(),
                get_proof().get_PoSCommitment().get_D_prim(),
                get_proof().get_PoSCommitment().get_F_prim(),
                get_proof().get_ProtocolInformation().get_rohash(),
                get_proof().get_ProtocolInformation().get_vbitlenro());
    }

    public GroupElement compute_A(Progress progress, BigInteger[] e) throws MathException {
        return compute_A(progress, get_proof().get_PermutationCommitment().get_u(), e);
    }

    public GroupElement compute_C(Progress progress, GroupElement[] h)
            throws MathException, ShuffleException {
        return compute_C(progress, get_proof().get_PermutationCommitment().get_u(), h);
    }

    public GroupElement compute_D(Progress progress, GroupElement[] h, BigInteger[] e)
            throws ShuffleException, MathException {
        return compute_D(progress, get_proof().get_PoSCommitment().get_B(), h, e,
                get_proof().get_ciphertexts().length);
    }

    public GroupElement compute_F(Progress progress, BigInteger[] e) throws MathException {
        return compute_F(progress, get_proof().get_ciphertexts(), e);
    }

    public boolean verify_A(Progress progress, BigInteger v, GroupElement A, GroupElement[] h)
            throws MathException {
        return verify_A(progress, v, A, get_proof().get_PoSCommitment().get_A_prim(),
                get_proof().get_ProtocolInformation().get_parsed_generator(), h,
                get_proof().get_PoSReply().get_kA(), get_proof().get_PoSReply().get_kE());
    }

    public boolean verify_B(Progress progress, BigInteger v, GroupElement[] h)
            throws MathException {
        return verify_B(progress, v, get_proof().get_PoSCommitment().get_B(),
                get_proof().get_PoSCommitment().get_B_prim(),
                get_proof().get_ProtocolInformation().get_parsed_generator(),
                get_proof().get_PoSReply().get_kB(), get_proof().get_PoSReply().get_kE(), h);
    }

    public boolean verify_C(Progress progress, BigInteger v, GroupElement C) throws MathException {
        return verify_C(progress, v, C, get_proof().get_PoSCommitment().get_C_prim(),
                get_proof().get_ProtocolInformation().get_parsed_generator(),
                get_proof().get_PoSReply().get_kC());
    }

    public boolean verify_D(Progress progress, BigInteger v, GroupElement D) throws MathException {
        return verify_D(progress, v, D, get_proof().get_PoSCommitment().get_D_prim(),
                get_proof().get_ProtocolInformation().get_parsed_generator(),
                get_proof().get_PoSReply().get_kD());
    }

    public boolean verify_F(Progress progress, BigInteger v, GroupElement F) throws MathException {
        return verify_F(progress, v, F, get_proof().get_PoSCommitment().get_F_prim(),
                get_proof().get_publickey(), get_proof().get_PoSReply().get_kE(),
                get_proof().get_PoSReply().get_kF(), get_proof().get_shuffled_ciphertexts());
    }

    /**
     * Verify the correctness of the shuffle.
     * <p>
     * Throws an exception specifying the reason for failed verification.
     * 
     * @return Boolean True if the proof verifies. If not, then an exception is thrown.
     * @throws ShuffleException If the verification fails, denoting a reason.
     * @throws MathException If computation fails.
     */
    public boolean verify_all() throws ShuffleException, MathException {
        console.enter(ShuffleStep.VERIFY);
        console.enter(ShuffleStep.VERIFY_PARAMS);
        byte[] rho = compute_rho();
        GroupElement[] h = compute_h(rho);
        console.enter(ShuffleStep.VERIFY_NI);
        byte[] s = compute_RO_seed(rho, h);
        BigInteger[] e = compute_e(s);
        BigInteger v = compute_v(rho, s);

        int N = get_proof().get_ciphertexts().length;
        Progress progress = console.enter(ShuffleStep.VERIFY_PERM, 5 * N + 11);
        GroupElement A = compute_A(progress, e);
        GroupElement C = compute_C(progress, h);
        GroupElement D = compute_D(progress, h, e);
        if (!verify_A(progress, v, A, h)) {
            throw new ShuffleException("A failed");
        }
        if (!verify_B(progress, v, h)) {
            throw new ShuffleException("B failed");
        }
        if (!verify_C(progress, v, C)) {
            throw new ShuffleException("C failed");
        }
        if (!verify_D(progress, v, D)) {
            throw new ShuffleException("D failed");
        }
        progress.finish();
        progress = console.enter(ShuffleStep.VERIFY_RERAND,
                2 * N + get_proof().get_PoSReply().get_kF().getElements().length + 3);
        GroupElement F = compute_F(progress, e);
        if (!verify_F(progress, v, F)) {
            throw new ShuffleException("F failed");
        }
        progress.finish();
        return true;
    }

    /**
     * Verify the correctness of the shuffle.
     * <p>
     * If {@literal throwexception} is false, then now exceptions are thrown during computation and
     * verification.
     * 
     * @param throwexception Boolean in
     * @return Boolean indicating the correctness of the shuffle.
     */
    public boolean verify_all(boolean throwexception) {
        try {
            return verify_all();
        } catch (ShuffleException | MathException e) {
            if (!throwexception) {
                return false;
            }
        }
        return true;
    }

    public byte[] compute_rho(String sid, String auxsid, String version, int statdist,
            int vbitlenro, int ebitlenro, String prg, String pgroup, String rohash) {
        String fullsid = String.format("%s.%s", sid, auxsid);
        Node n = new Node(new ByteTree[] {new Leaf(version), new Leaf(fullsid),
                new Leaf(Util.toBytes(statdist)), new Leaf(Util.toBytes(vbitlenro)),
                new Leaf(Util.toBytes(ebitlenro)), new Leaf(prg), new Leaf(pgroup),
                new Leaf(rohash)});
        byte[] input = n.getEncoded();
        MessageDigest H = DataParser.getHash(rohash);
        byte[] rho = H.digest(input);
        return rho;
    }

    public byte[] compute_RO_seed(byte[] rho, GroupElement g, GroupElement[] h, GroupElement[] u,
            GroupElement pk, GroupElement[] w, GroupElement[] w_prim, String rohash, String prg) {
        if (!(pk instanceof ProductGroupElement)) {
            throw new IllegalArgumentException("pk must be ProductGroupElement");
        }
        ByteTree[] nodes = new ByteTree[] {new Leaf(g), new Node(h), new Node(u),
                new Node((ProductGroupElement) pk), new Node(DataParser.toArray(w).getElements()),
                new Node(DataParser.toArray(w_prim).getElements()),};
        Node n = new Node(nodes);
        RO ro = new RO(rohash);
        byte[] out = new byte[DataParser.getHash(prg).getDigestLength()];
        ro.setAmount(out.length * 8);
        try {
            ro.write(rho);
            n.writeEncoded(ro);
        } catch (IOException e) {
            // checked
        }
        ro.read(out, out.length * 8);
        return out;
    }

    public BigInteger[] compute_e(byte[] s, int n_e, int N, String prg) {
        PRNG gen = new PRNG(prg, s);
        BigInteger[] e = new BigInteger[N];
        BigInteger mask = BigInteger.ONE.shiftLeft(n_e);
        for (int i = 0; i < N; i++) {
            byte[] ti = new byte[(n_e + 7) / 8];
            gen.read(ti);
            BigInteger ei = new BigInteger(1, ti).mod(mask);
            e[i] = ei;
        }
        return e;
    }

    public GroupElement[] compute_h(Group G_q, byte[] rho, int n_r, int N, String rohash,
            String prg) {
        if (!(G_q instanceof ModPGroup)) {
            throw new IllegalArgumentException("Only ModPGroup supported");
        }
        BigInteger TWO = BigInteger.valueOf(2);
        BigInteger p = ((ModPGroup) G_q).getOrder();
        int n_p = p.bitLength();
        Leaf l = new Leaf("generators");
        byte[] seed = new byte[rho.length + l.getEncodedLength()];
        System.arraycopy(rho, 0, seed, 0, rho.length);
        System.arraycopy(l.getEncoded(), 0, seed, rho.length, l.getEncodedLength());
        @SuppressWarnings("resource")
        RO ro = new RO(rohash, seed);
        byte[] out = new byte[DataParser.getHash(prg).getDigestLength()];
        ro.read(out, out.length * 8);
        PRNG gen = new PRNG(prg, out);
        GroupElement[] h = new GroupElement[N];
        BigInteger mask = BigInteger.ONE.shiftLeft(n_p + n_r);
        for (int i = 0; i < N; i++) {
            byte[] ti = new byte[(n_p + n_r + 7) / 8];
            gen.read(ti);
            BigInteger tip = new BigInteger(1, ti).mod(mask);
            BigInteger hi = tip.modPow(TWO, p);
            h[i] = new ModPGroupElement((ModPGroup) G_q, hi);
        }
        return h;
    }

    public BigInteger compute_v(byte[] rho, byte[] s, GroupElement A_prim, GroupElement[] B,
            GroupElement[] B_prim, GroupElement C_prim, GroupElement D_prim,
            ProductGroupElement F_prim, String rohash, int n_v) {
        ByteTree[] nodes = new ByteTree[] {new Node(B), new Leaf(A_prim), new Node(B_prim),
                new Leaf(C_prim), new Leaf(D_prim), new Node(F_prim)};
        Node n = new Node(new ByteTree[] {new Leaf(s), new Node(nodes),});
        byte[] seed = new byte[rho.length + n.getEncodedLength()];
        System.arraycopy(rho, 0, seed, 0, rho.length);
        System.arraycopy(n.getEncoded(), 0, seed, rho.length, n.getEncodedLength());
        @SuppressWarnings("resource")
        RO ro = new RO(rohash, seed);
        byte[] out = new byte[(n_v + 7) / 8];
        ro.read(out, n_v);
        return new BigInteger(1, out);
    }

    public GroupElement compute_A(Progress progress, GroupElement[] u, BigInteger[] e)
            throws MathException {
        GroupElement res = u[0].getGroup().getIdentity();
        for (int i = 0; i < u.length; i++) {
            GroupElement exped = u[i].scale(e[i]);
            res = res.op(exped);
            progress.increase(1);
        }
        return res;
    }

    public GroupElement compute_C(Progress progress, GroupElement[] u, GroupElement[] h)
            throws MathException, ShuffleException {
        if (u.length != h.length) {
            throw new ShuffleException("u and h length does not match");
        }
        GroupElement up = u[0].getGroup().getIdentity();
        GroupElement hp = h[0].getGroup().getIdentity();
        for (int i = 0; i < u.length; i++) {
            up = u[i].op(up);
            hp = h[i].op(hp);
            progress.increase(1);
        }
        GroupElement hpi = hp.inverse();
        GroupElement res = up.op(hpi);
        progress.increase(2);
        return res;
    }

    public GroupElement compute_D(Progress progress, GroupElement[] B, GroupElement[] h,
            BigInteger[] e, int N) throws ShuffleException, MathException {
        BigInteger ep = BigInteger.ONE;
        BigInteger q = null;
        if (h[0] instanceof ModPGroupElement) {
            ModPGroup group = (ModPGroup) h[0].getGroup();
            q = MathUtil.safePrimeOrder(group.getOrder());
        } else if (h[0] instanceof ECGroupElement) {
            q = h[0].getGroup().getOrder();
        } else {
            throw new ShuffleException("Can not find order");
        }
        for (int i = 0; i < e.length; i++) {
            ep = ep.multiply(e[i]);
            ep = ep.mod(q);
            progress.increase(1);
        }
        GroupElement hp = h[0].scale(ep);
        hp = hp.inverse();
        GroupElement ret = B[N - 1].op(hp);
        progress.increase(3);
        return ret;
    }

    public GroupElement compute_F(Progress progress, GroupElement[] w, BigInteger[] e)
            throws MathException {
        GroupElement res = w[0].getGroup().getIdentity();
        for (int i = 0; i < w.length; i++) {
            GroupElement exped = w[i].scale(e[i]);
            res = res.op(exped);
            progress.increase(1);
        }
        return res;
    }

    public boolean verify_A(Progress progress, BigInteger v, GroupElement A, GroupElement A_prim,
            GroupElement g, GroupElement[] h, BigInteger k_A, BigInteger[] k_E)
            throws MathException {
        GroupElement left = A.scale(v).op(A_prim);
        progress.increase(1);
        GroupElement right = h[0].getGroup().getIdentity();
        for (int i = 0; i < h.length; i++) {
            right = right.op(h[i].scale(k_E[i]));
            progress.increase(1);
        }
        right = right.op(g.scale(k_A));
        progress.increase(1);
        return left.equals(right);
    }

    public boolean verify_B(Progress progress, BigInteger v, GroupElement[] B,
            GroupElement[] B_prim, GroupElement g, BigInteger[] k_B, BigInteger[] k_E,
            GroupElement[] h) throws MathException {
        // the number of computations is different in threaded and non-threaded case. In threaded
        // case the thread which sees invalid proof stops and this is propagated to the controlling
        // thread. In non-threaded case, all values are computed and then checked one-by-one.
        boolean[] res = new boolean[B.length];
        GroupElement left = B[0].scale(v).op(B_prim[0]);
        GroupElement right = h[0].scale(k_E[0]).op(g.scale(k_B[0]));
        res[0] = left.equals(right);
        progress.increase(1);
        for (int i = 1; i < B.length; i++) {
            left = B[i].scale(v).op(B_prim[i]);
            right = B[i - 1].scale(k_E[i]).op(g.scale(k_B[i]));
            res[i] = left.equals(right);
            progress.increase(1);
        }
        for (int i = 0; i < res.length; i++) {
            if (!res[i]) {
                return false;
            }
        }
        return true;
    }

    public boolean verify_C(Progress progress, BigInteger v, GroupElement C, GroupElement C_prim,
            GroupElement g, BigInteger k_C) throws MathException {
        GroupElement left = C.scale(v).op(C_prim);
        progress.increase(1);
        GroupElement right = g.scale(k_C);
        progress.increase(1);
        return left.equals(right);
    }

    public boolean verify_D(Progress progress, BigInteger v, GroupElement D, GroupElement D_prim,
            GroupElement g, BigInteger k_D) throws MathException {
        GroupElement left = D.scale(v).op(D_prim);
        progress.increase(1);
        GroupElement right = g.scale(k_D);
        progress.increase(1);
        return left.equals(right);
    }

    public boolean verify_F(Progress progress, BigInteger v, GroupElement F, GroupElement F_prim,
            GroupElement pk, BigInteger[] k_E, ProductGroupElement k_F, GroupElement[] w_prim)
            throws MathException {
        // the number of computations differ in threaded and non-threaded case. In threaded case we
        // also aggregate the per-thread results.
        GroupElement left = F.scale(v).op(F_prim);
        GroupElement right = w_prim[0].getGroup().getIdentity();
        for (int i = 0; i < w_prim.length; i++) {
            right = right.op(w_prim[i].scale(k_E[i]));
            progress.increase(1);
        }
        BigInteger[] factors = new BigInteger[k_F.getElements().length];
        for (int i = 0; i < factors.length; i++) {
            factors[i] = ((ModPGroupElement) k_F.getElements()[i]).getValue().negate();
            progress.increase(1);
        }
        ProductGroupElement pkl = (ProductGroupElement) ((ProductGroupElement) pk).getElements()[0];
        ProductGroupElement pkr = (ProductGroupElement) ((ProductGroupElement) pk).getElements()[1];
        ProductGroupElement tmpl = pkl.scale(factors);
        progress.increase(1);
        ProductGroupElement tmpr = pkr.scale(factors);
        progress.increase(1);
        ProductGroupElement tmp = new ProductGroupElement((ProductGroup) pk.getGroup(), tmpl, tmpr);
        right = right.op(tmp);
        progress.increase(1);
        return left.equals(right);
    }
}
