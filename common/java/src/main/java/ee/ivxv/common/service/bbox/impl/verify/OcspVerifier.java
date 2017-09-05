package ee.ivxv.common.service.bbox.impl.verify;

import ee.ivxv.common.crypto.CryptoUtil;
import ee.ivxv.common.crypto.CryptoUtil.CertificateHolder;
import ee.ivxv.common.crypto.CryptoUtil.PublicKeyHolder;
import ee.ivxv.common.service.bbox.Result;
import ee.ivxv.common.service.bbox.impl.ResultException;
import java.security.cert.X509Certificate;
import java.util.ArrayList;
import java.util.List;
import org.bouncycastle.cert.ocsp.BasicOCSPResp;
import org.bouncycastle.cert.ocsp.CertificateID;
import org.bouncycastle.cert.ocsp.CertificateStatus;
import org.bouncycastle.cert.ocsp.OCSPResp;
import org.bouncycastle.cert.ocsp.SingleResp;

public class OcspVerifier {

    private final List<CertificateHolder> cas = new ArrayList<>();
    private final List<PublicKeyHolder> ocsps = new ArrayList<>();

    public OcspVerifier(List<X509Certificate> cas, List<X509Certificate> ocsps) {
        cas.forEach(c -> this.cas.add(CryptoUtil.withCertificate(c)));
        ocsps.forEach(c -> this.ocsps.add(CryptoUtil.withPublicKey(c.getPublicKey())));
    }

    public BasicOCSPResp verify(byte[] response) throws ResultException {
        try {
            return verifyInternal(response);
        } catch (ResultException e) {
            throw e;
        } catch (Exception e) {
            throw new ResultException(Result.OCSP_RESP_INVALID, e);
        }
    }

    private BasicOCSPResp verifyInternal(byte[] response) throws Exception {
        OCSPResp resp = new OCSPResp(response);

        if (resp.getStatus() != OCSPResp.SUCCESSFUL) {
            throw new ResultException(Result.OCSP_STATUS_NOT_SUCCESSFUL, resp.getStatus());
        }

        Object respObject = resp.getResponseObject();
        if (!(respObject instanceof BasicOCSPResp)) {
            throw new ResultException(Result.OCSP_NOT_BASIC);
        }
        BasicOCSPResp basicResp = (BasicOCSPResp) respObject;

        // Check OCSP response signature
        boolean sigOk = ocsps.stream().anyMatch(o -> o.verify(basicResp));
        if (!sigOk) {
            throw new ResultException(Result.OCSP_SIGNATURE_NOT_VALID);
        }

        for (SingleResp sr : basicResp.getResponses()) {
            if (sr.getCertStatus() != CertificateStatus.GOOD) {
                throw new ResultException(Result.OCSP_CERT_STATUS_NOT_GOOD, sr.getCertStatus());
            }
            CertificateID cid = sr.getCertID();

            // Check the issuer on the response is the CA
            if (!cas.stream().anyMatch(i -> i.matchesIssuer(cid))) {
                throw new ResultException(Result.OCSP_ISSUER_UNKNOWN);
            }
        }

        return basicResp;
    }

}
