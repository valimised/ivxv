package ee.ivxv.audit.util;

import ee.ivxv.common.crypto.elgamal.ElGamalDecryptionProof;
import ee.ivxv.common.model.Proof;
import ee.ivxv.common.util.*;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.Callable;
import java.util.concurrent.LinkedBlockingQueue;

public class InvalidDecProofs {
    private static final Path INVALID_PROOF_PATH = Paths.get("invalid");
    private final BlockingQueue<Object> in = new LinkedBlockingQueue<>();
    private final Proof invalidProofs;

    public InvalidDecProofs(String electionName) {
        invalidProofs = new Proof(electionName);
    }

    public InvalidProofHandler getResultWorker() {
        return new InvalidProofHandler();
    }

    public void setEot() {
        in.add(ee.ivxv.common.util.Util.EOT);
    }

    public void addInvalidProof(ElGamalDecryptionProof proof) {
        in.add(proof);
    }

    public int getCount() {
        return invalidProofs.getCount();
    }

    public void outputInvalidProofs(Path outDir) throws Exception {
        Json.write(invalidProofs, outDir.resolve(INVALID_PROOF_PATH));
    }

    private class InvalidProofHandler implements Callable<Void> {

        @Override public Void call() throws Exception {
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
