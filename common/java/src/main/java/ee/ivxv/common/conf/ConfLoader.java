package ee.ivxv.common.conf;

import ee.ivxv.common.M;
import ee.ivxv.common.service.container.Container;
import ee.ivxv.common.service.container.ContainerReader;
import ee.ivxv.common.service.container.DataFile;
import ee.ivxv.common.service.container.bdoc.BdocContainerReader;
import ee.ivxv.common.service.i18n.MessageException;
import ee.ivxv.common.util.I18nConsole;
import ee.ivxv.common.util.Util;
import java.io.FileInputStream;
import java.io.FileNotFoundException;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.security.cert.X509Certificate;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Map;
import java.util.Optional;
import java.util.Properties;
import java.util.Set;
import java.util.function.Consumer;
import java.util.stream.Stream;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * ConfLoader creates a new Conf instance with the data loaded from the specified location.
 */
public class ConfLoader {

    static final Logger log = LoggerFactory.getLogger(ConfLoader.class);

    public static final String CONF_FILE_NAME = "ivxv.properties";
    private static final int MAX_REPORTED_UNUSED_FILES = 20;

    static ContainerReader NOT_VALIDATING_CONTAINER_READER = new NotValidatingContainerReader();

    public static final String CA_KEY = "ca";
    public static final String OCSP_KEY = "ocsp";
    public static final String TSA_KEY = "tsa";

    private final I18nConsole console;
    private final ResourceLoader loader;

    // Fields for processing
    private final Set<String> usedFiles = new HashSet<>();

    ConfLoader(I18nConsole console, ResourceLoader loader) {
        this.console = console;
        this.loader = loader;
    }

    public static Conf load(Path path, I18nConsole console) {
        if (!path.toFile().exists()) {
            throw new MessageException(M.e_file_not_found, path);
        }

        ConfLoader loader = new ConfLoader(console, getLoader(path));

        console.println(Msg.m_loading_conf, path);

        return loader.load();
    }

    private static ResourceLoader getLoader(Path path) {
        if (isDir(path)) {
            return new DirLoader(path);
        } else if (NOT_VALIDATING_CONTAINER_READER.isContainer(path)) {
            return new ContainerLoader(path);
        }

        throw new MessageException(Msg.e_unsupported_conf_type, path);
    }

    private static boolean isDir(Path path) {
        return path.toFile().isDirectory();
    }

    Conf load() {
        ConfBuilder cb = new ConfBuilder();

        Properties props = loadProperties();

        loadCerts(props.getProperty(CA_KEY, ""), cb::withCaCerts);
        loadCerts(props.getProperty(OCSP_KEY, ""), cb::withOcspCerts);
        loadCerts(props.getProperty(TSA_KEY, ""), cb::withTsaCerts);

        // Report unknown properties
        props.keySet().stream()
                .filter(k -> !k.equals(CA_KEY) && !k.equals(OCSP_KEY) && !k.equals(TSA_KEY))
                .forEach(k -> console.println(Msg.w_unknown_property, k));

        try (Stream<String> files = loader.getAllFiles()) {
            files.filter(f -> !usedFiles.contains(f)).limit(MAX_REPORTED_UNUSED_FILES)
                    .forEach(this::reportUnusedFile);
        }

        return cb.build();
    }

    private Properties loadProperties() {
        InputStream in = load(CONF_FILE_NAME);
        if (in == null) {
            throw new MessageException(Msg.e_conf_file_not_found, CONF_FILE_NAME);
        }

        Properties props = new Properties();
        try (InputStreamReader reader = new InputStreamReader(in, Util.CHARSET)) {
            props.load(reader);
        } catch (Exception e) {
            throw new MessageException(e, Msg.e_conf_file_open_error, CONF_FILE_NAME,
                    e.getMessage());
        }

        return props;
    }

    private void loadCerts(String paths, Consumer<X509Certificate> consumer) {
        for (String p : paths.split(",")) {
            p = p.trim();
            if (p.isEmpty()) {
                continue;
            }
            log.debug("Loading certificate {}", p);
            try (InputStream in = load(p)) {
                if (in == null) {
                    throw new MessageException(ee.ivxv.common.M.e_cert_not_found, p);
                }
                X509Certificate cert = Util.readCertAsPem(in);
                consumer.accept(cert);
            } catch (MessageException e) {
                throw e;
            } catch (Exception e) {
                throw new MessageException(e, ee.ivxv.common.M.e_cert_read_error, p,
                        e.getMessage());
            }
        }
    }

    private InputStream load(String name) {
        usedFiles.add(name);
        return loader.getResource(name);
    }

    private void reportUnusedFile(String name) {
        log.warn("Unused file in configuration: {}", name);
        console.println(Msg.e_unused_file, name);
    }

    /**
     * ResourceLoader is an interface for loading resources from inside configuration location.
     */
    private interface ResourceLoader {
        InputStream getResource(String name);

        Stream<String> getAllFiles();
    }

    private static class DirLoader implements ResourceLoader {

        private final Path path;

        DirLoader(Path path) {
            this.path = path;
        }

        @Override
        public InputStream getResource(String name) {
            log.debug("Loading resource {} from directory {}", name, path);
            return openFileInputStream(name);
        }

        private InputStream openFileInputStream(String s) {
            try {
                return new FileInputStream(Paths.get(path.toString(), s).toFile());
            } catch (FileNotFoundException e) {
                // To be consistent with ContainerLoader. Handled in load(2).
                return null;
            }
        }

        @Override
        public Stream<String> getAllFiles() {
            try {
                // Stream is closed where it is used - super.load()
                @SuppressWarnings("resource")
                Stream<Path> paths = Files.walk(path);

                return paths.filter(p -> !p.equals(path)).map(p -> path.relativize(p).toString());
            } catch (IOException e) {
                throw new RuntimeException("Error occurred getting files of directory " + path, e);
            }
        }
    } // class DirLoader

    private static class ContainerLoader implements ResourceLoader {

        private final Path path;
        Map<String, DataFile> files = new HashMap<>();

        ContainerLoader(Path path) {
            this.path = path;
            readFiles();
        }

        private void readFiles() {
            Container c = NOT_VALIDATING_CONTAINER_READER.read(path.toString());

            for (DataFile file : c.getFiles()) {
                files.put(file.getName(), file);
            }
        }

        @Override
        public InputStream getResource(String name) {
            log.debug("Loading resource {} from container {}", name, path);
            return Optional.ofNullable(files.get(name)).map(f -> f.getStream()).orElse(null);
        }

        @Override
        public Stream<String> getAllFiles() {
            return files.keySet().stream();
        }
    } // class ContainerLoader

    private static class NotValidatingContainerReader extends BdocContainerReader {

        NotValidatingContainerReader() {
            super(ConfBuilder.aConf().build(), 1);
        }

        @Override
        protected void validate(org.digidoc4j.Container c, String ref) {
            // Don't validate
        }
    }

}
