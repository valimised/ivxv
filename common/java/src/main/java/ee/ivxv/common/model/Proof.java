package ee.ivxv.common.model;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;
import ee.ivxv.common.crypto.elgamal.ElGamalDecryptionProof;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

public class Proof {
    private final String election;
    private final List<ProofJson> proofs;

    public Proof(String election) {
        this.election = election;
        this.proofs = new ArrayList<>();
    }

    @JsonCreator
    public Proof( //
            @JsonProperty("election") String election, //
            @JsonProperty("proofs") List<ProofJson> proofs) {
        this.election = election;
        this.proofs = Collections.unmodifiableList(proofs);
    }

    public String getElection() {
        return election;
    }

    public List<ProofJson> getProofs() {
        return proofs;
    }

    public void addProof(ElGamalDecryptionProof proof) {
        proofs.add(new ProofJson(proof.ciphertext.getBytes(),
                proof.decrypted.getUTF8DecodedMessage(), proof.getBytes()));
    }

    @JsonIgnore
    public int getCount() {
        return proofs.size();
    }

    public static class ProofJson {
        private final byte[] ciphertext;
        private final String message;
        private final byte[] proof;

        @JsonCreator
        private ProofJson( //
                @JsonProperty("ciphertext") byte[] ciphertext, //
                @JsonProperty("message") String message, //
                @JsonProperty("proof") byte[] proof) {
            this.ciphertext = ciphertext;
            this.message = message;
            this.proof = proof;
        }

        public byte[] getCiphertext() {
            return ciphertext;
        }

        public String getMessage() {
            return message;
        }

        public byte[] getProof() {
            return proof;
        }
    }


}
