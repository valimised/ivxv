package ee.ivxv.common.service.container.bdoc;

import static ee.ivxv.common.util.Util.CHARSET;

import ee.ivxv.common.conf.Conf;
import ee.ivxv.common.service.container.Container;
import ee.ivxv.common.service.container.ContainerReader;
import ee.ivxv.common.service.container.InvalidContainerException;
import ee.ivxv.common.util.Util;

import java.io.*;
import java.nio.file.Path;
import java.security.cert.X509Certificate;
import java.util.Base64;
import java.util.List;
import java.util.Objects;
import java.util.Optional;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.function.Function;
import java.util.regex.Pattern;
import java.util.stream.Collectors;
import java.util.zip.ZipEntry;
import java.util.zip.ZipFile;
import java.util.zip.ZipInputStream;
import java.util.zip.ZipOutputStream;
import javax.xml.crypto.dsig.XMLSignature;
import javax.xml.parsers.DocumentBuilderFactory;

import ee.ivxv.common.zip.Zip;
import ee.ivxv.common.zip.ZipFileCommentException;
import org.apache.xml.security.Init;
import org.apache.xml.security.c14n.Canonicalizer;
import org.digidoc4j.*;
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
    private static final int MAX_BDOC_SIZE = 10 * 1024 * 1024; // 10 MiB
    // Although a valid vote has only [mimetype, META-INF/signatures0.xml, META-INF/manifest.xml, *.ballot] files,
    // we allow overall zip size to be < 500 MiB
    private static final int MAX_FILES_COUNT = 50;

    private final Configuration conf;

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
        return ConfigurationBuilder
                .aConfiguration()
                .withNThreads(nThreads)
                .withX509Certs(conf.getCaCerts())
                .withX509Certs(conf.getOcspCerts())
                .withX509Certs(conf.getTsaCerts())
                .build();
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
        ContainerBuilder cb = ContainerBuilder
                .aContainer()
                .fromExistingFile(path)
                .withConfiguration(conf);
        return read(cb, path, true, MAX_BDOC_SIZE);
    }

    @Override
    public final Container read(String path, int limitSize) throws InvalidContainerException {
        log.debug("readContainer({}) called with size limit {}", path, limitSize);
        ContainerBuilder cb = ContainerBuilder
                .aContainer()
                .fromExistingFile(path)
                .withConfiguration(conf);
        return read(cb, path, true, limitSize);
    }

    @Override
    public final Container read(InputStream input, String ref) throws InvalidContainerException {
        log.debug("readContainer(<InputStream>, {}) called", ref);
        ContainerBuilder cb = ContainerBuilder
                .aContainer()
                .fromStream(input)
                .withConfiguration(conf);
        return read(cb, ref, false, 0);
    }

    private Container read(ContainerBuilder containerBuilder, String ref, boolean isFullRefPath, int limitSize)
            throws InvalidContainerException {
        try {
            org.digidoc4j.Container c = containerBuilder.build();

            if (log.isDebugEnabled()) {
                String files = c
                        .getDataFiles()
                        .stream()
                        .map(DataFile::getName)
                        .collect(Collectors.joining(", "));
                Function<Signature, String> subjNameOrNull =
                        (s) -> Optional
                                .ofNullable(s.getSigningCertificate())
                                .map(X509Cert::getSubjectName)
                                .orElse("null");
                String signers = c
                        .getSignatures()
                        .stream()
                        .map(subjNameOrNull)
                        .collect(Collectors.joining(", "));
                log.debug("container {}, files: {}", ref, files);
                log.debug("container {}, signers: {}", ref, signers);
            }

            validate(c, ref);

            // NB! c.saveAsStream() sets general purpose flag's bit 3, however ZipInputStream cannot determine the
            // uncompressed/compressed sizes ahead from [data descriptor n], although bytes are read into memory by
            // that time
            if (isFullRefPath) {
                // digidoc4j doesn't validate appended bytes at the end of BDOC(zip).
                try (InputStream is = new BufferedInputStream(new FileInputStream(ref))) {
                    Zip.endOfCentralDirectoryRecord(is, MAX_FILES_COUNT, limitSize, limitSize);

                    // If at least 1 byte can be read, then there are appended bytes at the end of BDOC(zip)
                    if (is.read(new byte[1]) != -1) {
                        throw new ZipFileCommentException();
                    }
                }
            }

            return DataConverter.convert(c);
        } catch (ZipFileCommentException | IOException e) { // zip errors
            throw new InvalidContainerException(e.getMessage());
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
     * @param c digidoc container
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
        Objects.requireNonNull(ts);
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

        // Insert Timestamp
        String c14nEl = tsC14nAlg == null ? "" : TS_C14N_EL.replace(TS_C14N_ALG_KEY, tsC14nAlg);
        String tsRespStr = Base64.getEncoder().encodeToString(ts);
        String tsStr = TS_EL;
        tsStr = tsStr.replace(TS_C14N_EL_KEY, c14nEl);
        tsStr = tsStr.replaceAll(TS_RESP_KEY, tsRespStr);

        unsignedProps = unsignedProps.replace(TS_EL_KEY, tsStr);

        // First remove any traces of possible existing 'UnsignedProperties' element
        String cleanSigXml = REMOVE_USP
                .matcher(sigXml)
                .replaceAll("");
        // Then add new 'UnsignedProperties' element
        String result = ADD_USP
                .matcher(cleanSigXml)
                .replaceFirst("$1\n" + unsignedProps);

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

        ByteArrayOutputStream bos = new ByteArrayOutputStream();
        c11r.canonicalizeSubtree(nl.item(0), bos);

        return bos.toByteArray();
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

        private final TSLCertificateSource certSource = new TSLCertificateSourceImpl();
        private int nThreads;

        static ConfigurationBuilder aConfiguration() {
            return new ConfigurationBuilder();
        }

        ConfigurationBuilder withNThreads(int n) {
            nThreads = n;
            return this;
        }

        /**
         * Add x509 certificate to the runtime Trusted Service List (TSL).
         * TSL will be used by digidocj to ensure that {@code cert}
         * is authorized to act as an intermediate or root CA to validate:
         * <ul>
         *     <li>user's certificate</li>
         *     <li>OCSP response</li>
         *     <li>TSA timestamp</li>
         * </ul>
         * NB! Use with caution, if user's certificate issuer is {@code cert}
         * then digidocj will authorize that {@code cert} and no
         * full certificate chain needed
         *
         * @param cert intermediate/root CA (to validate user cert, OCSP response or TSA response)
         */
        void withX509Cert(X509Certificate cert) {
            certSource.addTSLCertificate(cert);
        }

        ConfigurationBuilder withX509Certs(List<X509Certificate> certs) {
            certs.forEach(this::withX509Cert);
            return this;
        }

        Configuration build() {
            Configuration conf = new Configuration();
            conf.setTSL(certSource);
            conf.setThreadExecutor(createExecutorService());
            conf.setAllowASN1UnsafeInteger(true);
            return conf;
        }

        private ExecutorService createExecutorService() {
            if (nThreads <= 0) {
                return Executors.newCachedThreadPool();
            }
            return Executors.newFixedThreadPool(nThreads);
        }
    }

}
