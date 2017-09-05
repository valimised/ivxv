package ee.ivxv.common.service.container;

import ee.ivxv.common.M;
import ee.ivxv.common.service.i18n.MessageException;
import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.nio.file.Path;

public interface ContainerReader {

    /**
     * Performs primary essential checks whether the path might be a valid container. Does not throw
     * an exception.
     * 
     * @param path
     * @return
     */
    boolean isContainer(Path path);

    /**
     * If {@code isContainer(path)} returns {@code false}, throws a {@code MessageException}.
     * 
     * @param path
     * @throws MessageException If the given path is not a container.
     */
    default void requireContainer(Path path) throws MessageException {
        if (!isContainer(path)) {
            throw new MessageException(M.e_file_not_container, path);
        }
    }

    /**
     * Reads a signed container from the specified path and returns it if the container is valid.
     * 
     * @param path
     * @return
     * @throws InvalidContainerException if the container is not valid.
     */
    Container read(String path) throws InvalidContainerException;

    /**
     * Reads a signed container from the specified stream and returns it if the container is valid.
     * 
     * @param input
     * @param ref The reference name of the container
     * @return
     * @throws InvalidContainerException if the container is not valid.
     */
    Container read(InputStream input, String ref) throws InvalidContainerException;

    /**
     * Combine a signed container that consists of the provided components.
     * 
     * @param data The initial container data without OCSP and TS data.
     * @param ocsp The OCSP response
     * @param ts The time stamp token data. If {@code null} no timestamp info is added.
     * @param tsC14nAlg The canonicalization algorithm to be used. May be {@code null}.
     * @return
     */
    byte[] combine(byte[] data, byte[] ocsp, byte[] ts, String tsC14nAlg);

    /**
     * Opens a signed container with the specified bytes.
     * 
     * @param bytes The bytes of the container.
     * @param ref The reference name of the container
     * @return
     * @throws InvalidContainerException
     */
    default Container open(byte[] bytes, String ref) throws InvalidContainerException {
        try (InputStream input = new ByteArrayInputStream(bytes)) {
            return read(input, ref);
        } catch (IOException e) {
            /*
             * Does not happen - the JavaDoc of {@code ByteArrayInputStream} says:
             * "Closing a <tt>ByteArrayInputStream</tt> has no effect", but must be handled anyway.
             */
            throw new RuntimeException(e);
        }
    }

    /**
     * @return Returns the file extension for the container type.
     */
    String getFileExtension();

    /**
     * Should be called before program exits, to gracefully shut down all the executor service and
     * threads used by the implementation. The default implementation does nothing.
     */
    default void shutdown() {
        // Nothing
    }

}
