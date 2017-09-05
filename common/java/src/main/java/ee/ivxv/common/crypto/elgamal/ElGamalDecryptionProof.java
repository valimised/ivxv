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

    // this constructor is for a ready-made proof
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

    // this constructor is for a ready-made proof, but the proof is given as an
    // array instead of separate values
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

    // this constructor is for a ready-made and asn1 packed proof
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

    // this constructor is for a proof-to-be-made. we set the values we know so
    // that we could compute the Fiat-Shamir challenge
    public ElGamalDecryptionProof(ElGamalCiphertext ciphertext, Plaintext decrypted,
            ElGamalPublicKey publickey) {
        this.ciphertext = ciphertext;
        this.decrypted = decrypted;
        this.publickey = publickey;
    }

    // this constructor is for a proof-to-be-made, but the API allows setting
    // every value sequentially.
    ElGamalDecryptionProof(ElGamalPublicKey publickey) {
        this.publickey = publickey;
    }

    public void setCiphertext(ElGamalCiphertext ciphertext) {
        this.ciphertext = ciphertext;
    }

    public Plaintext getDecrypted() {
        return this.decrypted;
    }

    public void setDecrypted(Plaintext decrypted) {
        this.decrypted = decrypted;
    }

    public void setMessageCommitment(GroupElement com) {
        this.msgCommitment = com;
    }

    public void setKeyCommitment(GroupElement com) {
        this.keyCommitment = com;
    }

    public void setResponse(BigInteger resp) {
        this.response = resp;
    }

    public boolean verifyProof() throws MathException {
        BigInteger k = computeChallenge();
        return verifySecretKeyProof(k) && verifyDecryptionProof(k);
    }

    public boolean verifySecretKeyProof(BigInteger challenge) throws MathException {
        // verify that c1^s = a * (c2/d)^k
        ElGamalParameters params = publickey.getParameters();
        GroupElement d = params.getGroup().encode(params.getGroup().pad(decrypted));
        GroupElement left = ciphertext.getBlind().scale(response);
        GroupElement right =
                d.inverse().op(ciphertext.getBlindedMessage()).scale(challenge).op(msgCommitment);
        return left.equals(right);
    }

    public boolean verifyDecryptionProof(BigInteger challenge) throws MathException {
        // verify that g^s = b*y^k
        GroupElement left = publickey.getParameters().getGenerator().scale(response);
        GroupElement right = publickey.getKey().scale(challenge).op(keyCommitment);
        return left.equals(right);
    }

    public BigInteger computeChallenge() {
        byte[] hashSeed =
            new Sequence(
                    // NI_PROOF_DOMAIN is a bytes representation of word
                    // "DECRYPTION" to separate NI instances
                    NI_PROOF_DOMAIN,
                    // public key parameters bytes representation consists of:
                    // - key type OID (mod p or EC)
                    // - key parameters
                    //   * order
                    //   * generator
                    //   * election identifier
                    // - public key byte representation (depends on group
                    //   implementation)
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
                    msgCommitment.getBytes(),
                    keyCommitment.getBytes()
                    ).encode();
        byte[] hash = hashFun.digest(hashSeed);
        return new BigInteger(1, hash);
    }

    public byte[] getBytes() {
        return new Sequence(
                msgCommitment.getBytes(),
                keyCommitment.getBytes(),
                new Field(response).encode()).encode();
    }
}
