package ee.ivxv.common.util;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.math.BigInteger;
import java.net.URL;
import java.nio.charset.Charset;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
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
     * @param string
     * @return Returns UTF-8 encoded byte array.
     */
    public static byte[] toBytes(String string) {
        return string.getBytes(CHARSET);
    }

    /**
     * @param bytes
     * @return Returns a string decoded from the bytes using UTF-8 encoding.
     */
    public static String toString(byte[] bytes) {
        return new String(bytes, CHARSET);
    }

    public static byte[] toBytes(InputStream stream) {
        return toBytes(stream, new byte[1024]);
    }

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

    public static byte[] toBytes(int i) {
        byte[] ret = new byte[4];
        ret[0] = (byte) ((i >>> 24) & 0xff);
        ret[1] = (byte) ((i >>> 16) & 0xff);
        ret[2] = (byte) ((i >>> 8) & 0xff);
        ret[3] = (byte) ((i >>> 0) & 0xff);
        return ret;
    }

    public static int toInt(byte[] b) {
        // we assume 4 bytes. if b is other length, then the result may be
        // undefined
        int k = 0;
        for (int i = 0; i < b.length && i < 4; i++) {
            k |= (b[i] << (24 - (i * 8)));
        }
        return k;
    }

    public static void createFile(Path path) throws IOException {
        if (path.getParent() != null) {
            Files.createDirectories(path.getParent());
        }
        Files.createFile(path);
    }

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

    public static String encodeCertificate(byte[] bytes) {
        return encodeKey(bytes, BEGIN_CERTIFICATE, END_CERTIFICATE);
    }

    public static String encodePublicKey(byte[] bytes) {
        return encodeKey(bytes, BEGIN_PUB_KEY, END_PUB_KEY);
    }

    public static String encodePrivateKey(byte[] bytes) {
        return encodeKey(bytes, BEGIN_PRIVATE_KEY, END_PRIVATE_KEY);
    }

    private static String encodeKey(byte[] bytes, String prefix, String suffix) {
        String base64Data = Base64.getMimeEncoder().encodeToString(bytes);
        String out = prefix + "\n" + base64Data + "\n" + suffix;
        return out;
    }

    public static byte[] decodeCertificate(String certString) throws IllegalArgumentException {
        return decodeKey(certString, BEGIN_CERTIFICATE, END_CERTIFICATE);
    }

    public static byte[] decodePublicKey(String keyString) throws IllegalArgumentException {
        return decodeKey(keyString, BEGIN_PUB_KEY, END_PUB_KEY);
    }

    public static byte[] decodePrivateKey(String keyString) throws IllegalArgumentException {
        return decodeKey(keyString, BEGIN_PRIVATE_KEY, END_PRIVATE_KEY);
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

    public static BigInteger safePrimeOrder(BigInteger p) {
        return p.subtract(BigInteger.ONE).divide(BigInteger.valueOf(2));
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

    public static X509Certificate readCertAsPem(InputStream pem) throws CertificateException {
        CertificateFactory certificateFactory = CertificateFactory.getInstance(X509);
        return (X509Certificate) certificateFactory.generateCertificate(pem);
    }

}
