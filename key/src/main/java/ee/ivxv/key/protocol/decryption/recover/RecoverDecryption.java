package ee.ivxv.key.protocol.decryption.recover;

import ee.ivxv.common.crypto.Plaintext;
import ee.ivxv.common.crypto.elgamal.ElGamalCiphertext;
import ee.ivxv.common.crypto.elgamal.ElGamalDecryptionProof;
import ee.ivxv.common.crypto.elgamal.ElGamalParameters;
import ee.ivxv.common.crypto.elgamal.ElGamalPrivateKey;
import ee.ivxv.common.math.LagrangeInterpolation;
import ee.ivxv.common.math.MathException;
import ee.ivxv.common.service.smartcard.Cards;
import ee.ivxv.key.protocol.DecryptionProtocol;
import ee.ivxv.key.protocol.ProtocolException;
import ee.ivxv.key.protocol.ProtocolUtil;
import ee.ivxv.key.protocol.ThresholdParameters;
import java.io.IOException;
import java.math.BigInteger;

// this protocol gets the keyshares from the cards, reconstructs the key in
// memory and then performs software decryption
public class RecoverDecryption implements DecryptionProtocol {
    private final Cards cards;
    private final ThresholdParameters tparams;
    private final byte[] cardShareAID;
    private final byte[] cardShareName;
    private final boolean withProof;
    private ElGamalPrivateKey sk;

    public RecoverDecryption(Cards cards, ThresholdParameters tparams, byte[] cardShareAID,
            byte[] cardShareName) throws ProtocolException {
        this(cards, tparams, cardShareAID, cardShareName, true);
    }

    public RecoverDecryption(Cards cards, ThresholdParameters tparams, byte[] cardShareAID,
            byte[] cardShareName, boolean withProof) throws ProtocolException {
        if (cards.count() < tparams.getThreshold()) {
            throw new ProtocolException("Fewer cards available than threshold");
        }
        this.cards = cards;
        this.tparams = tparams;
        this.cardShareAID = cardShareAID;
        this.cardShareName = cardShareName;
        this.withProof = withProof;
        recoverKey();
    }

    private void recoverKey() throws ProtocolException {
        if (this.sk == null) {
            this.sk = forceKeyRecover();
        }
    }

    private ElGamalPrivateKey forceKeyRecover() throws ProtocolException {
        byte[][] blobs = ProtocolUtil.getAllBlobs(tparams, cards, cardShareAID, cardShareName);
        ElGamalPrivateKey[] parsedKeys = parseAllBlobs(blobs);
        if (!filterKeys(parsedKeys)) {
            throw new ProtocolException("Key share parameters mismatch");
        }
        ElGamalPrivateKey secretKey = combineKeys(parsedKeys);
        return secretKey;
    }

    private ElGamalPrivateKey[] parseAllBlobs(byte[][] blobs) throws ProtocolException {
        ElGamalPrivateKey[] parsed = new ElGamalPrivateKey[blobs.length];
        for (int i = 0; i < blobs.length; i++) {
            if (blobs[i] == null) {
                parsed[i] = null;
                continue;
            }
            try {
                parsed[i] = new ElGamalPrivateKey(blobs[i]);
            } catch (IllegalArgumentException e) {
                throw new ProtocolException(
                        "Exception while parsing secret share: " + e.toString());
            }

        }
        return parsed;
    }

    private boolean filterKeys(ElGamalPrivateKey[] keys) throws ProtocolException {
        // this method checks that the parameters for all the keys are the
        // same. If not, then we do not perform any recovery.
        if (keys.length == 0) {
            throw new ProtocolException("No key shares ready for filtering");
        }
        ElGamalParameters params = null;
        for (ElGamalPrivateKey key : keys) {
            if (key == null) {
                continue;
            }
            if (params == null) {
                params = key.getParameters();
            } else if (!params.equals(key.getParameters())) {
                return false;
            }
        }
        return true;
    }

    private ElGamalPrivateKey combineKeys(ElGamalPrivateKey[] keys) throws ProtocolException {
        if (keys.length == 0) {
            throw new ProtocolException("No key shares ready for filtering");
        }
        if (keys.length < tparams.getThreshold()) {
            throw new ProtocolException("Fewer keys available than is the threshold");
        }
        ElGamalParameters params = null;
        BigInteger k = BigInteger.ZERO;
        for (int i = 0; i < keys.length; i++) {
            if (keys[i] == null) {
                continue;
            }
            params = keys[i].getParameters();
            BigInteger p = keys[i].getSecretPart();
            p = p.multiply(LagrangeInterpolation.basisPolynomial(params.getGeneratorOrder(), keys,
                    BigInteger.valueOf(i + 1)));
            k = k.add(p).mod(params.getGeneratorOrder());
        }
        return new ElGamalPrivateKey(params, k);
    }

    @Override
    public ElGamalDecryptionProof decryptMessage(byte[] msg)
            throws ProtocolException, IllegalArgumentException, IOException {
        if (this.sk == null) {
            throw new ProtocolException("Secret key not reconstructed");
        }
        ElGamalCiphertext ct = new ElGamalCiphertext(this.sk.getParameters(), msg);
        try {
            if (withProof) {
                return this.sk.provableDecrypt(ct);
            } else {
                Plaintext pt = this.sk.decrypt(ct);
                return new ElGamalDecryptionProof(ct, pt, this.sk.getPublicKey());
            }
        } catch (MathException e) {
            throw new ProtocolException("Arithmetic error: " + e.toString());
        }
    }
}
