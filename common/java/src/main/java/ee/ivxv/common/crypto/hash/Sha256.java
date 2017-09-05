package ee.ivxv.common.crypto.hash;

import java.io.IOException;
import java.io.InputStream;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;

public class Sha256 implements HashFunction {
    private MessageDigest md;

    public Sha256() {
        try {
            md = MessageDigest.getInstance("SHA-256");
        } catch (NoSuchAlgorithmException e) {
            throw new RuntimeException("SHA-256 algorithm not found: " + e.getMessage(), e);
        }
    }

    @Override
    public byte[] digest(byte[] input) {
        return md.digest(input);
    }

    @Override
    public byte[] digest(InputStream input) throws IOException {
        byte[] buffer = new byte[2048];

        for (int read; (read = input.read(buffer)) != -1;) {
            md.update(buffer, 0, read);
        }

        return md.digest();
    }

    @Override
    public String toString() {
        return "SHA-256";
    }
}
