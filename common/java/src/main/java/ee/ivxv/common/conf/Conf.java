package ee.ivxv.common.conf;

import java.security.cert.X509Certificate;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

/**
 * Conf is common runtime configuration of any IVXV application.
 */
public class Conf {

    private final List<X509Certificate> caCerts;
    private final List<X509Certificate> ocspCerts;
    private final List<X509Certificate> tsaCerts;

    public Conf() {
        this(new ArrayList<>(), new ArrayList<>(), new ArrayList<>());
    }

    public Conf(List<X509Certificate> caCerts, List<X509Certificate> ocspCerts,
            List<X509Certificate> tsaCerts) {
        this.caCerts = Collections.unmodifiableList(caCerts);
        this.ocspCerts = Collections.unmodifiableList(ocspCerts);
        this.tsaCerts = Collections.unmodifiableList(tsaCerts);
    }

    public List<X509Certificate> getCaCerts() {
        return caCerts;
    }

    public List<X509Certificate> getOcspCerts() {
        return ocspCerts;
    }

    public List<X509Certificate> getTsaCerts() {
        return tsaCerts;
    }

}
