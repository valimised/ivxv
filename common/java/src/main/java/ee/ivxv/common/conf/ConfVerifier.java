package ee.ivxv.common.conf;

import ee.ivxv.common.cli.AppContext;
import ee.ivxv.common.service.container.Container;
import ee.ivxv.common.service.i18n.MessageException;
import ee.ivxv.common.util.ContainerHelper;
import ee.ivxv.common.util.I18nConsole;
import java.nio.file.Path;
import java.security.cert.X509Certificate;
import java.util.List;

/**
 * ConfVerifier verifies the configuration integrity and correctness.
 */
public class ConfVerifier {

    private static final String OCSP_EXT = "1.3.6.1.5.5.7.3.9";
    private static final String TSA_EXT = "1.3.6.1.5.5.7.3.8";

    public static void verify(Conf conf) {
        conf.getCaCerts().forEach(c -> requireCa(c));
        conf.getOcspCerts().forEach(c -> requireExt(c, OCSP_EXT, Msg.e_ocsp_not_ocsp_cert));
        conf.getTsaCerts().forEach(c -> requireExt(c, TSA_EXT, Msg.e_tsp_not_tsp_cert));
    }

    private static void requireCa(X509Certificate cert) {
        if (!isCa(cert)) {
            throw new MessageException(Msg.e_ca_not_ca_cert, cert);
        }
    }

    private static boolean isCa(X509Certificate cert) {
        return cert.getBasicConstraints() >= 0;
    }

    private static void requireExt(X509Certificate cert, String ext, Enum<?> key) {
        if (!hasExt(cert, ext)) {
            throw new MessageException(key, cert, ext);
        }
    }

    private static boolean hasExt(X509Certificate cert, String ext) {
        try {
            List<String> exts = cert.getExtendedKeyUsage();
            return exts != null && exts.contains(ext);
        } catch (Exception e) {
            throw new RuntimeException(e);
        }
    }

    public static void verifySignature(AppContext<?> ctx, Path path) {
        ctx.container.requireContainer(path);

        I18nConsole console = new I18nConsole(ctx.i.console, ctx.i.i18n);
        Container c = ctx.container.read(path.toString());
        ContainerHelper ch = new ContainerHelper(console, c);

        ch.reportSignatures(Msg.m_conf_arg_for_cont);
    }

}
