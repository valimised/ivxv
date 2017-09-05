package ee.ivxv.common.conf;

import java.security.cert.X509Certificate;
import java.util.ArrayList;
import java.util.List;

public class ConfBuilder {

    private final List<X509Certificate> caCerts = new ArrayList<>();
    private final List<X509Certificate> ocspCerts = new ArrayList<>();
    private final List<X509Certificate> tsaCerts = new ArrayList<>();

    public static ConfBuilder aConf() {
        return new ConfBuilder();
    }

    public ConfBuilder withCaCerts(X509Certificate... certs) {
        return addCerts(caCerts, certs);
    }

    public ConfBuilder withOcspCerts(X509Certificate... certs) {
        return addCerts(ocspCerts, certs);
    }

    public ConfBuilder withTsaCerts(X509Certificate... certs) {
        return addCerts(tsaCerts, certs);
    }

    public Conf build() {
        return new Conf(caCerts, ocspCerts, tsaCerts);
    }

    private ConfBuilder addCerts(List<X509Certificate> dest, X509Certificate... certs) {
        for (X509Certificate cert : certs) {
            dest.add(cert);
        }
        return this;
    }

}
