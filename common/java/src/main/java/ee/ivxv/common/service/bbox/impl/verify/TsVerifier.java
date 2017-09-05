package ee.ivxv.common.service.bbox.impl.verify;

import ee.ivxv.common.crypto.CryptoUtil.PublicKeyHolder;
import ee.ivxv.common.service.bbox.Result;
import ee.ivxv.common.service.bbox.impl.ResultException;
import java.math.BigInteger;
import org.bouncycastle.asn1.tsp.MessageImprint;
import org.bouncycastle.asn1.x509.AlgorithmIdentifier;
import org.bouncycastle.operator.DefaultDigestAlgorithmIdentifierFinder;
import org.bouncycastle.operator.DigestAlgorithmIdentifierFinder;
import org.bouncycastle.tsp.TimeStampToken;
import org.bouncycastle.tsp.TimeStampTokenInfo;

public class TsVerifier {

    private static final String NONE_WITH_RSA = "NONEWithRsa";
    private static final DigestAlgorithmIdentifierFinder digestAlgFinder =
            new DefaultDigestAlgorithmIdentifierFinder();

    private final PublicKeyHolder key;

    public TsVerifier(PublicKeyHolder key) {
        this.key = key;
    }

    public TimeStampToken verify(TimeStampToken token) throws ResultException {
        try {
            return verifyInternal(token);
        } catch (ResultException e) {
            throw e;
        } catch (Exception e) {
            throw new ResultException(Result.REG_RESP_INVALID, e);
        }
    }

    private TimeStampToken verifyInternal(TimeStampToken token) throws Exception {
        TimeStampTokenInfo info = token.getTimeStampInfo();

        TsSignature s = parseNonce(info.getNonce());

        AlgorithmIdentifier digestAlg = digestAlgFinder.find(s.getAlg());
        if (digestAlg == null || !info.getMessageImprintAlgOID().equals(digestAlg.getAlgorithm())) {
            throw new ResultException(Result.REG_NONCE_ALG_MISMATCH,
                    digestAlg == null ? "null" : digestAlg.getAlgorithm(),
                    info.getMessageImprintAlgOID());
        }

        byte[] data = new MessageImprint(digestAlg, info.getMessageImprintDigest()).getEncoded();
        if (!key.verify(data, s.getSignature(), NONE_WITH_RSA)) {
            throw new ResultException(Result.REG_NONCE_SIG_INVALID);
        }

        return token;
    }

    private TsSignature parseNonce(BigInteger nonce) {
        if (nonce == null) {
            throw new ResultException(Result.REG_NO_NONCE);
        }
        try {
            return TsSignature.fromBytes(nonce.toByteArray());
        } catch (Exception e) {
            throw new ResultException(Result.REG_NONCE_NOT_SIG);
        }
    }

}
