package ee.ivxv.common.crypto;

import ee.ivxv.common.util.Util;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.security.InvalidKeyException;
import java.security.KeyFactory;
import java.security.NoSuchAlgorithmException;
import java.security.PublicKey;
import java.security.Signature;
import java.security.SignatureException;
import java.security.cert.CertificateEncodingException;
import java.security.cert.X509Certificate;
import java.security.spec.X509EncodedKeySpec;
import org.bouncycastle.asn1.ASN1ObjectIdentifier;
import org.bouncycastle.cert.X509CertificateHolder;
import org.bouncycastle.cert.ocsp.BasicOCSPResp;
import org.bouncycastle.cert.ocsp.CertificateID;
import org.bouncycastle.cert.ocsp.OCSPException;
import org.bouncycastle.operator.AlgorithmNameFinder;
import org.bouncycastle.operator.DefaultAlgorithmNameFinder;
import org.bouncycastle.operator.bc.BcDigestCalculatorProvider;

public class CryptoUtil {

    static final AlgorithmNameFinder ALGORITHM_FINDER = new DefaultAlgorithmNameFinder();

    public static PublicKeyHolder loadPublicKey(Path path, ASN1ObjectIdentifier alg)
            throws Exception {
        byte[] keyBytes = Util.decodePublicKey(Util.toString(Files.readAllBytes(path)));
        return withPublicKey(loadPublicKey(keyBytes, ALGORITHM_FINDER.getAlgorithmName(alg)));
    }

    public static PublicKey loadPublicKey(byte[] bytes, String algorithm) throws Exception {
        KeyFactory kf = KeyFactory.getInstance(algorithm);
        X509EncodedKeySpec keySpec = new X509EncodedKeySpec(bytes);
        PublicKey key = kf.generatePublic(keySpec);

        return key;
    }

    public static PublicKeyHolder withPublicKey(PublicKey key) {
        return new PublicKeyHolder(key);
    }

    public static CertificateHolder withCertificate(X509Certificate cert) {
        return new CertificateHolder(cert);
    }

    /**
     * PublicKeyHolder is a helper class for common operations with public key.
     */
    public static class PublicKeyHolder {

        private final PublicKey key;

        PublicKeyHolder(PublicKey key) {
            this.key = key;
        }

        public boolean verify(BasicOCSPResp res) {
            return verify(res.getTBSResponseData(), res.getSignature(), res.getSignatureAlgOID());
        }

        public boolean verify(byte[] data, byte[] signature, ASN1ObjectIdentifier algId) {
            return verify(data, signature, ALGORITHM_FINDER.getAlgorithmName(algId));
        }

        public boolean verify(byte[] data, byte[] signature, String algorithm) {
            try {
                Signature dsa = createSignature(algorithm);

                dsa.initVerify(key);
                dsa.update(data);

                return dsa.verify(signature);
            } catch (InvalidKeyException | NoSuchAlgorithmException | SignatureException e) {
                throw new RuntimeException(e.getMessage(), e);
            }
        }

        Signature createSignature(String algorithm) throws NoSuchAlgorithmException {
            return Signature.getInstance(algorithm);
        }

    }

    /**
     * CertificateHolder is a helper class for common operations with X509 certificates.
     */
    public static class CertificateHolder {

        private static final BcDigestCalculatorProvider DCP = new BcDigestCalculatorProvider();

        private final X509CertificateHolder cert;

        CertificateHolder(X509Certificate cert) {
            try {
                this.cert = new X509CertificateHolder(cert.getEncoded());
            } catch (CertificateEncodingException | IOException e) {
                throw new RuntimeException(e.getMessage(), e);
            }
        }

        public boolean matchesIssuer(CertificateID certId) {
            try {
                return certId.matchesIssuer(cert, DCP);
            } catch (OCSPException e) {
                throw new RuntimeException(e.getMessage(), e);
            }
        }

    }

}
