package ee.ivxv.common.crypto.elgamal;

import ee.ivxv.common.asn1.ASN1DecodingException;
import ee.ivxv.common.asn1.Field;
import ee.ivxv.common.asn1.Sequence;
import ee.ivxv.common.crypto.Plaintext;
import ee.ivxv.common.crypto.hash.HashFunction;
import ee.ivxv.common.crypto.hash.Sha256;
import ee.ivxv.common.math.GroupElement;
import ee.ivxv.common.math.MathException;
import java.math.BigInteger;

/**
 * ElGamalDecryptionProof holds the decrypted message and proof of correct decryption.
 * <p>
 * To prove the correctness of decryption, the prover has to prove two aspects - that it knows the
 * private key and that the private key was used for decrypting the value. To give the simultaneous
 * proof of both, the prover at first commits to the private key and message, publishes the
 * commitments. Then, the verifier sends a challenge to the prover and the prover computes the
 * response. Then, the verifier verifies the response.
 * <p>
 * Futhermore, as the proof of decryption must be non-interactive, then the challenge is computed
 * from the previous messages in the protocol.
 * <p>
 * For more detailed description of the proof protocol, see the framework documentation.
 */
public class ElGamalDecryptionProof {
    private static final byte[] NI_PROOF_DOMAIN = new Field("DECRYPTION").encode();
    private final HashFunction hashFun = new Sha256();
    public final ElGamalPublicKey publickey;
    public ElGamalCiphertext ciphertext;
    public Plaintext decrypted;
    // this is the a in the whitepaper
    public GroupElement msgCommitment;
    // this is the b in the whitepaper
    public GroupElement keyCommitment;
    // this is the s in the whitepaper
    public BigInteger response;

    /**
     * Initialize using all proof components.
     * <p>
     * Use this constructor if the proof is already constructed and the goal is to verify the
     * correctness of decryption.
     * 
     * @param ciphertext
     * @param decrypted
     * @param publickey
     * @param msgCommitment
     * @param keyCommitment
     * @param response
     */
    ElGamalDecryptionProof(ElGamalCiphertext ciphertext, Plaintext decrypted,
            ElGamalPublicKey publickey, GroupElement msgCommitment, GroupElement keyCommitment,
            BigInteger response) {
        this.ciphertext = ciphertext;
        this.decrypted = decrypted;
        this.publickey = publickey;
        this.msgCommitment = msgCommitment;
        this.keyCommitment = keyCommitment;
        this.response = response;
    }

    /**
     * Initialize using all proof components where commitments are given as an array.
     * <p>
     * Use this constructor if the proof is already constructed and the goal is to verify the
     * correctness of decryption.
     * 
     * @param ciphertext
     * @param decrypted
     * @param publickey
     * @param commitments
     * @param response
     * @throws IllegalArgumentException
     */
    ElGamalDecryptionProof(ElGamalCiphertext ciphertext, Plaintext decrypted,
            ElGamalPublicKey publickey, GroupElement[] commitments, BigInteger response)
            throws IllegalArgumentException {
        this(ciphertext, decrypted, publickey);
        if (commitments.length != 2) {
            throw new IllegalArgumentException("Commitments must consist of two elements");
        }
        this.msgCommitment = commitments[0];
        this.keyCommitment = commitments[1];
        this.response = response;
    }

    /**
     * Initialize using serialized value.
     * <p>
     * Use this constructor if the proof is already constructed and serialized.
     * 
     * @see #getBytes()
     * 
     * @param ciphertext
     * @param decrypted
     * @param publickey
     * @param packed
     */
    public ElGamalDecryptionProof(ElGamalCiphertext ciphertext, Plaintext decrypted,
            ElGamalPublicKey publickey, byte[] packed) {
        this(ciphertext, decrypted, publickey);
        Sequence seq = new Sequence();
        byte[][] parts;
        try {
            seq.readFromBytes(packed);
            parts = seq.getBytes();
        } catch (ASN1DecodingException e) {
            throw new IllegalArgumentException("Parsing proof failed: " + e.toString());
        }
        if (parts.length != 3) {
            throw new IllegalArgumentException("Parsing proof failed: invalid proof length");
        }
        this.msgCommitment = publickey.getParameters().getGroup().getElement(parts[0]);
        this.keyCommitment = publickey.getParameters().getGroup().getElement(parts[1]);
        Field respField = new Field();
        try {
            respField.readFromBytes(parts[2]);
            this.response = respField.getInteger();
        } catch (ASN1DecodingException e) {
            throw new IllegalArgumentException("Parsing proof failed: " + e.toString());
        }
    }

    /**
     * Initialize using components for creating a proof.
     * 
     * @param ciphertext
     * @param decrypted
     * @param publickey
     */
    public ElGamalDecryptionProof(ElGamalCiphertext ciphertext, Plaintext decrypted,
            ElGamalPublicKey publickey) {
        this.ciphertext = ciphertext;
        this.decrypted = decrypted;
        this.publickey = publickey;
    }

    /**
     * Initialize using components for creating a proof.
     * 
     * @param publickey
     */
    ElGamalDecryptionProof(ElGamalPublicKey publickey) {
        this.publickey = publickey;
    }

    /**
     * Set the ciphertext.
     * 
     * @param ciphertext
     */
    public void setCiphertext(ElGamalCiphertext ciphertext) {
        this.ciphertext = ciphertext;
    }

    /**
     * Get the decrypted value.
     * 
     * @return
     */
    public Plaintext getDecrypted() {
        return this.decrypted;
    }

    /**
     * Set the decrypted value.
     * 
     * @param decrypted
     */
    public void setDecrypted(Plaintext decrypted) {
        this.decrypted = decrypted;
    }

    /**
     * Set the message commitment.
     * <p>
     * The message commitment should be the value {@literal blind^r}
     * 
     * @param com
     */
    public void setMessageCommitment(GroupElement com) {
        this.msgCommitment = com;
    }

    /**
     * Set the key commitment.
     * <p>
     * The key commitment should be the value {@literal g^r}.
     * 
     * @param com
     */
    public void setKeyCommitment(GroupElement com) {
        this.keyCommitment = com;
    }

    /**
     * Set the prover response.
     * <p>
     * The prover response should be the value {@literal k*x+r}, where {@literal k} is the
     * challenge.
     * 
     * @param resp
     */
    public void setResponse(BigInteger resp) {
        this.response = resp;
    }

    /**
     * Verify the correctness of proofs.
     * 
     * @see #verifyDecryptionProof(BigInteger)
     * @see #verifySecretKeyProof(BigInteger)
     * 
     * @return Boolean indicating if both proofs verified.
     * @throws MathException
     */
    public boolean verifyProof() throws MathException {
        BigInteger k = computeChallenge();
        return verifySecretKeyProof(k) && verifyDecryptionProof(k);
    }

    /**
     * Verify the proof of knowledge of the secret key
     * 
     * @param challenge
     * @return
     * @throws MathException
     */
    public boolean verifySecretKeyProof(BigInteger challenge) throws MathException {
        // verify that c1^s = a * (c2/d)^k
        ElGamalParameters params = publickey.getParameters();
        GroupElement d = params.getGroup().encode(params.getGroup().pad(decrypted));
        GroupElement left = ciphertext.getBlind().scale(response);
        GroupElement right =
                d.inverse().op(ciphertext.getBlindedMessage()).scale(challenge).op(msgCommitment);
        return left.equals(right);
    }

    /**
     * Verify the proof of correct decryption.
     * 
     * @param challenge
     * @return
     * @throws MathException
     */
    public boolean verifyDecryptionProof(BigInteger challenge) throws MathException {
        // verify that g^s = b*y^k
        GroupElement left = publickey.getParameters().getGenerator().scale(response);
        GroupElement right = publickey.getKey().scale(challenge).op(keyCommitment);
        return left.equals(right);
    }

    /**
     * Compute the challenge for the prover.
     * <p>
     * The challenge is the value {@literal H("DECRYPTION" || pk || ct || dec || a || b)}, where
     * {@literal H} is a hash function, {@literal "DECRYPTION"} is a domain separation value in
     * bytes, {@literal pk} is the serialized public key, {@literal ct} is the serialzied
     * ciphertext, {@literal dec} is the serialized decrypted message, {@literal a} is the
     * serialized message commitment and {@literal b} is the key commitment.
     * 
     * @return Challenge encoded as a positive BigInteger
     */
    public BigInteger computeChallenge() {
        byte[] hashSeed = new Sequence(
                // NI_PROOF_DOMAIN is a bytes representation of word
                // "DECRYPTION" to separate NI instances
                NI_PROOF_DOMAIN,
                // public key parameters bytes representation consists of:
                // - key type OID (mod p or EC)
                // - key parameters
                // * order
                // * generator
                // * election identifier
                // - public key byte representation (depends on group
                // implementation)
                publickey.getBytes(),
                // ciphertext byte representation consists of:
                // - key type OID (mod p or EC)
                // - blind byte representation (group dependable)
                // - blinded message byte representation (group dependable)
                ciphertext.getBytes(),
                // decrypted bytes representation consists of:
                // - asn1 encoding of the plaintext
                decrypted.getBytes(),
                // msg ja key commitments consists of bytes representation
                // of group elements (group specific)
                msgCommitment.getBytes(), keyCommitment.getBytes()).encode();
        byte[] hash = hashFun.digest(hashSeed);
        return new BigInteger(1, hash);
    }

    /**
     * Serialize the proof.
     * <p>
     * Structure:
     * {@code
     * SEQUENCE (
     *   msgCommitment GroupElement
     *   keyCommitment GroupElement
     *   response      INTEGER
     *   )
     * }
     * 
     * @return
     */
    public byte[] getBytes() {
        return new Sequence(msgCommitment.getBytes(), keyCommitment.getBytes(),
                new Field(response).encode()).encode();
    }
}
