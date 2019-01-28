package ee.ivxv.common.crypto;

import java.math.BigInteger;
import java.security.KeyFactory;
import java.security.NoSuchAlgorithmException;
import java.security.NoSuchProviderException;
import java.security.SignatureException;
import java.security.interfaces.RSAPrivateCrtKey;
import java.security.interfaces.RSAPrivateKey;
import java.security.interfaces.RSAPublicKey;
import java.security.spec.InvalidKeySpecException;
import java.security.spec.PKCS8EncodedKeySpec;
import java.security.spec.RSAPrivateCrtKeySpec;
import java.security.spec.RSAPrivateKeySpec;
import java.security.spec.RSAPublicKeySpec;
import java.security.spec.X509EncodedKeySpec;
import org.bouncycastle.crypto.CryptoException;
import org.bouncycastle.crypto.Digest;
import org.bouncycastle.crypto.digests.SHA256Digest;
import org.bouncycastle.crypto.engines.RSABlindedEngine;
import org.bouncycastle.crypto.params.RSAKeyParameters;
import org.bouncycastle.crypto.signers.PSSSigner;

/**
 * Helper functions for operating with signatures.
 */
public class SignatureUtil {
    /**
     * Helper functions for operating with RSA signatures.
     */
    public static class RSA {
        /**
         * Create a RSA private key using values with Chinese Remainder Theorem
         * 
         * @param e Public exponent
         * @param d Private exponent
         * @param n Modulus
         * @return
         */
        public static RSAPrivateCrtKey paramsToRSAPrivateKeyCrt(BigInteger e, BigInteger d,
                BigInteger n) {
            RSAPrivateCrtKeySpec spec = new RSAPrivateCrtKeySpec(n, e, d, BigInteger.ZERO,
                    BigInteger.ZERO, BigInteger.ZERO, BigInteger.ZERO, BigInteger.ZERO);
            KeyFactory factory = getRsaKeyFactory();
            RSAPrivateCrtKey sk;
            try {
                sk = (RSAPrivateCrtKey) factory.generatePrivate(spec);
            } catch (InvalidKeySpecException ex) {
                // in practice we don't get exception as RSA is supported
                throw new RuntimeException(ex);
            }
            return sk;
        }

        /**
         * Create a RSA private key using values
         * 
         * @param d Private exponent
         * @param n Modulus
         * @return
         */
        public static RSAPrivateKey paramsToRSAPrivateKey(BigInteger d, BigInteger n) {
            RSAPrivateKeySpec spec = new RSAPrivateKeySpec(n, d);
            KeyFactory factory = getRsaKeyFactory();
            RSAPrivateKey sk;
            try {
                sk = (RSAPrivateKey) factory.generatePrivate(spec);
            } catch (InvalidKeySpecException ex) {
                // in practice we don't get exception as RSA is supported
                throw new RuntimeException(ex);
            }
            return sk;
        }

        /**
         * Create RSA public key using values
         * 
         * @param e Public exponent
         * @param n Modulus
         * @return
         */
        public static RSAPublicKey paramsToRSAPublicKey(BigInteger e, BigInteger n) {
            RSAPublicKeySpec spec = new RSAPublicKeySpec(n, e);
            KeyFactory factory = getRsaKeyFactory();
            RSAPublicKey pk;
            try {
                pk = (RSAPublicKey) factory.generatePublic(spec);
            } catch (InvalidKeySpecException ex) {
                // in practice we don't get exception as RSA is supported
                throw new RuntimeException(ex);
            }
            return pk;
        }

        /**
         * Create RSA private key from PKCS8 serialized value
         * 
         * @param blob PKCS8 serialized RSA private key value
         * @return
         */
        public static RSAPrivateCrtKey bytesToRSAPrivateKeyCrt(byte[] blob) {
            if (blob == null) {
                return null;
            }
            PKCS8EncodedKeySpec spec = new PKCS8EncodedKeySpec(blob);
            KeyFactory factory = getRsaKeyFactory();
            RSAPrivateCrtKey sk;
            try {
                sk = (RSAPrivateCrtKey) factory.generatePrivate(spec);
            } catch (InvalidKeySpecException ex) {
                // in practice we don't get exception as RSA is supported
                throw new RuntimeException(ex);
            }
            return sk;
        }

        /**
         * Create RSA public key from X509 serialized value
         * 
         * @param blob X509 serialized RSA public key
         * @return
         */
        public static RSAPublicKey bytesToRSAPublicKey(byte[] blob) {
            if (blob == null) {
                return null;
            }
            X509EncodedKeySpec pkspec = new X509EncodedKeySpec(blob);
            KeyFactory factory = getRsaKeyFactory();
            RSAPublicKey pk;
            try {
                pk = (RSAPublicKey) factory.generatePublic(pkspec);
            } catch (InvalidKeySpecException ex) {
                // in practice we don't get exception as RSA is supported
                throw new RuntimeException(ex);
            }
            return pk;
        }

        /**
         * Generate RSA signature with PSS padding.
         * <p>
         * For the padding, the following parameters are used: SHA-256 hash, SHA-256 hash for MGF,
         * hash length 32 bytes, trailer byte 0xbc.
         * 
         * @param msg Message to be signed
         * @param key Key to use for signing
         * @param salt Salt to use for signing.
         * @return
         * @throws SignatureException
         */
        public static byte[] generatePSSSignature(byte[] msg, RSAPrivateKey key, byte[] salt)
                throws SignatureException {
            if (msg == null) {
                return null;
            }
            byte[] ret;
            RSAKeyParameters rsapar =
                    new RSAKeyParameters(true, key.getModulus(), key.getPrivateExponent());
            RSABlindedEngine rsaeng = new RSABlindedEngine();
            rsaeng.init(false, rsapar);
            Digest msgDigest = getPSSDigest();
            Digest mgfDigest = getPSSDigest();
            PSSSigner pss = new PSSSigner(rsaeng, msgDigest, mgfDigest, salt, (byte) 0xbc);
            pss.init(true, rsapar);
            pss.update(msg, 0, msg.length);
            try {
                ret = pss.generateSignature();
            } catch (CryptoException ex) {
                throw new SignatureException(ex);
            }
            return ret;
        }

        /**
         * Verify RSA signature with PSS padding
         * <p>
         * For the padding, the following parameters are used: SHA-256 hash, SHA-256 hash for MGF,
         * hash length 32 bytes, trailer byte 0xbc.
         * 
         * @param msg Message to be verified
         * @param key Key for verifying the message
         * @param sig Signature
         * @return Boolean indicating if verification succeeded.
         */
        public static boolean verifyPSSSignature(byte[] msg, RSAPublicKey key, byte[] sig) {
            RSAKeyParameters rsapar =
                    new RSAKeyParameters(false, key.getModulus(), key.getPublicExponent());
            RSABlindedEngine rsaeng = new RSABlindedEngine();
            rsaeng.init(false, rsapar);
            Digest msgDigest = getPSSDigest();
            Digest mgfDigest = getPSSDigest();
            PSSSigner pss = new PSSSigner(rsaeng, msgDigest, mgfDigest, 32, (byte) 0xbc);
            pss.init(false, rsapar);
            pss.update(msg, 0, msg.length);
            boolean ret = pss.verifySignature(sig);
            return ret;
        }

        /**
         * Add PSS padding to the message.
         * 
         * @param msg Message to be padded.
         * @param modulus Modulus of the private key.
         * @param salt Salt to use for padding.
         * @return Padded message.
         * @throws SignatureException When padding fails
         */
        public static byte[] PSSEncode(byte[] msg, BigInteger modulus, byte[] salt)
                throws SignatureException {
            RSAPrivateKey sk = paramsToRSAPrivateKey(BigInteger.ONE, modulus);
            return generatePSSSignature(msg, sk, salt);
        }

        /**
         * Get hash function to use for PSS padding. Returns SHA-256
         * 
         * @return SHA-256 hash function
         */
        public static Digest getPSSDigest() {
            return new SHA256Digest();
        }

        /**
         * Get the length of the hash function used for PSS padding. Returns 32.
         * 
         * @return 32
         */
        public static int getPSSDigestLength() {
            return getPSSDigest().getDigestSize();
        }

        private static KeyFactory getRsaKeyFactory() {
            try {
                // Must use provider name to prevent using BouncyCastle
                return KeyFactory.getInstance("RSA", "SunRsaSign");
            } catch (NoSuchAlgorithmException | NoSuchProviderException ex) {
                // in practice we don't get exception as RSA is supported
                throw new RuntimeException(ex);
            }
        }
    }

    /**
     * Strip excessive bytes from the beggining of the signature
     * <p>
     * @deprecated due to logic error. Do not start using as this method will be removed.
     * 
     * @param sig
     * @return
     */
    public static byte[] stripSignature(byte[] sig) {
        // TODO: in some cases, the first byte may be 0 legitimately. Stripping must depend on the
        // size of the modulus.
        byte[] ret;
        if (sig[0] == (byte) 0) {
            ret = new byte[sig.length - 1];
            System.arraycopy(sig, 1, ret, 0, sig.length - 1);
        } else {
            ret = sig;
        }
        return ret;
    }
}
