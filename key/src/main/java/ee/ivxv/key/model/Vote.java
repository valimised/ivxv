package ee.ivxv.key.model;

import com.fasterxml.jackson.annotation.JsonIgnore;
import ee.ivxv.common.crypto.elgamal.ElGamalDecryptionProof;

public class Vote {
    private final String district;
    private final String station;
    private final String question;
    private final byte[] vote;
    @JsonIgnore
    private ElGamalDecryptionProof proof;

    public Vote(String district, String station, String question, byte[] vote) {
        this.district = district;
        this.station = station;
        this.question = question;
        this.vote = vote;
    }

    public String getDistrict() {
        return district;
    }

    public String getStation() {
        return station;
    }

    public String getQuestion() {
        return question;
    }

    public byte[] getVote() {
        return vote;
    }

    public ElGamalDecryptionProof getProof() {
        return proof;
    }

    public void setProof(ElGamalDecryptionProof proof) {
        this.proof = proof;
    }
}
