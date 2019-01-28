package ee.ivxv.common.model;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;
import ee.ivxv.common.crypto.elgamal.ElGamalDecryptionProof;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

/**
 * JSON serializable list of proof of correct decryptions.
 */
public class Proof {
    private final String election;
    private final List<ProofJson> proofs;

    /**
     * Initialize using election identifier.
     * 
     * @param election
     */
    public Proof(String election) {
        this.election = election;
        this.proofs = new ArrayList<>();
    }

    /**
     * Initialize using serialized values.
     * 
     * @param election
     * @param proofs
     */
    @JsonCreator
    public Proof( //
            @JsonProperty("election") String election, //
            @JsonProperty("proofs") List<ProofJson> proofs) {
        this.election = election;
        this.proofs = Collections.unmodifiableList(proofs);
    }

    /**
     * Get election identifier.
     * 
     * @return
     */
    public String getElection() {
        return election;
    }

    /**
     * Get list of serializable decryption proofs.
     * 
     * @return
     */
    public List<ProofJson> getProofs() {
        return proofs;
    }

    /**
     * Add decryption proof.
     * 
     * @param proof
     */
    public void addProof(ElGamalDecryptionProof proof) {
        proofs.add(new ProofJson(proof.ciphertext.getBytes(),
                proof.decrypted.getUTF8DecodedMessage(), proof.getBytes()));
    }

    /**
     * Get the number of decryption proofs.
     * 
     * @return
     */
    @JsonIgnore
    public int getCount() {
        return proofs.size();
    }

    /**
     * JSON serializable representation of single decryption proof.
     */
    public static class ProofJson {
        private final byte[] ciphertext;
        private final String message;
        private final byte[] proof;

        /**
         * Initialize using serialized values.
         * 
         * @param ciphertext
         * @param message
         * @param proof
         */
        @JsonCreator
        private ProofJson( //
                @JsonProperty("ciphertext") byte[] ciphertext, //
                @JsonProperty("message") String message, //
                @JsonProperty("proof") byte[] proof) {
            this.ciphertext = ciphertext;
            this.message = message;
            this.proof = proof;
        }

        /**
         * Get the ciphertext for which the proof is for.
         * 
         * @return
         */
        public byte[] getCiphertext() {
            return ciphertext;
        }

        /**
         * Get the message for which the proof is for.
         * 
         * @return
         */
        public String getMessage() {
            return message;
        }

        /**
         * Get serialized proof.
         * 
         * @return
         */
        public byte[] getProof() {
            return proof;
        }
    }


}
