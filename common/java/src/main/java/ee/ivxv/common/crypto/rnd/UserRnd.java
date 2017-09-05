package ee.ivxv.common.crypto.rnd;

import ee.ivxv.common.util.Util;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.lang.ProcessBuilder;
import java.net.InetAddress;
import java.net.Socket;
import java.nio.file.Path;

/**
 * Entropy source which calls external program to obtain entropy. Communication with external
 * program is done using sockets by sending the amount of bytes requested (encoded as C int). The
 * external program is supposed to reply with the exact amount of requested bytes.
 *
 */
public class UserRnd implements Rnd {
    private Process process;
    private final boolean cont;
    private final int port;
    private final Path externalProg;

    /**
     * Initialize the external entropy random source. The entropy will be provided by the program
     * defined by {@link externalProg}. The communication is performed over local network, where the
     * program should listen on {@link port}. The parameter {@link cont} defines, if the program is
     * started once or started for read.
     * 
     * @param externalProg Location of external program to ask for entropy.
     * @param cont Define if the program should not be killed after every entropy query.
     * @param port The port to communicate with the external program.
     */
    public UserRnd(Path externalProg, boolean cont, int port) throws IOException {
        if (cont) {
            this.process = executeExternal(externalProg, port);
        }
        this.cont = cont;
        this.port = port;
        this.externalProg = externalProg;
    }

    private static class QueryException extends IOException {

        public QueryException(String string) {
            super(string);
        }
    }


    private Process executeExternal(Path location, int port) throws IOException {
        Process p = new ProcessBuilder(location.toString(), Integer.toString(port)).start();
        return p;
    }

    private void forceShutdown() {
        process.destroy();
        process = null;
    }

    private byte[] queryExternal(Socket s, int amount, boolean must) throws IOException {
        OutputStream out = s.getOutputStream();
        InputStream in = s.getInputStream();
        byte[] sendValue = Util.toBytes(amount);
        byte[] ret;
        int available, read;
        while (true) {
            out.write(sendValue);
            available = in.read();
            if (available == 0) {
                if (must) {
                    try {
                        Thread.sleep(100);
                    } catch (InterruptedException e) {
                        shutdown();
                        s.close();
                        throw new QueryException("Unexpected interrupt");
                    }
                    continue;
                } else {
                    ret = new byte[0];
                    break;
                }
            } else if (available == 0xff) {
                ret = new byte[amount];
                read = in.read(ret);
                if (read != amount) {
                    shutdown();
                    s.close();
                    throw new QueryException("Short response");
                }
                break;
            } else {
                shutdown();
                s.close();
                throw new QueryException("Corrupted response");
            }
        }
        shutdown();
        s.close();
        return ret;
    }

    private byte[] queryProg(int amount, boolean must) throws IOException {
        Socket s = null;
        byte[] ret = null;
        for (int i = 0; i < 5; i++) {
            start();
            try {
                s = new Socket(InetAddress.getLocalHost(), port);
                ret = queryExternal(s, amount, must);
                s.close();
            } catch (QueryException e) {
                throw new IOException(e.getMessage());
            } catch (IOException e) {
                if (i == 4) {
                    throw e;
                }
            } finally {
                shutdown();
            }
            if (ret != null) {
                return ret;
            }
        }
        return null;
    }

    private void start() throws IOException {
        if (!cont || (process != null && !process.isAlive())) {
            process = executeExternal(externalProg, port);
        }
    }

    private void shutdown() {
        if (!cont) {
            forceShutdown();
        }
    }

    /**
     * Ask the external program for entropy and write the result to output buffer.
     * 
     * @param buf The output buffer.
     * @param offset The offset to start storing the read bytes.
     * @param len The number of bytes to ask.
     * @return The actual number of bytes read.
     * @throws IOException On exception while communicating with the source.
     */
    @Override
    public int read(byte[] buf, int offset, int len) throws IOException {
        byte[] ent = queryProg(len, false);
        int to_copy = ent.length < len ? ent.length : len;
        System.arraycopy(ent, 0, buf, offset, to_copy);
        return to_copy;
    }

    /**
     * Ask the external program for entropy and write the result to output buffer.
     * 
     * @param buf The output buffer.
     * @param offset The offset to start storing the read bytes.
     * @param len The number of bytes to ask.
     * @return The number of bytes requested.
     * @throws IOException On exception while communicating with the source.
     */
    @Override
    public int mustRead(byte[] buf, int offset, int len) throws IOException {
        byte[] ent = queryProg(len, true);
        System.arraycopy(ent, 0, buf, offset, len);
        return len;
    }

    /**
     * @return false
     */
    @Override
    public boolean isFinite() {
        return false;
    }


    /**
     * Shut down the external entropy source if it is running.
     */
    @Override
    public void close() {
        if (cont) {
            forceShutdown();
        }
    }
}
