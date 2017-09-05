package ee.ivxv.audit.util;

import java.io.BufferedReader;
import java.io.IOException;
import java.nio.file.DirectoryStream;
import java.nio.file.FileVisitResult;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.SimpleFileVisitor;
import java.nio.file.attribute.BasicFileAttributes;
import java.util.HashSet;
import java.util.Set;

public class Util {
    /**
     * Verifies that given directory is empty or non-existent. Creates it if the latter is true.
     * 
     * @return boolean indicating weather given path is suitable for further use.
     * @throws IOException
     */
    public static boolean prepareEmptyDir(Path outputPath) throws IOException {
        if (Files.isDirectory(outputPath)) {
            return isEmptyDir(outputPath);
        } else {
            Files.createDirectories(outputPath);
        }
        return true;
    }

    private static boolean isEmptyDir(Path outputPath) throws IOException {
        try (DirectoryStream<Path> dirStream = Files.newDirectoryStream(outputPath)) {
            if (dirStream.iterator().hasNext()) {
                // Output directory is not empty
                return false;
            }
        }
        return true;
    }

    /**
     * Reads given file line by line and stores the data in a set
     *
     * @return set containing lines from input file
     * @throws IOException
     */
    public static Set<String> readFileToSet(Path input) throws IOException {
        Set<String> output = new HashSet<>();
        try (BufferedReader reader = Files.newBufferedReader(input)) {
            String line;
            while ((line = reader.readLine()) != null) {
                output.add(line);
            }
        }
        return output;
    }

    public static class CopyFileVisitor extends SimpleFileVisitor<Path> {
        private final Path targetPath;
        private Path sourcePath = null;

        public CopyFileVisitor(Path targetPath) {
            this.targetPath = targetPath;
        }

        @Override
        public FileVisitResult preVisitDirectory(final Path dir, final BasicFileAttributes attrs)
                throws IOException {
            if (sourcePath == null) {
                sourcePath = dir;
            }
            Files.createDirectories(targetPath.resolve(sourcePath.relativize(dir)));

            return FileVisitResult.CONTINUE;
        }

        @Override
        public FileVisitResult visitFile(final Path file, final BasicFileAttributes attrs)
                throws IOException {
            Files.copy(file, targetPath.resolve(sourcePath.relativize(file)));
            return FileVisitResult.CONTINUE;
        }
    }
}
