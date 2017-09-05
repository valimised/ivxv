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

public class SignatureUtil {
    public static class RSA {
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

        public static RSAPrivateKey paramsToRSAPrivateCrt(BigInteger d, BigInteger n) {
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
            PSSSigner pss = new PSSSigner(rsaeng, msgDigest, mgfDigest, salt, (byte) 1);
            pss.init(true, rsapar);
            pss.update(msg, 0, msg.length);
            try {
                ret = pss.generateSignature();
            } catch (CryptoException ex) {
                throw new SignatureException(ex);
            }
            return ret;
        }

        public static boolean verifyPSSSignature(byte[] msg, RSAPublicKey key, byte[] sig) {
            RSAKeyParameters rsapar =
                    new RSAKeyParameters(false, key.getModulus(), key.getPublicExponent());
            RSABlindedEngine rsaeng = new RSABlindedEngine();
            rsaeng.init(false, rsapar);
            Digest msgDigest = getPSSDigest();
            Digest mgfDigest = getPSSDigest();
            PSSSigner pss = new PSSSigner(rsaeng, msgDigest, mgfDigest, 32, (byte) 1);
            pss.init(false, rsapar);
            pss.update(msg, 0, msg.length);
            boolean ret = pss.verifySignature(sig);
            return ret;
        }

        public static byte[] PSSEncode(byte[] msg, BigInteger modulus, byte[] salt)
                throws SignatureException {
            RSAPrivateKey sk = paramsToRSAPrivateCrt(BigInteger.ONE, modulus);
            return generatePSSSignature(msg, sk, salt);
        }

        public static Digest getPSSDigest() {
            return new SHA256Digest();
        }

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

    public static byte[] stripSignature(byte[] sig) {
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
