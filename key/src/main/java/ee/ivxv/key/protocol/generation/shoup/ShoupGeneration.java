package ee.ivxv.key.protocol.generation.shoup;

import ee.ivxv.common.asn1.RSAParams;
import ee.ivxv.common.crypto.SignatureUtil;
import ee.ivxv.common.crypto.rnd.Rnd;
import ee.ivxv.common.math.IntegerConstructor;
import ee.ivxv.common.math.MathUtil;
import ee.ivxv.common.math.Polynomial;
import ee.ivxv.common.service.smartcard.Cards;
import ee.ivxv.key.protocol.GenerationProtocol;
import ee.ivxv.key.protocol.ProtocolException;
import ee.ivxv.key.protocol.ProtocolUtil;
import ee.ivxv.key.protocol.ThresholdParameters;
import java.io.IOException;
import java.math.BigInteger;
import java.security.interfaces.RSAPublicKey;

/**
 * Generate the key for using with RSA algorithm.
 * <p>
 * The private key shares are stored on card tokens.
 */
public class ShoupGeneration implements GenerationProtocol {
    private final int modLen;
    private final ThresholdParameters tparams;
    private final Rnd rnd;
    private byte[][] sharestorage;

    /**
     * Initialize the protocol using values.
     * 
     * @param cards List of cards to store the private key shares on.
     * @param modLen Length of modulus in bits.
     * @param tparams Threshold scheme parameters indicating the number of shares and the threshold
     *        for signing.
     * @param rnd Random source.
     * @param cardShareAID Authentication identifier to be applied for private key shares.
     * @param cardShareName Private key share name on card tokens.
     * @throws IOException When the number of cards is less than the number of expected shares.
     */
    public ShoupGeneration(Cards cards, int modLen, ThresholdParameters tparams, Rnd rnd,
            byte[] cardShareAID, byte[] cardShareName, byte[][] sharestorage) throws IOException {
        if (cards.count() < tparams.getParties()) {
            throw new IOException("Fewer cards available than requested number of shareholders");
        }
        this.tparams = tparams;
        this.rnd = rnd;
        this.modLen = modLen;
        this.sharestorage = sharestorage;
    }

    /**
     * Generate RSA private key shares and output serialized public key.
     * <p>
     * The algorithm for constructing the key is shortly as follows: {@code
     *  1. generate primes p, q of size nLen
     *  2. generate public exponent 0 < e < (p-1)(q-1)
     *  3. find private exponent d = e^-1 mod (p-1)(q-1)
     *  4. generate polynomial with free entry d
     *  5. compute shares si = f(i) for every node 1 <= i <= n
     *  6. encode the shares (si, n) as ASN1 DER
     *  7. store the ASN1-encoded shares on card i
     *  8. encode the public key (n, e) as ASN1 DER
     *  9. output encoded public key
     * }
     * 
     * @return Serialized {@link java.security.interfaces.RSAPublicKey} instance.
     * @throws ProtocolException When exception occurs during protocol run
     * @throws IOException When exception occurs during card token communication.
     */
    @Override
    public byte[] generateKey() throws ProtocolException, IOException {
        // 1. generate primes p, q of size nLen
        // 2. generate public exponent 0 < e < (p-1)(q-1)
        // 3. find private exponent d = e^-1 mod (p-1)(q-1)
        // 4. generate polynomial with free entry d
        // 5. compute shares si = f(i) for every node 1 <= i <= n
        // 6. encode the shares (si, n) as ASN1 DER
        // 7. store the ASN1-encoded shares on card i
        // 8. encode the public key (n, e) as ASN1 DER
        // 9. output encoded public key
        BigInteger p, q, n, e, d;
        Polynomial pol;
        do {
            // we want e such that GCD(e, phi(p*q))=1. as e is actually fixed
            // below, then we need to search for suitable p and q
            p = generateCofactor();
            q = generateCofactor();
            n = computeModulus(p, q);
            e = generatePublicExponent(p, q);
        } while (!verifyCofactors(p, q, e));
        d = computePrivateExponent(p, q, e);
        pol = generatePolynomial(p, q, d);
        BigInteger[] shares = ProtocolUtil.generateShares(pol, tparams.getParties());
        RSAParams[] packedShares = packShares(e, shares, n);
        storeShares(packedShares);
        RSAPublicKey pk = SignatureUtil.RSA.paramsToRSAPublicKey(e, n);
        return pk.getEncoded();
    }

    BigInteger generateCofactor() throws IOException {
        BigInteger limit = BigInteger.ONE.shiftLeft(modLen / 2).subtract(BigInteger.ONE);
        return IntegerConstructor.constructPrime(rnd, limit);
    }

    BigInteger computeModulus(BigInteger p, BigInteger q) {
        return p.multiply(q);
    }

    BigInteger generatePublicExponent(BigInteger p, BigInteger q) {
        // we use a fixed public exponent for fast verification and signature
        // share combination
        return new BigInteger("65537");
    }

    boolean verifyCofactors(BigInteger p, BigInteger q, BigInteger e) {
        return e.gcd(MathUtil.phiSemiprime(p, q)).compareTo(BigInteger.ONE) == 0;
    }

    BigInteger computePrivateExponent(BigInteger p, BigInteger q, BigInteger e) {
        return e.modInverse(MathUtil.phiSemiprime(p, q));
    }

    Polynomial generatePolynomial(BigInteger p, BigInteger q, BigInteger d) throws IOException {
        return new Polynomial(tparams.getThreshold() - 1, MathUtil.phiSemiprime(p, q), d, rnd);
    }

    RSAParams[] packShares(BigInteger e, BigInteger[] shares, BigInteger n) {
        RSAParams[] packedShares = new RSAParams[shares.length];
        for (int i = 0; i < packedShares.length; i++) {
            packedShares[i] = SignatureUtil.RSA.paramsToRSAParams(e, shares[i], n);
        }
        return packedShares;
    }

    void storeShares(RSAParams[] shares) {
        for (int i = 0; i < shares.length; i++) {
            sharestorage[i] = shares[i].encode();
        }
    }
}
