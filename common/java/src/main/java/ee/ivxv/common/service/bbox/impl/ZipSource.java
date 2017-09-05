package ee.ivxv.common.service.bbox.impl;

import static ee.ivxv.common.util.Util.CHARSET;

import ee.ivxv.common.service.bbox.InvalidBboxException;
import java.io.InputStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.function.BiConsumer;
import java.util.function.Consumer;
import java.util.zip.ZipEntry;
import java.util.zip.ZipFile;
import java.util.zip.ZipInputStream;

class ZipSource implements FileSource {

    private final Path path;

    ZipSource(Path path) {
        this.path = path;
    }

    private static void processZippedStream(Path path,
            BiConsumer<ZipEntry, InputStream> processor) {
        try (ZipInputStream zis = new ZipInputStream(Files.newInputStream(path), CHARSET)) {
            for (ZipEntry ze; (ze = zis.getNextEntry()) != null;) {
                if (ze.isDirectory()) {
                    continue;
                }
                processor.accept(ze, zis);
            }
        } catch (Exception e) {
            throw new RuntimeException(new InvalidBboxException(path, e));
        }
    }

    private static void processZipFile(Path path, Consumer<ZipEntry> processor) {
        try (ZipFile zip = openZipFile(path)) {
            zip.stream().forEach(ze -> {
                if (ze.isDirectory()) {
                    return;
                }
                processor.accept(ze);
            });
        } catch (Exception e) {
            throw new InvalidBboxException(path, e);
        }
    }

    private static ZipFile openZipFile(Path path) {
        try {
            return new ZipFile(path.toFile());
        } catch (Exception e) {
            throw new InvalidBboxException(path, e);
        }
    }

    @Override
    public void processFiles(BiConsumer<String, InputStream> processor) {
        processZippedStream(path, (ze, in) -> {
            processor.accept(ze.getName(), in);
        });
    }

    @Override
    public void list(Consumer<String> processor) {
        processZipFile(path, ze -> {
            processor.accept(ze.getName());
        });
    }

}
