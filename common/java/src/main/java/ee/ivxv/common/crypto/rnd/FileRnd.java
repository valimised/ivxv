package ee.ivxv.common.crypto.rnd;

import java.io.IOException;
import java.nio.ByteBuffer;
import java.nio.channels.FileChannel;
import java.nio.file.Path;
import java.nio.file.StandardOpenOption;

/**
 * Entropy source which reads directly from a file or a stream. Direct use in application is
 * strongly discouraged, as the number of read bytes may be less than requested. Also, the input is
 * not conditioned in any way.
 * <p>
 * It is intended to be used together with {@link CombineRnd}, adding this as a source to it.
 */
public class FileRnd implements Rnd {
    protected static int BUF_SIZE = 1024;
    private final FileChannel fileChannel;
    private final ByteBuffer fileBuf = ByteBuffer.allocate(BUF_SIZE);
    private final boolean finite;

    /**
     * Initialize the instance using a path.
     * 
     * @param rndSource Path to file or stream.
     * @param finite Boolean indicating whether the path target should be considered as a finite
     *        file or a infinite stream. False positive (i.e. indicating that a infinite source is
     *        finite) leads to infinite loop during initialization of {@link CombineRnd}. If not
     *        sure, set {@code false}.
     * @throws IOException On exception while opening the source.
     */
    public FileRnd(Path rndSource, boolean finite) throws IOException {
        fileChannel = FileChannel.open(rndSource, StandardOpenOption.READ);
        fileBuf.flip();
        this.finite = finite;
    }

    /**
     * Read the source and write the result to output buffer.
     * 
     * @param buf The output buffer.
     * @param offset The offset to start storing the read bytes.
     * @param len The number of bytes to read.
     * @return The actual number of bytes read.
     * @throws IOException On exception while reading the source.
     */
    @Override
    public int read(byte[] buf, int offset, int len) throws IOException {
        int read = 0;
        int to_read, refilled;
        while (read < len) {
            refilled = refill();
            if (refilled == -1) {
                // we have reached end of file. fileBuf is empty
                return read;
            }
            to_read = Math.min(len - read, fileBuf.remaining());
            fileBuf.get(buf, offset + read, to_read);
            read += to_read;
            if (refilled < fileBuf.capacity()) {
                // we have reached end of available stream. do not go further for now.
                return read;
            }
        }
        return read;
    }

    /**
     * Read the source and write exactly the number of bytes requested into output buffer.
     * 
     * @param buf The output buffer.
     * @param offset The offset to start storing the read bytes.
     * @param len The number of bytes to read.
     * @return The number of bytes requested.
     * @throws IOException On exception while reading the source.
     */
    @Override
    public int mustRead(byte[] buf, int offset, int len) throws IOException {
        int read = 0;
        int to_read, refilled;
        while (read < len) {
            refilled = refill();
            if (refilled == -1) {
                // we have reached end of file. fileBuf is empty
                throw new IOException("End of file");
            }
            to_read = Math.min(len - read, fileBuf.remaining());
            fileBuf.get(buf, offset + read, to_read);
            read += to_read;
        }
        return read;
    }

    /**
     * @return The finite argument given during initialization.
     */
    @Override
    public boolean isFinite() {
        return finite;
    }

    private int refill() throws IOException {
        int read = fileBuf.remaining();
        if (!fileBuf.hasRemaining()) {
            fileBuf.clear();
            read = fileChannel.read(fileBuf);
            fileBuf.flip();
        }
        return read;
    }

    /**
     * Try to close the source file.
     */
    @Override
    public void close() {
        try {
            fileChannel.close();
        } catch (IOException e) {
            // unhandleable exception
        }
    }
}
