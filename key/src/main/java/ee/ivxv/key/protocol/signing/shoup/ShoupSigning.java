package ee.ivxv.key.protocol.signing.shoup;

import ee.ivxv.common.crypto.SignatureUtil;
import ee.ivxv.common.crypto.rnd.Rnd;
import ee.ivxv.common.math.LagrangeInterpolation;
import ee.ivxv.common.math.MathUtil;
import ee.ivxv.common.service.smartcard.Cards;
import ee.ivxv.key.protocol.ProtocolException;
import ee.ivxv.key.protocol.ProtocolUtil;
import ee.ivxv.key.protocol.SigningProtocol;
import ee.ivxv.key.protocol.ThresholdParameters;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.OutputStream;
import java.math.BigInteger;
import java.security.SignatureException;
import java.security.interfaces.RSAPrivateCrtKey;
import java.security.spec.InvalidKeySpecException;
import org.bouncycastle.asn1.pkcs.PKCSObjectIdentifiers;
import org.bouncycastle.asn1.x509.AlgorithmIdentifier;

public class ShoupSigning implements SigningProtocol {
    private final Cards cards;
    private final ThresholdParameters tparams;
    private final byte[] cardShareAID;
    private final byte[] cardShareName;
    private final Rnd rnd;
    private final ByteArrayOutputStream out = new ByteArrayOutputStream();

    public ShoupSigning(Cards cards, ThresholdParameters tparams, byte[] cardShareAID,
            byte[] cardShareName, Rnd rnd) throws ProtocolException {
        if (cards.count() < tparams.getThreshold()) {
            throw new ProtocolException("Fewer cards available than threshold");
        }
        this.cards = cards;
        this.tparams = tparams;
        this.cardShareAID = cardShareAID;
        this.cardShareName = cardShareName;
        this.rnd = rnd;
    }

    @Override
    public byte[] sign(byte[] msg) throws ProtocolException {
        // 1. get available blobs
        // 2. parse RSAPrivateKey from every blob
        // 2b. check that no keys have different modulus
        // 3. use RSA-PSS signing using every RSAPrivateKEy
        // 4. compute Lagrange coeficcient for every share
        // 5. exponentiate modulo n every signature share
        // 6. multiply the shares
        // 7. verify the signature

        byte[] salt = new byte[SignatureUtil.RSA.getPSSDigestLength()];
        byte[] signature;
        byte[][] blobs, sigShares;
        blobs = ProtocolUtil.getAllBlobs(tparams, cards, cardShareAID, cardShareName);
        RSAPrivateCrtKey[] parsedKeys;
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

    private RSAPrivateCrtKey[] unpackAllBlobs(byte[][] blobs) throws InvalidKeySpecException {
        RSAPrivateCrtKey[] parsed = new RSAPrivateCrtKey[blobs.length];
        for (int i = 0; i < blobs.length; i++) {
            parsed[i] = SignatureUtil.RSA.bytesToRSAPrivateKeyCrt(blobs[i]);
        }
        return parsed;
    }

    private boolean filterKeys(RSAPrivateCrtKey[] keys) throws ProtocolException {
        if (keys.length == 0) {
            throw new ProtocolException("No key shares parsed");
        }
        BigInteger mod = null;
        BigInteger pubexp = null;
        for (RSAPrivateCrtKey sk : keys) {
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

    private byte[][] generateSignatureShares(byte[] msg, RSAPrivateCrtKey[] keys, byte[] salt)
            throws SignatureException {
        byte[][] shares = new byte[keys.length][];
        for (int i = 0; i < shares.length; i++) {
            if (keys[i] == null) {
                continue;
            }
            shares[i] = SignatureUtil.RSA.generatePSSSignature(msg, keys[i], salt);
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
        encShare = SignatureUtil.RSA.PSSEncode(msg, n, salt);
        c = new BigInteger(1, encShare);
        c = c.modPow(bezout[0], n);
        sigShare = sigShare.multiply(c).mod(n);
        return sigShare.toByteArray();
    }

    private BigInteger getModulus(RSAPrivateCrtKey[] keys) {
        for (RSAPrivateCrtKey sk : keys) {
            if (sk == null) {
                continue;
            }
            return sk.getModulus();
        }
        return null;
    }

    private BigInteger getPublicExponent(RSAPrivateCrtKey[] keys) {
        for (RSAPrivateCrtKey sk : keys) {
            if (sk == null) {
                continue;
            }
            return sk.getPublicExponent();
        }
        return null;
    }

    @Override
    public AlgorithmIdentifier getAlgorithmIdentifier() {
        return new AlgorithmIdentifier(PKCSObjectIdentifiers.sha256WithRSAEncryption);
    }

    @Override
    public OutputStream getOutputStream() {
        return out;
    }

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
