package ee.ivxv.key.model;

import com.fasterxml.jackson.annotation.JsonIgnore;
import ee.ivxv.common.crypto.elgamal.ElGamalDecryptionProof;

/**
 * JSON serializable structure for holding the encrypted vote using metadata and decryption result.
 */
public class Vote {
    private final String district;
    private final String station;
    private final String question;
    private final byte[] vote;
    @JsonIgnore
    private ElGamalDecryptionProof proof;

    /**
     * Initialize using values
     * 
     * @param district District identifier
     * @param station Station identifier
     * @param question Question identifier
     * @param vote Decrypted vote
     */
    public Vote(String district, String station, String question, byte[] vote) {
        this.district = district;
        this.station = station;
        this.question = question;
        this.vote = vote;
    }

    /**
     * Get the district identifier of the vote.
     * 
     * @return
     */
    public String getDistrict() {
        return district;
    }

    /**
     * Get the station identifier of the vote.
     * 
     * @return
     */
    public String getStation() {
        return station;
    }

    /**
     * Get the question identifier of the vote.
     * 
     * @return
     */
    public String getQuestion() {
        return question;
    }

    /**
     * Get the vote.
     * 
     * @return
     */
    public byte[] getVote() {
        return vote;
    }

    /**
     * Get the decryption proof.
     * <p>
     * Depending on the decryption protocol, the proof of correct decryption value may be null. The
     * values for ciphertext, decryption and public key must be set.
     * 
     * @return
     */
    public ElGamalDecryptionProof getProof() {
        return proof;
    }

    /**
     * Set the decryption proof.
     * 
     * @param proof
     */
    public void setProof(ElGamalDecryptionProof proof) {
        this.proof = proof;
    }
}
