package ee.ivxv.audit.util;

import ee.ivxv.common.crypto.elgamal.ElGamalDecryptionProof;
import ee.ivxv.common.model.Proof;
import ee.ivxv.common.util.*;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.Callable;
import java.util.concurrent.LinkedBlockingQueue;

/**
 * Class for handling ciphertexts with invalid proofs.
 */
public class InvalidDecProofs {
    private static final Path INVALID_PROOF_PATH = Paths.get("invalid");
    private final BlockingQueue<Object> in = new LinkedBlockingQueue<>();
    private final Proof invalidProofs;

    /**
     * Initialize using the election identifier.
     * 
     * @param electionName
     */
    public InvalidDecProofs(String electionName) {
        invalidProofs = new Proof(electionName);
    }

    /**
     * Get the worker for parsing all added proofs.
     * 
     * @return
     */
    public InvalidProofHandler getResultWorker() {
        return new InvalidProofHandler();
    }

    /**
     * Set the terminator after last added invalid proof.
     */
    public void setEot() {
        in.add(ee.ivxv.common.util.Util.EOT);
    }

    /**
     * Add invalid proof.
     * 
     * @param proof
     */
    public void addInvalidProof(ElGamalDecryptionProof proof) {
        in.add(proof);
    }

    /**
     * Get the total number of invalid proofs.
     * 
     * @return
     */
    public int getCount() {
        return invalidProofs.getCount();
    }

    /**
     * Serialize the structure to directory.
     * 
     * @param outDir
     * @throws Exception
     */
    public void outputInvalidProofs(Path outDir) throws Exception {
        Json.write(invalidProofs, outDir.resolve(INVALID_PROOF_PATH));
    }

    private class InvalidProofHandler implements Callable<Void> {

        @Override
        public Void call() throws Exception {
            Object obj;
            while ((obj = in.take()) != ee.ivxv.common.util.Util.EOT) {
                if (obj instanceof ElGamalDecryptionProof) {
                    invalidProofs.addProof((ElGamalDecryptionProof) obj);
                } else {
                    throw new IllegalArgumentException(
                            "Unexpected decryption result type: " + obj.getClass());
                }
            }
            return null;
        }
    }
}
