package ee.ivxv.audit.shuffle;

import ee.ivxv.common.crypto.elgamal.ElGamalCiphertext;
import ee.ivxv.common.crypto.elgamal.ElGamalPublicKey;
import ee.ivxv.common.math.Group;
import ee.ivxv.common.math.GroupElement;
import ee.ivxv.common.math.ProductGroup;
import ee.ivxv.common.math.ProductGroupElement;
import java.io.IOException;
import java.math.BigInteger;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;

/**
 * Class for holding variables for non-interactive shuffle proof.
 */
public class ShuffleProof {
    private final ProtocolInformation prot;
    private final PermutationCommitment pc;
    private final PoSCommitment posc;
    private final PoSReply posr;
    private final GroupElement[] ciphs, shuffled;
    private final GroupElement pk;

    public static final String CIPHERTEXTS_PATH = "Ciphertexts.bt";
    public static final String SHUFFLED_CIPHERTEXTS_PATH = "ShuffledCiphertexts.bt";
    public static final String PUBLICKEY_PATH = "FullPublicKey.bt";
    public static final String PROOFS_PATH = "proofs";
    public static final String PC_PATH = "PermutationCommitment01.bt";
    public static final String POSC_PATH = "PoSCommitment01.bt";
    public static final String POSR_PATH = "PoSReply01.bt";

    public ShuffleProof(ProtocolInformation prot, PermutationCommitment pc, PoSCommitment posc,
            PoSReply posr, GroupElement[] ciphs, GroupElement[] shuffled, GroupElement pk) {
        this.prot = prot;
        this.pc = pc;
        this.posc = posc;
        this.posr = posr;
        this.ciphs = ciphs;
        this.shuffled = shuffled;
        this.pk = pk;
    }

    public ShuffleProof(ProtocolInformation prot, PermutationCommitment pc, PoSCommitment posc,
            PoSReply posr, ElGamalCiphertext[] ciphs, ElGamalCiphertext[] shuffled,
            ElGamalPublicKey pk) {
        if (ciphs.length != shuffled.length) {
            throw new IllegalArgumentException("Ciphertext and shuffled length does not match");
        }
        GroupElement[] cge = new GroupElement[ciphs.length];
        GroupElement[] scge = new GroupElement[shuffled.length];
        for (int i = 0; i < ciphs.length; i++) {
            cge[i] = ciphs[i].getAsProductGroupElement();
            scge[i] = shuffled[i].getAsProductGroupElement();
        }
        this.prot = prot;
        this.pc = pc;
        this.posc = posc;
        this.posr = posr;
        this.ciphs = cge;
        this.shuffled = scge;
        this.pk = pk.getAsProductGroupElement();
    }

    public ShuffleProof(Path protpath, Path proofdir) throws IOException, ShuffleException {
        Path pcpath = Paths.get(proofdir.toString(), PROOFS_PATH, PC_PATH);
        Path poscpath = Paths.get(proofdir.toString(), PROOFS_PATH, POSC_PATH);
        Path posrpath = Paths.get(proofdir.toString(), PROOFS_PATH, POSR_PATH);
        Path ciphspath = Paths.get(proofdir.toString(), CIPHERTEXTS_PATH);
        Path shuffledpath = Paths.get(proofdir.toString(), SHUFFLED_CIPHERTEXTS_PATH);
        Path pkpath = Paths.get(proofdir.toString(), PUBLICKEY_PATH);
        prot = new ProtocolInformation(protpath);
        pc = new PermutationCommitment(prot, pcpath);
        posc = new PoSCommitment(prot, poscpath);
        posr = new PoSReply(prot, posrpath);
        ByteTree ctsbt = ByteTree.parse(Files.readAllBytes(ciphspath));
        ByteTree sctsbt = ByteTree.parse(Files.readAllBytes(shuffledpath));
        ByteTree pkbt = ByteTree.parse(Files.readAllBytes(pkpath));
        Group group = prot.get_parsed_pgroup();
        Group ciphgroup = get_ciphertext_group(prot, group);
        ciphs = DataParser.getAsElementArray(ciphgroup, ctsbt);
        shuffled = DataParser.getAsElementArray(ciphgroup, sctsbt);
        pk = DataParser.getAsElement(ciphgroup, pkbt);
    }

    public ProtocolInformation get_ProtocolInformation() {
        return prot;
    }

    public PermutationCommitment get_PermutationCommitment() {
        return pc;
    }

    public PoSCommitment get_PoSCommitment() {
        return posc;
    }

    public PoSReply get_PoSReply() {
        return posr;
    }

    public GroupElement[] get_ciphertexts() {
        return ciphs;
    }

    public GroupElement[] get_shuffled_ciphertexts() {
        return shuffled;
    }

    public GroupElement get_publickey() {
        return pk;
    }

    private static ProductGroup get_ciphertext_group(ProtocolInformation prot, Group group) {
        ProductGroup prodgroup = new ProductGroup(group, prot.get_keywidth());
        ProductGroup ciphgroup = new ProductGroup(prodgroup, 2);
        return ciphgroup;
    }

    /**
     * Permutation commitment holder.
     */
    public static class PermutationCommitment {
        private final GroupElement[] u;

        /**
         * Parse permutation commitment from a bytetree.
         * 
         * @param bt
         */
        PermutationCommitment(ProtocolInformation prot, byte[] bt) {
            ByteTree root = ByteTree.parse(bt);
            this.u = DataParser.getAsElementArray(prot.get_parsed_pgroup(), root);
        }

        PermutationCommitment(ProtocolInformation prot, Path path) throws IOException {
            this(prot, Files.readAllBytes(path));
        }

        public GroupElement[] get_u() {
            return u;
        }
    }

    /**
     * Proof of shuffle commitment holder.
     */
    public static class PoSCommitment {
        private GroupElement A_prim, C_prim, D_prim;
        private GroupElement[] B, B_prim;
        private ProductGroupElement F_prim;

        /**
         * Parse proof of shuffle commitment from bytetree.
         * 
         * @param bt Byte array holding proof of shuffle commitment.
         */
        PoSCommitment(ProtocolInformation prot, byte[] bt) {
            ByteTree root = ByteTree.parse(bt);
            Group group = prot.get_parsed_pgroup();
            ProductGroup ciphgroup = get_ciphertext_group(prot, group);
            B = DataParser.getAsElementArray(group, root, 0);
            A_prim = DataParser.getAsElement(group, root, 1);
            B_prim = DataParser.getAsElementArray(group, root, 2);
            C_prim = DataParser.getAsElement(group, root, 3);
            D_prim = DataParser.getAsElement(group, root, 4);
            F_prim = (ProductGroupElement) DataParser.getAsElement(ciphgroup, root, 5);
        }

        PoSCommitment(ProtocolInformation prot, Path path) throws IOException {
            this(prot, Files.readAllBytes(path));
        }

        public GroupElement get_A_prim() {
            return A_prim;
        }

        public GroupElement get_C_prim() {
            return C_prim;
        }

        public GroupElement get_D_prim() {
            return D_prim;
        }

        public GroupElement[] get_B() {
            return B;
        }

        public GroupElement[] get_B_prim() {
            return B_prim;
        }

        public ProductGroupElement get_F_prim() {
            return F_prim;
        }
    }

    /**
     * Holder for proof of shuffle reply.
     */
    public static class PoSReply {
        private BigInteger kA, kC, kD;
        private BigInteger[] kB, kE;
        // strictly taken - kF is not from the group but is an array (of array) of bigints. take
        // into account when using in computation
        private ProductGroupElement kF;

        /**
         * Parse proof of shuffle reply from bytetree.
         * 
         * @param bt Byte array holding proof of shuffle in bytetree format.
         */
        PoSReply(ProtocolInformation prot, byte[] bt) {
            ByteTree root = ByteTree.parse(bt);
            Group group = prot.get_parsed_pgroup();
            ProductGroup prodgroup = new ProductGroup(group, prot.get_keywidth());
            kA = DataParser.getAsInteger(root, 0);
            kB = DataParser.getAsIntegerArray(root, 1);
            kC = DataParser.getAsInteger(root, 2);
            kD = DataParser.getAsInteger(root, 3);
            kE = DataParser.getAsIntegerArray(root, 4);
            kF = (ProductGroupElement) DataParser.getAsElement(prodgroup, root, 5);
        }

        PoSReply(ProtocolInformation prot, Path path) throws IOException {
            this(prot, Files.readAllBytes(path));
        }

        public BigInteger get_kA() {
            return kA;
        }

        public BigInteger get_kC() {
            return kC;
        }

        public BigInteger get_kD() {
            return kD;
        }

        public BigInteger[] get_kB() {
            return kB;
        }

        public BigInteger[] get_kE() {
            return kE;
        }

        public ProductGroupElement get_kF() {
            return kF;
        }
    }
}
