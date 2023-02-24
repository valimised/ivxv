package ee.ivxv.common.crypto;

import ee.ivxv.common.asn1.ASN1DecodingException;
import ee.ivxv.common.asn1.RSAParams;
import java.io.IOException;
import java.math.BigInteger;
import java.security.KeyFactory;
import java.security.NoSuchAlgorithmException;
import java.security.NoSuchProviderException;
import java.security.SignatureException;
import java.security.interfaces.RSAPublicKey;
import java.security.spec.InvalidKeySpecException;
import java.security.spec.RSAPublicKeySpec;
import java.security.spec.X509EncodedKeySpec;
import org.bouncycastle.asn1.ASN1Encodable;
import org.bouncycastle.asn1.ASN1Integer;
import org.bouncycastle.asn1.ASN1Primitive;
import org.bouncycastle.asn1.DERSequence;
import org.bouncycastle.asn1.nist.NISTObjectIdentifiers;
import org.bouncycastle.asn1.pkcs.PKCSObjectIdentifiers;
import org.bouncycastle.asn1.pkcs.RSASSAPSSparams;
import org.bouncycastle.asn1.x509.AlgorithmIdentifier;
import org.bouncycastle.asn1.x509.SubjectPublicKeyInfo;
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
         * Helper functions for operating with RSA signatures using PSS padding scheme.
         */
        public static class RSA_PSS {
            public final static String HASH_FUN = "SHA-256";
            public final static String MFG1_FUN = "SHA-256";
            public final static int HASH_LENGTH = 32;
            public final static byte TRAILER_BYTE = (byte) 0xbc;

            /**
             * Generate RSA signature with PSS padding.
             * <p>
             * For the padding, the following parameters are used: SHA-256 hash, SHA-256 hash for
             * MGF, hash length 32 bytes, trailer byte 0xbc.
             *
             * @param msg Message to be signed
             * @param key Key to use for signing
             * @param salt Salt to use for signing.
             * @return
             * @throws SignatureException
             */
            public static byte[] generateSignature(byte[] msg, RSAParams key, byte[] salt)
                    throws SignatureException {
                if (msg == null) {
                    return null;
                }
                byte[] ret;
                RSAKeyParameters rsapar =
                        new RSAKeyParameters(true, key.getModulus(), key.getPrivateExponent());
                RSABlindedEngine rsaeng = new RSABlindedEngine();
                rsaeng.init(false, rsapar);
                Digest msgDigest = getDigest();
                Digest mgfDigest = getDigest();
                PSSSigner pss = new PSSSigner(rsaeng, msgDigest, mgfDigest, salt, TRAILER_BYTE);
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
             * For the padding, the following parameters are used: SHA-256 hash, SHA-256 hash for
             * MGF, hash length 32 bytes, trailer byte 0xbc.
             *
             * @param msg Message to be verified
             * @param key Key for verifying the message
             * @param sig Signature
             * @return Boolean indicating if verification succeeded.
             */
            public static boolean verifySignature(byte[] msg, RSAPublicKey key, byte[] sig) {
                RSAKeyParameters rsapar =
                        new RSAKeyParameters(false, key.getModulus(), key.getPublicExponent());
                RSABlindedEngine rsaeng = new RSABlindedEngine();
                rsaeng.init(false, rsapar);
                Digest msgDigest = getDigest();
                Digest mgfDigest = getDigest();
                PSSSigner pss =
                        new PSSSigner(rsaeng, msgDigest, mgfDigest, HASH_LENGTH, TRAILER_BYTE);
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
            public static byte[] encode(byte[] msg, BigInteger modulus, byte[] salt)
                    throws SignatureException {
                RSAParams sk = RSAParams.createPrivateKeyParams(BigInteger.ONE, modulus);
                return generateSignature(msg, sk, salt);
            }

            /**
             * Get hash function to use for PSS padding. Returns SHA-256
             *
             * @return SHA-256 hash function
             */
            private static Digest getDigest() {
                return new SHA256Digest();
            }

            /**
             * Get the AlgorithmIdentifier for a RSA key usable with PSS padding.
             *
             * @return
             */
            public static ASN1Primitive getDefaultAlgorithmIdentifier() {
                AlgorithmIdentifier hashAID =
                        new AlgorithmIdentifier(NISTObjectIdentifiers.id_sha256);
                RSASSAPSSparams params = new RSASSAPSSparams(hashAID,
                        new AlgorithmIdentifier(PKCSObjectIdentifiers.id_mgf1, hashAID),
                        new ASN1Integer(32), new ASN1Integer(1));
                return params.toASN1Primitive();
            }

            private static ASN1Primitive getSubjectPublicKey(BigInteger e, BigInteger n) {
                return new DERSequence(
                        new ASN1Encodable[] {new ASN1Integer(n), new ASN1Integer(e)});
            }

            private static SubjectPublicKeyInfo getSubjectPublicKeyInfo(BigInteger e,
                    BigInteger n) {
                SubjectPublicKeyInfo spki = null;
                try {
                    spki = new SubjectPublicKeyInfo(
                            new AlgorithmIdentifier(PKCSObjectIdentifiers.id_RSASSA_PSS),
                            getSubjectPublicKey(e, n));
                } catch (IOException e1) {
                    // the exception is checked
                }
                return spki;
            }

            /**
             * Get SubjectPublicKeyInfo of the given RSA public key.
             *
             * @param pub RSA public key
             * @return
             */
            public static SubjectPublicKeyInfo getSubjectPublicKeyInfo(RSAPublicKey pub) {
                return getSubjectPublicKeyInfo(pub.getPublicExponent(), pub.getModulus());
            }
        }

        /**
         * Create an encodable RSA parameters holder object.
         *
         * @param e Public exponent
         * @param d Private exponent
         * @param n Modulus
         * @return
         */
        public static RSAParams paramsToRSAParams(BigInteger e, BigInteger d, BigInteger n) {
            return new RSAParams(e, d, n);
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
        public static RSAParams bytesToRSAParams(byte[] blob) {
            if (blob == null) {
                return null;
            }
            RSAParams sk = new RSAParams();
            try {
                sk.readFromBytes(blob);
            } catch (ASN1DecodingException e) {
                throw new RuntimeException(e);
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
     * Strip excessive bytes from the beginning of the signature.
     *
     * @deprecated due to logic error. Do not start using as this method will be removed.
     *
     * @param sig
     * @return
     */
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
