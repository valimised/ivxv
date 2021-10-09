package ee.ivxv.common.service.container.bdoc;

import static ee.ivxv.common.util.Util.CHARSET;

import ee.ivxv.common.conf.Conf;
import ee.ivxv.common.service.container.Container;
import ee.ivxv.common.service.container.ContainerReader;
import ee.ivxv.common.service.container.InvalidContainerException;
import ee.ivxv.common.util.Util;
import eu.europa.esig.dss.tsl.Condition;
import eu.europa.esig.dss.tsl.ServiceInfo;
import eu.europa.esig.dss.tsl.ServiceInfoStatus;
import eu.europa.esig.dss.util.TimeDependentValues;
import eu.europa.esig.dss.x509.CertificateToken;
import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.nio.file.Path;
import java.security.cert.X509Certificate;
import java.util.Arrays;
import java.util.Base64;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.regex.Pattern;
import java.util.stream.Collectors;
import java.util.zip.ZipEntry;
import java.util.zip.ZipFile;
import java.util.zip.ZipInputStream;
import java.util.zip.ZipOutputStream;
import javax.xml.crypto.dsig.XMLSignature;
import javax.xml.parsers.DocumentBuilderFactory;
import org.apache.xml.security.Init;
import org.apache.xml.security.c14n.Canonicalizer;
import org.digidoc4j.Configuration;
import org.digidoc4j.ContainerBuilder;
import org.digidoc4j.ContainerValidationResult;
import org.digidoc4j.TSLCertificateSource;
import org.digidoc4j.impl.asic.tsl.TSLCertificateSourceImpl;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.w3c.dom.Document;
import org.w3c.dom.NodeList;

public class BdocContainerReader implements ContainerReader {

    private static final Logger log = LoggerFactory.getLogger(BdocContainerReader.class);

    public static final String FILE_EXTENSION = "bdoc";
    private static final String SIG_FILE_PREFIX = "META-INF/signatures";
    private static final String SIG_VALUE_EL = "SignatureValue";

    private Configuration conf;

    static {
        Init.init();
    }

    public BdocContainerReader(Conf conf, int nThreads) {
        this(createConfiguration(conf, nThreads));
        log.info("BdocContainerReader instantiated with thread count {}", nThreads);
    }

    private BdocContainerReader(Configuration conf) {
        this.conf = conf;
    }

    private static Configuration createConfiguration(Conf conf, int nThreads) {
        ConfigurationBuilder cb = ConfigurationBuilder.aConfiguration().withNThreads(nThreads);

        conf.getCaCerts().forEach(cb::withCaCert);
        conf.getOcspCerts().forEach(cb::withOcspCert);
        conf.getTsaCerts().forEach(cb::withTsaCert);

        return cb.build();
    }

    @Override
    public boolean isContainer(Path path) {
        try (ZipFile zip = new ZipFile(path.toFile())) {
            return true;
        } catch (Exception e) {
            return false;
        }
    }

    @Override
    public final Container read(String path) throws InvalidContainerException {
        log.debug("readContainer({}) called", path);
        return read(ContainerBuilder.aContainer().fromExistingFile(path).withConfiguration(conf),
                path);
    }

    @Override
    public final Container read(InputStream input, String ref) throws InvalidContainerException {
        log.debug("readContainer(<InputStream>, {}) called", ref);
        return read(ContainerBuilder.aContainer().fromStream(input).withConfiguration(conf), ref);
    }

    private Container read(ContainerBuilder containerBuilder, String ref)
            throws InvalidContainerException {
        try {
            org.digidoc4j.Container c = containerBuilder.build();

            if (log.isDebugEnabled()) {
                String files = c.getDataFiles().stream().map(f -> f.getName())
                        .collect(Collectors.joining(", "));
                String signers = c.getSignatures().stream()
                        .map(s -> Optional.ofNullable(s.getSigningCertificate())
                                .map(sc -> sc.getSubjectName()).orElse("null"))
                        .collect(Collectors.joining(", "));
                log.debug("container {}, files: {}", ref, files);
                log.debug("container {}, signers: {}", ref, signers);
            }

            validate(c, ref);

            return DataConverter.convert(c);
        } catch (InvalidContainerException e) {
            throw e;
        } catch (Throwable e) {
            throw new InvalidContainerException(ref, e);
        }
    }

    /**
     * Validates the specified container, i.e. throws a runtime exception if the container is not
     * valid. Subclasses can override it to bypass validation, which is generally not recommended.
     * 
     * @param c
     * @param ref The reference name of the container
     * @throws InvalidContainerException if the specified BDOC container is invalid.
     */
    protected void validate(org.digidoc4j.Container c, String ref)
            throws InvalidContainerException {
        ContainerValidationResult vr = c.validate();

        if (!vr.isValid()) {
            log.warn("BDOC container does not validate! Validation report: {}", vr.getReport());
            throw new InvalidContainerException(ref);
        }
    }

    @Override
    public byte[] combine(byte[] bdoc, byte[] ocsp, byte[] ts, String tsC14nAlg) {
        ByteArrayOutputStream out = new ByteArrayOutputStream(5000);
        byte[] buffer = new byte[1024];

        log.debug("combine(): combining container data");

        try (ZipInputStream zis = new ZipInputStream(new ByteArrayInputStream(bdoc), CHARSET);
                ZipOutputStream zos = new ZipOutputStream(out, CHARSET)) {
            for (ZipEntry ze; (ze = zis.getNextEntry()) != null;) {
                zos.putNextEntry(new ZipEntry(ze.getName()));

                // Modify signature file, copy others
                if (ze.getName().startsWith(SIG_FILE_PREFIX)) {
                    String xml = Util.toString(Util.toBytes(zis, buffer));

                    zos.write(addDataToSignature(xml, ocsp, ts, tsC14nAlg));
                } else {
                    for (int len; (len = zis.read(buffer)) > 0;) {
                        zos.write(buffer, 0, len);
                    }
                }
            }
            zos.closeEntry();
        } catch (IOException e) {
            throw new RuntimeException(e);
        }

        return out.toByteArray();
    }

    private byte[] addDataToSignature(String sigXml, byte[] ocsp, byte[] ts, String tsC14nAlg) {
        String unsignedProps = UNSIGNED_PROPS_EL;

        // Insert OCSP response
        String ocspRespStr = Base64.getEncoder().encodeToString(ocsp);
        unsignedProps = unsignedProps.replace(OCSP_RESP_KEY, ocspRespStr);

        // Insert Timestamp, if provided
        String tsStr = "";
        if (ts != null) {
            String c14nEl = tsC14nAlg == null ? "" : TS_C14N_EL.replace(TS_C14N_ALG_KEY, tsC14nAlg);
            String tsRespStr = Base64.getEncoder().encodeToString(ts);

            tsStr = TS_EL;
            tsStr = tsStr.replace(TS_C14N_EL_KEY, c14nEl);
            tsStr = tsStr.replaceAll(TS_RESP_KEY, tsRespStr);
        }
        unsignedProps = unsignedProps.replace(TS_EL_KEY, tsStr);

        // First remove any traces of possible existing 'UnsignedProperties' element
        String cleanSigXml = REMOVE_USP.matcher(sigXml).replaceAll("");
        // Then add new 'UnsignedProperties' element
        String result = ADD_USP.matcher(cleanSigXml).replaceFirst("$1\n" + unsignedProps);

        return Util.toBytes(result);
    }

    @Override
    public byte[] getTimestampData(byte[] bdoc, String c14nAlg) {
        try (ZipInputStream zis = new ZipInputStream(new ByteArrayInputStream(bdoc), CHARSET)) {
            for (ZipEntry ze; (ze = zis.getNextEntry()) != null;) {
                if (ze.getName().startsWith(SIG_FILE_PREFIX)) {
                    return calculateTimestampData(Util.toBytes(zis), c14nAlg);
                }
            }
        } catch (Exception e) {
            throw new RuntimeException(e);
        }

        throw new RuntimeException("Signature file with prefix" + SIG_FILE_PREFIX + " not found");
    }

    private byte[] calculateTimestampData(byte[] sigXml, String c14nAlg) throws Exception {
        DocumentBuilderFactory dbf = DocumentBuilderFactory.newInstance();
        dbf.setNamespaceAware(true);
        Document doc = dbf.newDocumentBuilder().parse(new ByteArrayInputStream(sigXml));
        NodeList nl = doc.getElementsByTagNameNS(XMLSignature.XMLNS, SIG_VALUE_EL);

        Canonicalizer c11r = Canonicalizer.getInstance(c14nAlg);

        return c11r.canonicalizeSubtree(nl.item(0));
    }

    @Override
    public String getFileExtension() {
        return FILE_EXTENSION;
    }

    @Override
    public void shutdown() {
        if (conf.getThreadExecutor() != null) {
            conf.getThreadExecutor().shutdown();
        }
    }

    private static final Pattern REMOVE_USP = Pattern.compile(
            "<\\s*\\w+:UnsignedProperties[^>]*>.*</\\s*\\w+:UnsignedProperties[^>]*>",
            Pattern.DOTALL);
    private static final Pattern ADD_USP = Pattern.compile("(</\\s*\\w+:SignedProperties[^>]*>)");

    private static final String OCSP_RESP_KEY = "OCSP_RESPONSE";
    private static final String TS_EL_KEY = "TS_ELEMENT";
    private static final String TS_RESP_KEY = "TS_RESPONSE";
    private static final String TS_C14N_EL_KEY = "TS_C14N_EL";
    private static final String TS_C14N_ALG_KEY = "TS_C14N_ALG";

    // (Re-)defining namespaces for local use to be sure
    private static final String UNSIGNED_PROPS_EL = "" //
            + "        <xades:UnsignedProperties\n" //
            + "            xmlns:xades=\"http://uri.etsi.org/01903/v1.3.2#\"\n"
            + "            xmlns:ds=\"http://www.w3.org/2000/09/xmldsig#\">\n"
            + "          <xades:UnsignedSignatureProperties>\n" //
            + TS_EL_KEY //
            + "            <xades:RevocationValues>\n" //
            + "              <xades:OCSPValues>\n" //
            + "                <xades:EncapsulatedOCSPValue Id=\"N0\">\n" //
            + "                  " + OCSP_RESP_KEY + "\n" //
            + "                </xades:EncapsulatedOCSPValue>\n" //
            + "              </xades:OCSPValues>\n" //
            + "            </xades:RevocationValues>\n"
            + "          </xades:UnsignedSignatureProperties>\n"
            + "        </xades:UnsignedProperties>";

    private static final String TS_EL = "" //
            + "            <xades:SignatureTimeStamp Id=\"S0-T0\">\n" //
            + TS_C14N_EL_KEY //
            + "              <xades:EncapsulatedTimeStamp>\n" //
            + "                " + TS_RESP_KEY + "\n" //
            + "              </xades:EncapsulatedTimeStamp>\n" //
            + "            </xades:SignatureTimeStamp>\n";

    private static final String TS_C14N_EL = "" //
            + "              <ds:CanonicalizationMethod Algorithm=\"" + TS_C14N_ALG_KEY + "\" />\n";

    /**
     * ConfigurationBuilder is helper class for composing configuration.
     */
    static class ConfigurationBuilder {

        private static final String OCSP_SERVICE_INFO_TYPE =
                "http://uri.etsi.org/TrstSvc/Svctype/Certstatus/OCSP/QC";
        private static final String TSA_SERVICE_INFO_TYPE =
                "http://uri.etsi.org/TrstSvc/Svctype/TSA";
        private static final String OCSP_SERVICE_INFO_STATUS =
                "http://uri.etsi.org/TrstSvc/TrustedList/Svcstatus/undersupervision";
        private static final String TSA_SERVICE_INFO_STATUS =
                "http://uri.etsi.org/TrstSvc/TrustedList/Svcstatus/undersupervision";

        private TSLCertificateSource certSource = new TSLCertificateSourceImpl();
        private int nThreads;

        static ConfigurationBuilder aConfiguration() {
            return new ConfigurationBuilder();
        }

        ConfigurationBuilder withNThreads(int n) {
            nThreads = n;
            return this;
        }

        ConfigurationBuilder withCaCert(X509Certificate cert) {
            certSource.addTSLCertificate(cert);
            return this;
        }

        ConfigurationBuilder withOcspCert(X509Certificate cert) {
            certSource.addCertificate(new CertificateToken(cert), createOcspServiceInfo(cert));
            return this;
        }

        ConfigurationBuilder withTsaCert(X509Certificate cert) {
            certSource.addCertificate(new CertificateToken(cert), createTsaServiceInfo(cert));
            return this;
        }

        Configuration build() {
            Configuration conf = new Configuration();
            conf.setTSL(certSource);
            conf.setThreadExecutor(createExecutorService());
            conf.setAllowASN1UnsafeInteger(true);
            return conf;
        }

        private ServiceInfo createOcspServiceInfo(X509Certificate cert) {
            ServiceInfo serviceInfo = new ServiceInfo();

            Map<String, List<Condition>> qualifiers = new HashMap<>(); // Must not be null!
            ServiceInfoStatus status =
                    new ServiceInfoStatus(OCSP_SERVICE_INFO_TYPE, OCSP_SERVICE_INFO_STATUS,
                            qualifiers, null, null, null, cert.getNotBefore(), cert.getNotAfter());

            serviceInfo.setStatus(new TimeDependentValues<>(Arrays.asList(status)));
            return serviceInfo;
        }

        private ServiceInfo createTsaServiceInfo(X509Certificate cert) {
            ServiceInfo serviceInfo = new ServiceInfo();

            Map<String, List<Condition>> qualifiers = new HashMap<>(); // Must not be null!
            ServiceInfoStatus status =
                    new ServiceInfoStatus(TSA_SERVICE_INFO_TYPE, TSA_SERVICE_INFO_STATUS,
                            qualifiers, null, null, null, cert.getNotBefore(), cert.getNotAfter());

            serviceInfo.setStatus(new TimeDependentValues<>(Arrays.asList(status)));
            return serviceInfo;
        }

        private ExecutorService createExecutorService() {
            if (nThreads <= 0) {
                return Executors.newCachedThreadPool();
            }
            return Executors.newFixedThreadPool(nThreads);
        }
    }

}
