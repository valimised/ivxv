package ee.ivxv.key.protocol.signing.shoup;

import ee.ivxv.common.asn1.RSAParams;
import ee.ivxv.common.crypto.SignatureUtil;
import ee.ivxv.common.crypto.rnd.Rnd;
import ee.ivxv.common.math.LagrangeInterpolation;
import ee.ivxv.common.math.MathUtil;
import ee.ivxv.common.service.smartcard.IndexedBlob;
import ee.ivxv.key.protocol.ProtocolException;
import ee.ivxv.key.protocol.SigningProtocol;
import ee.ivxv.key.protocol.ThresholdParameters;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.OutputStream;
import java.math.BigInteger;
import java.security.SignatureException;
import java.security.spec.InvalidKeySpecException;
import java.util.Set;
import org.bouncycastle.asn1.ASN1Primitive;
import org.bouncycastle.asn1.pkcs.PKCSObjectIdentifiers;
import org.bouncycastle.asn1.x509.AlgorithmIdentifier;

/**
 * Protocol for signing a message using threshold number of card tokens with RSA signing scheme.
 */
public class ShoupSigning implements SigningProtocol {
    private final Set<IndexedBlob> blobs;
    private final ThresholdParameters tparams;
    private final Rnd rnd;
    private final ByteArrayOutputStream out = new ByteArrayOutputStream();

    /**
     * Initialize the protocol using values.
     * 
     * @param blobs The set of blobs that contain the private key shares.
     * @param tparams Threshold parameters.
     * @param rnd Random source for signing.
     * @throws ProtocolException When the number of available cards is less than the threshold.
     */
    public ShoupSigning(Set<IndexedBlob> blobs, ThresholdParameters tparams, Rnd rnd)
            throws ProtocolException {
        if (blobs.size() < tparams.getThreshold()) {
            throw new ProtocolException("Fewer cards available than threshold");
        }
        this.blobs = blobs;
        this.tparams = tparams;
        this.rnd = rnd;
    }

    /**
     * Sign the message using RSA-PSS signing scheme.
     * <p>
     * The algorithm for constructing the signature is as follows: {@code
     *  1. get available blobs
     *  2. parse RSAPrivateKey from every blob
     *  2b. check that no keys have different modulus
     *  3. use RSA-PSS signing using every RSAPrivateKEy
     *  4. compute Lagrange coeficcient for every share
     *  5. exponentiate modulo n every signature share
     *  6. multiply the shares
     *  7. verify the signature
     * }
     * 
     * @return RSA-PSS signature
     * @throws ProtocolException When exception occurs during card token communication or signature
     *         share generation.
     */
    @Override
    public byte[] sign(byte[] msg) throws ProtocolException {

        byte[] salt = new byte[SignatureUtil.RSA.RSA_PSS.HASH_LENGTH];
        byte[] signature;
        byte[][] sigShares;
        RSAParams[] parsedKeys;
        try {
            parsedKeys = unpackAllBlobs(blobs);
        } catch (InvalidKeySpecException ex) {
            throw new ProtocolException("Invalid key blob: " + ex.toString());
        }
        if (!filterKeys(parsedKeys)) {
            throw new ProtocolException("Key share parameters mismatch");
        }
        BigInteger n = getModulus(parsedKeys);
        BigInteger e = getPublicExponent(parsedKeys);
        try {
            rnd.read(salt, 0, salt.length);
        } catch (IOException ex) {
            throw new ProtocolException("Reading from random source failed", ex);
        }
        try {
            sigShares = generateSignatureShares(msg, parsedKeys, salt);
        } catch (SignatureException ex) {
            throw new ProtocolException("Signature share generation failed: " + ex.toString());
        }
        try {
            signature = combineSignatureShares(sigShares, msg, n, e, salt);
        } catch (SignatureException ex) {
            throw new ProtocolException("Signature combining failed: " + ex.toString());
        }
        signature = SignatureUtil.stripSignature(signature);
        return signature;
    }

    private RSAParams[] unpackAllBlobs(Set<IndexedBlob> blobs) throws InvalidKeySpecException {
        RSAParams[] parsed = new RSAParams[tparams.getParties()];
        for (IndexedBlob blob : blobs) {
            parsed[blob.getIndex() - 1] = SignatureUtil.RSA.bytesToRSAParams(blob.getBlob());
        }
        return parsed;
    }

    private boolean filterKeys(RSAParams[] keys) throws ProtocolException {
        if (keys.length == 0) {
            throw new ProtocolException("No key shares parsed");
        }
        BigInteger mod = null;
        BigInteger pubexp = null;
        for (RSAParams sk : keys) {
            if (sk == null) {
                continue;
            }
            if (mod == null) {
                mod = sk.getModulus();
            }
            if (pubexp == null) {
                pubexp = sk.getPublicExponent();
            }
            if (sk.getModulus().compareTo(mod) != 0) {
                return false;
            }
            if (sk.getPublicExponent().compareTo(pubexp) != 0) {
                return false;
            }
        }
        return true;
    }

    private byte[][] generateSignatureShares(byte[] msg, RSAParams[] keys, byte[] salt)
            throws SignatureException {
        byte[][] shares = new byte[keys.length][];
        for (int i = 0; i < shares.length; i++) {
            if (keys[i] == null) {
                continue;
            }
            shares[i] = SignatureUtil.RSA.RSA_PSS.generateSignature(msg, keys[i], salt);
        }
        return shares;
    }

    private byte[] combineSignatureShares(byte[][] shares, byte[] msg, BigInteger n, BigInteger e,
            byte[] salt) throws SignatureException {
        BigInteger c;
        BigInteger sigShare = BigInteger.ONE;
        BigInteger exp;
        byte[] encShare;
        for (int i = 0; i < shares.length; i++) {
            if (shares[i] == null) {
                continue;
            }
            c = new BigInteger(1, shares[i]);
            exp = LagrangeInterpolation.basisInverselessPolynomial(shares,
                    BigInteger.valueOf(i + 1));
            c = c.modPow(exp, n);
            sigShare = sigShare.multiply(c).mod(n);
        }
        BigInteger[] bezout = MathUtil.extendedEuclidean(e,
                MathUtil.factorial(BigInteger.valueOf(tparams.getParties())));
        sigShare = sigShare.modPow(bezout[1], n);
        encShare = SignatureUtil.RSA.RSA_PSS.encode(msg, n, salt);
        c = new BigInteger(1, encShare);
        c = c.modPow(bezout[0], n);
        sigShare = sigShare.multiply(c).mod(n);
        return sigShare.toByteArray();
    }

    private BigInteger getModulus(RSAParams[] keys) {
        for (RSAParams sk : keys) {
            if (sk == null) {
                continue;
            }
            return sk.getModulus();
        }
        return null;
    }

    private BigInteger getPublicExponent(RSAParams[] keys) {
        for (RSAParams sk : keys) {
            if (sk == null) {
                continue;
            }
            return sk.getPublicExponent();
        }
        return null;
    }

    /**
     * Get the algorithm identifier for the signing scheme.
     * 
     * @return SHA256-with-RSA-Encryption
     */
    @Override
    public AlgorithmIdentifier getAlgorithmIdentifier() {
        ASN1Primitive params = SignatureUtil.RSA.RSA_PSS.getDefaultAlgorithmIdentifier();
        return new AlgorithmIdentifier(PKCSObjectIdentifiers.id_RSASSA_PSS, params);
    }

    /**
     * Get the stream for writing the message value to be signed.
     * 
     * @return Stream for writing the message.
     */
    @Override
    public OutputStream getOutputStream() {
        return out;
    }

    /**
     * Sign the message written to the output stream.
     * <p>
     * Sign the message written to the stream output in {@link #getOutputStream()}
     * 
     * @see #getOutputStream()
     * @return RSA-PSS signature
     */
    @Override
    public byte[] getSignature() {
        byte[] msg = out.toByteArray();
        out.reset();
        try {
            return sign(msg);
        } catch (ProtocolException e) {
            throw new RuntimeException(e);
        }
    }
}
