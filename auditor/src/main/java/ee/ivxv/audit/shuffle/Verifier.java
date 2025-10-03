package ee.ivxv.audit.shuffle;

import ee.ivxv.audit.shuffle.ByteTree.Leaf;
import ee.ivxv.audit.shuffle.ByteTree.Node;
import ee.ivxv.audit.shuffle.ShuffleConsole.ShuffleStep;
import ee.ivxv.common.math.*;
import ee.ivxv.common.service.console.Progress;
import ee.ivxv.common.util.Util;

import java.io.IOException;
import java.math.BigInteger;
import java.util.function.Supplier;

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

    /**
     * See:
     * <a href="https://github.com/verificatum/verificatum-vmn/blob/9d9c45e37fa043881e58d71901ce45e387b31271/src/java/com/verificatum/protocol/mixnet/MixNetElGamalVerifyFiatShamirSession.java#L186">globalPrefix</a>
     *
     * @return
     */
    public byte[] compute_rho() {
        ProtocolInformation p = proof.get_ProtocolInformation();

        String fullsid = String.format("%s.%s", p.get_sid(), p.get_auxsid());

        Node n = new Node(new ByteTree[]{
                new Leaf(p.get_version()),
                new Leaf(fullsid),
                new Leaf(Util.toBytes(p.get_statdist())),
                new Leaf(Util.toBytes(p.get_vbitlenro())),
                new Leaf(Util.toBytes(p.get_ebitlenro())),
                new Leaf(p.get_prg()),
                new Leaf(p.get_pgroup()),
                new Leaf(p.get_rohash())
        });

        return DataParser.getHash(p.get_rohash()).digest(n.getEncoded());
    }

    /**
     * See:
     * <a href="https://github.com/verificatum/verificatum-vmn/blob/9d9c45e37fa043881e58d71901ce45e387b31271/src/java/com/verificatum/protocol/hvzk/PoSTW.java#L126">prgSeed</a>
     *
     * @param rho
     * @param h
     * @return
     */
    public byte[] compute_RO_seed(byte[] rho, GroupElement[] h) {
        ProductGroupElement publickey = (ProductGroupElement) get_proof().get_publickey();

        ProtocolInformation p = proof.get_ProtocolInformation();
        GroupElement generator = get_proof().get_ProtocolInformation().get_parsed_generator();
        String pgroup = p.get_pgroup();

        ByteTree[] nodes;
        ByteTree generatorNode;

        if (pgroup.contains("ECqPGroup")) {
            generatorNode = new Node((ECGroupElement) generator);
        } else {
            generatorNode = new Leaf(generator);
        }

        nodes = new ByteTree[]{
                generatorNode,
                new Node(h),
                new Node(get_proof().get_PermutationCommitment().get_u()),
                new Node(publickey),
                new Node(DataParser.toArray(get_proof().get_ciphertexts()).getElements()),
                new Node(DataParser.toArray(get_proof().get_shuffled_ciphertexts()).getElements())
        };

        Node node = new Node(nodes);
        RO ro = new RO(get_proof().get_ProtocolInformation().get_rohash());
        byte[] out = new byte[DataParser.getHash(get_proof().get_ProtocolInformation().get_prg()).getDigestLength()];
        ro.setAmount(out.length * 8);

        try {
            ro.write(rho);
            node.writeEncoded(ro);
        } catch (IOException e) {
            // checked
        }

        ro.read(out, out.length * 8);
        return out;
    }

    /**
     * See:
     * <a href="https://github.com/verificatum/verificatum-vmn/blob/9d9c45e37fa043881e58d71901ce45e387b31271/src/java/com/verificatum/protocol/hvzk/PoSCBasicTW.java#L354">e</a>
     *
     * @param s
     * @return
     */
    public BigInteger[] compute_e(byte[] s) {
        PRNG gen = new PRNG(get_proof().get_ProtocolInformation().get_prg(), s);
        int n_e = get_proof().get_ProtocolInformation().get_ebitlenro();
        int N = get_proof().get_ciphertexts().length;
        BigInteger[] scalars = new BigInteger[N];
        BigInteger mask = BigInteger.ONE.shiftLeft(n_e);

        for (int i = 0; i < N; i++) {
            byte[] ti = new byte[(n_e + 7) / 8];
            gen.read(ti);
            scalars[i] = new BigInteger(1, ti).mod(mask);
        }
        return scalars;
    }

    /**
     * For ECGroup see:
     * <a href="https://github.com/verificatum/verificatum-vcr/blob/97974cfc4ebbb323e49396222823e226cae2bebe/src/java/com/verificatum/arithm/ECqPGroup.magic#L391">ec</a>
     * For ModpGroup see:
     * <a href="https://github.com/verificatum/verificatum-vcr/blob/97974cfc4ebbb323e49396222823e226cae2bebe/src/java/com/verificatum/arithm/ModPGroup.java#L778">modp</a>
     *
     * @param rho
     * @return
     */
    public GroupElement[] compute_h(byte[] rho) {
        Group G_q = get_proof().get_ProtocolInformation().get_parsed_pgroup();
        BigInteger p = get_proof().get_ProtocolInformation().get_parsed_pgroup().getOrder();
        int n_r = get_proof().get_ProtocolInformation().get_statdist();
        int n_p = p.bitLength();
        int N = get_proof().get_ciphertexts().length;
        String prg = get_proof().get_ProtocolInformation().get_prg();
        String rohash = get_proof().get_ProtocolInformation().get_rohash();

        Leaf l = new Leaf("generators");

        byte[] seed = new byte[rho.length + l.getEncodedLength()];

        System.arraycopy(rho, 0, seed, 0, rho.length);
        System.arraycopy(l.getEncoded(), 0, seed, rho.length, l.getEncodedLength());

        @SuppressWarnings("resource")
        RO ro = new RO(rohash, seed);

        byte[] out = new byte[DataParser.getHash(prg).getDigestLength()];
        ro.read(out, out.length * 8);
        PRNG gen = new PRNG(prg, out);

        GroupElement[] elements = new GroupElement[N];
        BigInteger mask = BigInteger.ONE.shiftLeft(n_p + n_r);

        if (G_q instanceof ECGroup) {
            // Each dprg.get() invocation produces new byte array
            Supplier<byte[]> dprg = () -> {
                byte[] rand = new byte[(n_p + n_r + 7) / 8];
                gen.read(rand);
                return new BigInteger(1, rand).mod(mask).toByteArray();
            };

            elements = new ECGroup().pseudoRandomElements(N, dprg);
        } else {
            BigInteger TWO = BigInteger.valueOf(2);
            for (int i = 0; i < N; i++) {
                byte[] rand = new byte[(n_p + n_r + 7) / 8];
                gen.read(rand);
                BigInteger value = new BigInteger(1, rand).mod(mask).modPow(TWO, p);
                elements[i] = new ModPGroupElement((ModPGroup) G_q, value);
            }
        }

        return elements;
    }

    /**
     * See:
     * <a href="https://github.com/verificatum/verificatum-vmn/blob/9d9c45e37fa043881e58d71901ce45e387b31271/src/java/com/verificatum/protocol/hvzk/PoSCBasicTW.java#L597">v</a>
     *
     * @param rho
     * @param s
     * @return
     */
    public BigInteger compute_v(byte[] rho, byte[] s) {
        GroupElement A_prim = get_proof().get_PoSCommitment().get_A_prim();
        GroupElement C_prim = get_proof().get_PoSCommitment().get_C_prim();
        GroupElement D_prim = get_proof().get_PoSCommitment().get_D_prim();
        int n_v = get_proof().get_ProtocolInformation().get_vbitlenro();

        ByteTree aPrim, cPrim, dPrim;

        if (get_proof().get_ProtocolInformation().get_pgroup().contains("ECqPGroup")) {
            aPrim = new Node((ECGroupElement) A_prim);
            cPrim = new Node((ECGroupElement) C_prim);
            dPrim = new Node((ECGroupElement) D_prim);

        } else {
            aPrim = new Leaf(A_prim);
            cPrim = new Leaf(C_prim);
            dPrim = new Leaf(D_prim);
        }

        ByteTree[] nodes = new ByteTree[]{
                new Node(get_proof().get_PoSCommitment().get_B()),
                aPrim,
                new Node(get_proof().get_PoSCommitment().get_B_prim()),
                cPrim,
                dPrim,
                new Node(get_proof().get_PoSCommitment().get_F_prim())
        };

        Node n = new Node(new ByteTree[]{new Leaf(s), new Node(nodes)});

        byte[] seed = new byte[rho.length + n.getEncodedLength()];
        System.arraycopy(rho, 0, seed, 0, rho.length);
        System.arraycopy(n.getEncoded(), 0, seed, rho.length, n.getEncodedLength());
        @SuppressWarnings("resource")
        RO ro = new RO(get_proof().get_ProtocolInformation().get_rohash(), seed);
        byte[] out = new byte[(n_v + 7) / 8];
        ro.read(out, n_v);
        return new BigInteger(1, out);
    }

    /**
     * See:
     * <a href="https://github.com/verificatum/verificatum-vmn/blob/9d9c45e37fa043881e58d71901ce45e387b31271/src/java/com/verificatum/protocol/hvzk/PoSBasicTW.java#L408">A</a>
     *
     * @param progress
     * @param e
     * @return
     * @throws MathException
     */
    public GroupElement compute_A(Progress progress, BigInteger[] e) throws MathException {
        GroupElement[] u = get_proof().get_PermutationCommitment().get_u();
        GroupElement res = u[0].getGroup().getIdentity();
        for (int i = 0; i < u.length; i++) {
            res = res.op(u[i].scale(e[i]));
            progress.increase(1);
        }
        return res;
    }

    /**
     * See:
     * <a href="https://github.com/verificatum/verificatum-vmn/blob/9d9c45e37fa043881e58d71901ce45e387b31271/src/java/com/verificatum/protocol/hvzk/PoSBasicTW.java#L1013">C</a>
     *
     * @param progress
     * @param h
     * @return
     * @throws MathException
     * @throws ShuffleException
     */
    public GroupElement compute_C(Progress progress, GroupElement[] h)
            throws MathException, ShuffleException {
        GroupElement[] u = get_proof().get_PermutationCommitment().get_u();

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

    /**
     * See:
     * <a href="https://github.com/verificatum/verificatum-vmn/blob/9d9c45e37fa043881e58d71901ce45e387b31271/src/java/com/verificatum/protocol/hvzk/PoSBasicTW.java#L1014">D</a>
     *
     * @param progress
     * @param h
     * @param e
     * @return
     * @throws ShuffleException
     * @throws MathException
     */
    public GroupElement compute_D(Progress progress, GroupElement[] h, BigInteger[] e)
            throws ShuffleException, MathException {
        BigInteger ep = BigInteger.ONE;
        BigInteger q;
        int N = get_proof().get_ciphertexts().length;

        if (h[0] instanceof ModPGroupElement) {
            ModPGroup group = (ModPGroup) h[0].getGroup();
            q = MathUtil.safePrimeOrder(group.getOrder());
        } else if (h[0] instanceof ECGroupElement) {
            q = h[0].getGroup().getOrder();
        } else {
            throw new ShuffleException("Unknown element type: " + h[0].getClass().getSimpleName());
        }

        for (BigInteger bigInteger : e) {
            ep = ep.multiply(bigInteger).mod(q);
            progress.increase(1);
        }

        GroupElement hp = h[0].scale(ep);
        hp = hp.inverse();
        GroupElement ret = get_proof().get_PoSCommitment().get_B()[N - 1].op(hp);

        progress.increase(3);
        return ret;
    }

    /**
     * See:
     * <a href="https://github.com/verificatum/verificatum-vmn/blob/9d9c45e37fa043881e58d71901ce45e387b31271/src/java/com/verificatum/protocol/hvzk/PoSBasicTW.java#L409">F</a>
     *
     * @param progress
     * @param e
     * @return
     * @throws MathException
     */
    public GroupElement compute_F(Progress progress, BigInteger[] e) throws MathException {
        GroupElement[] w = get_proof().get_ciphertexts();
        GroupElement res = w[0].getGroup().getIdentity();

        for (int i = 0; i < w.length; i++) {
            GroupElement exped = w[i].scale(e[i]);
            res = res.op(exped);
            progress.increase(1);
        }

        return res;
    }

    /**
     * See:
     * <a href="https://github.com/verificatum/verificatum-vmn/blob/9d9c45e37fa043881e58d71901ce45e387b31271/src/java/com/verificatum/protocol/hvzk/PoSBasicTW.java#L1020">verdictA</a>
     *
     * @param progress
     * @param v
     * @param A
     * @param h
     * @param e
     * @return
     * @throws MathException
     */
    public boolean verify_A(Progress progress, BigInteger v, GroupElement A, GroupElement[] h)
            throws MathException {
        GroupElement left = A.scale(v).op(get_proof().get_PoSCommitment().get_A_prim());
        progress.increase(1);
        GroupElement right = h[0].getGroup().getIdentity();

        for (int i = 0; i < h.length; i++) {
            right = right.op(h[i].scale(get_proof().get_PoSReply().get_kE()[i]));
            progress.increase(1);
        }

        right = get_proof().get_ProtocolInformation().get_parsed_generator()
                .scale(get_proof().get_PoSReply().get_kA())
                .op(right);

        progress.increase(1);
        return left.equals(right);
    }

    /**
     * See:
     * <a href="https://github.com/verificatum/verificatum-vmn/blob/9d9c45e37fa043881e58d71901ce45e387b31271/src/java/com/verificatum/protocol/hvzk/PoSBasicTW.java#L1035">verdictB</a>
     *
     * @param progress
     * @param v
     * @param h
     * @return
     * @throws MathException
     */
    public boolean verify_B(Progress progress, BigInteger v, GroupElement[] h)
            throws MathException {
        GroupElement[] B = get_proof().get_PoSCommitment().get_B();
        GroupElement[] Bp = get_proof().get_PoSCommitment().get_B_prim();
        BigInteger[] kB = get_proof().get_PoSReply().get_kB();
        BigInteger[] kE = get_proof().get_PoSReply().get_kE();
        GroupElement g = get_proof().get_ProtocolInformation().get_parsed_generator();

        for (int i = 1; i < B.length; i++) {
            GroupElement left = B[i].scale(v).op(Bp[i]);
            GroupElement right = B[i - 1].scale(kE[i]).op(g.scale(kB[i]));
            if (!left.equals(right)) {
                return false;
            }
            progress.increase(1);
        }

        return true;
    }

    /**
     * See:
     * <a href="https://github.com/verificatum/verificatum-vmn/blob/9d9c45e37fa043881e58d71901ce45e387b31271/src/java/com/verificatum/protocol/hvzk/PoSBasicTW.java#L1048">verdictC</a>
     *
     * @param progress
     * @param v
     * @param C
     * @return
     * @throws MathException
     */
    public boolean verify_C(Progress progress, BigInteger v, GroupElement C) throws MathException {
        GroupElement left = C.scale(v).op(get_proof().get_PoSCommitment().get_C_prim());
        progress.increase(1);

        GroupElement right = get_proof().get_ProtocolInformation().get_parsed_generator()
                .scale(get_proof().get_PoSReply().get_kC());
        progress.increase(1);

        return left.equals(right);
    }

    /**
     * See:
     * <a href="https://github.com/verificatum/verificatum-vmn/blob/9d9c45e37fa043881e58d71901ce45e387b31271/src/java/com/verificatum/protocol/hvzk/PoSBasicTW.java#L1055">verdictD</a>
     *
     * @param progress
     * @param v
     * @param D
     * @return
     * @throws MathException
     */
    public boolean verify_D(Progress progress, BigInteger v, GroupElement D) throws MathException {
        GroupElement left = D.scale(v).op(get_proof().get_PoSCommitment().get_D_prim());
        progress.increase(1);

        GroupElement right = get_proof().get_ProtocolInformation().get_parsed_generator()
                .scale(get_proof().get_PoSReply().get_kD());
        progress.increase(1);

        return left.equals(right);
    }

    /**
     * See:
     * <a href="https://github.com/verificatum/verificatum-vmn/blob/9d9c45e37fa043881e58d71901ce45e387b31271/src/java/com/verificatum/protocol/hvzk/PoSBasicTW.java#L1062">verdictF</a>
     *
     * @param progress
     * @param v
     * @param F
     * @return
     * @throws MathException
     */
    public boolean verify_F(Progress progress, BigInteger v, GroupElement F) throws MathException {
        // the number of computations differ in threaded and non-threaded case. In threaded case we
        // also aggregate the per-thread results.
        GroupElement pk = get_proof().get_publickey();
        BigInteger[] kF = get_proof().get_PoSReply().get_kF();
        GroupElement[] wp = get_proof().get_shuffled_ciphertexts();
        BigInteger[] kE = get_proof().get_PoSReply().get_kE();
        GroupElement Fp = get_proof().get_PoSCommitment().get_F_prim();

        GroupElement left = F.scale(v).op(Fp);
        GroupElement right = wp[0].getGroup().getIdentity();

        for (int i = 0; i < wp.length; i++) {
            right = right.op(wp[i].scale(kE[i]));
            progress.increase(1);
        }

        BigInteger[] factors = new BigInteger[kF.length];
        for (int i = 0; i < factors.length; i++) {
            factors[i] = kF[i].negate();
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
                2 * N + get_proof().get_PoSReply().get_kF().length + 3);
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
}
