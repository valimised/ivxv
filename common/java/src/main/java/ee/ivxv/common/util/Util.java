package ee.ivxv.common.util;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.net.URL;
import java.nio.charset.Charset;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.security.cert.CertificateException;
import java.security.cert.CertificateFactory;
import java.security.cert.X509Certificate;
import java.util.Arrays;
import java.util.Base64;

public class Util {

    public static final Charset CHARSET = StandardCharsets.UTF_8;
    public static final String X509 = "X.509";
    public static final String BEGIN_CERTIFICATE = "-----BEGIN CERTIFICATE-----";
    public static final String END_CERTIFICATE = "-----END CERTIFICATE-----";
    public static final String BEGIN_PUB_KEY = "-----BEGIN PUBLIC KEY-----";
    public static final String END_PUB_KEY = "-----END PUBLIC KEY-----";
    public static final String BEGIN_PRIVATE_KEY = "-----BEGIN PRIVATE KEY-----";
    public static final String END_PRIVATE_KEY = "-----END PRIVATE KEY-----";
    public static final String EOT = Character.toString((char) 0x04); // EOT character
    public static final String UNIT_SEPARATOR = Character.toString((char) 0x1F); // Unit separator
                                                                                 // character

    /**
     * Return byte array with UTF-8 encoded string.
     * 
     * @param string
     * @return Returns UTF-8 encoded byte array.
     */
    public static byte[] toBytes(String string) {
        return string.getBytes(CHARSET);
    }

    /**
     * Decode byte array as UTF-8 encoded string.
     * 
     * @param bytes
     * @return Returns a string decoded from the bytes using UTF-8 encoding.
     */
    public static String toString(byte[] bytes) {
        return new String(bytes, CHARSET);
    }

    /**
     * Return the InputStream as a byte array.
     * 
     * @param stream
     * @return
     */
    public static byte[] toBytes(InputStream stream) {
        return toBytes(stream, new byte[1024]);
    }

    /**
     * Return the InputStream as a byte array using the buffer for storing temporary values.
     * 
     * @param stream
     * @param buffer
     * @return
     */
    public static byte[] toBytes(InputStream stream, byte[] buffer) {
        ByteArrayOutputStream bytes = new ByteArrayOutputStream();

        try {
            for (int len; (len = stream.read(buffer)) != -1;) {
                bytes.write(buffer, 0, len);
            }
        } catch (IOException e) {
            throw new RuntimeException(e);
        }

        return bytes.toByteArray();
    }

    /**
     * Return the InputStream as an UTF-8 encoded string using the buffer for storing temporary
     * values.
     * 
     * @param stream
     * @param buffer
     * @return
     */
    public static String toString(InputStream stream, byte[] buffer) {
        return new String(toBytes(stream, buffer), CHARSET);
    }

    /**
     * Return the integer encoded as a big-endian byte array.
     * 
     * @param i
     * @return
     */
    public static byte[] toBytes(int i) {
        byte[] ret = new byte[4];
        ret[0] = (byte) ((i >>> 24) & 0xff);
        ret[1] = (byte) ((i >>> 16) & 0xff);
        ret[2] = (byte) ((i >>> 8) & 0xff);
        ret[3] = (byte) ((i >>> 0) & 0xff);
        return ret;
    }

    /**
     * Return the integer encoded as 4-byte big-endian byte array.
     * 
     * @param b
     * @return
     */
    public static int toInt(byte[] b) {
        // we assume 4 bytes. if b is other length, then the result may be
        // undefined
        int k = 0;
        for (int i = 0; i < b.length && i < 4; i++) {
            k |= (b[i] << (24 - (i * 8)));
        }
        return k;
    }

    /**
     * Create a file at a given path.
     * 
     * @param path
     * @throws IOException
     */
    public static void createFile(Path path) throws IOException {
        if (path.getParent() != null) {
            Files.createDirectories(path.getParent());
        }
        Files.createFile(path);
    }

    /**
     * Concatenate the byte arrays into a single byte array.
     * 
     * @param first
     * @param rest
     * @return
     */
    public static byte[] concatAll(byte[] first, byte[]... rest) {
        int totalLength = first.length;
        for (byte[] array : rest) {
            totalLength += array.length;
        }
        byte[] result = Arrays.copyOf(first, totalLength);
        int offset = first.length;
        for (byte[] array : rest) {
            System.arraycopy(array, 0, result, offset, array.length);
            offset += array.length;
        }
        return result;
    }

    /**
     * MIME-encode the byte array as a certificate.
     * 
     * @param bytes
     * @return
     */
    public static String encodeCertificate(byte[] bytes) {
        return encodeKey(bytes, BEGIN_CERTIFICATE, END_CERTIFICATE);
    }

    /**
     * MIME-encode the byte array as a public key.
     * 
     * @param bytes
     * @return
     */
    public static String encodePublicKey(byte[] bytes) {
        return encodeKey(bytes, BEGIN_PUB_KEY, END_PUB_KEY);
    }

    /**
     * MIME-encode the byte array as a private key.
     * 
     * @param bytes
     * @return
     */
    public static String encodePrivateKey(byte[] bytes) {
        return encodeKey(bytes, BEGIN_PRIVATE_KEY, END_PRIVATE_KEY);
    }

    private static String encodeKey(byte[] bytes, String prefix, String suffix) {
        String base64Data = Base64.getMimeEncoder().encodeToString(bytes);
        String out = prefix + "\n" + base64Data + "\n" + suffix;
        return out;
    }

    /**
     * Decode certificate from MIME-encoded string.
     * 
     * @param certString
     * @return
     * @throws IllegalArgumentException
     */
    public static byte[] decodeCertificate(String certString) throws IllegalArgumentException {
        return decodeKey(certString, BEGIN_CERTIFICATE, END_CERTIFICATE);
    }

    /**
     * Decode certificate from MIME-encoded file.
     * 
     * @param keyString
     * @return
     * @throws IllegalArgumentException
     */
    public static byte[] decodeCertificate(Path path) throws IllegalArgumentException, IOException {
        String certString = new String(Files.readAllBytes(path), Util.CHARSET);
        return decodeCertificate(certString);
    }

    /**
     * Decode public key from MIME-encoded string.
     * 
     * @param keyString
     * @return
     * @throws IllegalArgumentException
     */
    public static byte[] decodePublicKey(String keyString) throws IllegalArgumentException {
        return decodeKey(keyString, BEGIN_PUB_KEY, END_PUB_KEY);
    }

    /**
     * Decode public key from MIME-encoded file.
     * 
     * @param path
     * @return Public key bytes
     * @throws IllegalArgumentException
     * @throws IOException
     */
    public static byte[] decodePublicKey(Path path) throws IllegalArgumentException, IOException {
        String keyString = new String(Files.readAllBytes(path), Util.CHARSET);
        return decodePublicKey(keyString);
    }

    /**
     * Decode private key from MIME-encoded string.
     * 
     * @param keyString
     * @return
     * @throws IllegalArgumentException
     */
    public static byte[] decodePrivateKey(String keyString) throws IllegalArgumentException {
        return decodeKey(keyString, BEGIN_PRIVATE_KEY, END_PRIVATE_KEY);
    }

    /**
     * Decode private key from MIME-encoded file.
     * 
     * @param path
     * @return Private key bytes
     * @throws IllegalArgumentException
     * @throws IOException
     */
    public static byte[] decodePrivateKey(Path path) throws IllegalArgumentException, IOException {
        String keyString = new String(Files.readAllBytes(path), Util.CHARSET);
        return decodePrivateKey(keyString);
    }

    private static byte[] decodeKey(String keyString, String prefix, String suffix)
            throws IllegalArgumentException {
        keyString = keyString.trim();
        if (!keyString.startsWith(prefix) || !keyString.endsWith(suffix)) {
            throw new IllegalArgumentException("The key does not have expected format");
        }
        keyString = keyString.substring(keyString.indexOf(prefix) + prefix.length(),
                keyString.indexOf(suffix));
        return Base64.getMimeDecoder().decode(toBytes(keyString));
    }

    /**
     * @return An input stream for reading the resource, or <tt>null</tt> if the resource could not
     *         be found
     */
    public static InputStream getResource(String name) {
        return Util.class.getClassLoader().getResourceAsStream(name);
    }

    /**
     * @return A <tt>URL</tt> object for reading the resource, or <tt>null</tt> if the resource
     *         could not be found or the invoker doesn't have adequate privileges to get the
     *         resource.
     */
    public static URL getResourceUrl(String name) {
        return Util.class.getClassLoader().getResource(name);
    }

    /**
     * Read the input stream and return it as an X509Certificate.
     * 
     * @param pem
     * @return
     * @throws CertificateException
     */
    public static X509Certificate readCertAsPem(InputStream pem) throws CertificateException {
        CertificateFactory certificateFactory = CertificateFactory.getInstance(X509);
        return (X509Certificate) certificateFactory.generateCertificate(pem);
    }

    /**
     * Join sanitized identifier and name, return as Path.
     * 
     * Everything except lower and upper case letters, digits and symbols '-', '_' are replaced with
     * underscores in the identifier. More formally, the regular expression
     * <code>[^a-zA-Z0-9-_]</code> is applied and all matches are replaced with underscores.
     * 
     * The identifier and name are joined with hyphen '-'.
     * 
     * @param identifier
     * @param name
     * @return
     */
    public static Path prefixedPath(String identifier, String name) {
        return Paths.get(sanitize(identifier) + "-" + name);
    }

    /**
     * Sanitize the identifier.
     * 
     * Every match for the regular expression <code>[^a-zA-Z0-9-_]</code> are replaced with
     * underscores '_'. The regular expression correspond to all characters which are NOT lower or
     * upper case ASCII letters, digits or symbols '-', '_'.
     * 
     * @param identifier
     * @return
     */
    public static String sanitize(String identifier) {
        // this pattern matches anything which is not:
        // * lower case letters
        // * upper case letters
        // * digits
        // * symbols '-' and '_'
        String exclude = "[^a-zA-Z0-9-_]";
        return identifier.replaceAll(exclude, "_");
    }
}
