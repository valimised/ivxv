package ee.ivxv.key.protocol.generation.shoup;

import ee.ivxv.common.crypto.SignatureUtil;
import ee.ivxv.common.crypto.rnd.Rnd;
import ee.ivxv.common.math.IntegerConstructor;
import ee.ivxv.common.math.MathUtil;
import ee.ivxv.common.math.Polynomial;
import ee.ivxv.common.service.smartcard.Card;
import ee.ivxv.common.service.smartcard.Cards;
import ee.ivxv.common.service.smartcard.SmartCardException;
import ee.ivxv.key.protocol.GenerationProtocol;
import ee.ivxv.key.protocol.ProtocolException;
import ee.ivxv.key.protocol.ProtocolUtil;
import ee.ivxv.key.protocol.ThresholdParameters;
import java.io.IOException;
import java.math.BigInteger;
import java.security.interfaces.RSAPrivateCrtKey;
import java.security.interfaces.RSAPublicKey;

public class ShoupGeneration implements GenerationProtocol {
    private final Cards cards;
    private final int modLen;
    private final ThresholdParameters tparams;
    private final Rnd rnd;
    private final byte[] cardShareAID;
    private final byte[] cardShareName;

    public ShoupGeneration(Cards cards, int modLen, ThresholdParameters tparams, Rnd rnd,
            byte[] cardShareAID, byte[] cardShareName) throws IOException {
        if (cards.count() < tparams.getParties()) {
            throw new IOException("Fewer cards available than requested number of shareholders");
        }
        this.cards = cards;
        this.tparams = tparams;
        this.rnd = rnd;
        this.cardShareAID = cardShareAID;
        this.cardShareName = cardShareName;
        this.modLen = modLen;
    }

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
        RSAPrivateCrtKey[] packedShares = packShares(e, shares, n);
        try {
            storeShares(packedShares);
        } catch (SmartCardException ex) {
            throw new ProtocolException(
                    "Error while communicating with smart card: " + ex.toString());
        }

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

    RSAPrivateCrtKey[] packShares(BigInteger e, BigInteger[] shares, BigInteger n) {
        RSAPrivateCrtKey[] packedShares = new RSAPrivateCrtKey[shares.length];
        for (int i = 0; i < packedShares.length; i++) {
            packedShares[i] = SignatureUtil.RSA.paramsToRSAPrivateKeyCrt(e, shares[i], n);
        }
        return packedShares;
    }

    void storeShares(RSAPrivateCrtKey[] shares) throws SmartCardException {
        for (int i = 0; i < shares.length; i++) {
            Card card = cards.getCard(i);
            card.storeIndexedBlob(cardShareAID, cardShareName, shares[i].getEncoded(),
                    i + 1);
        }
    }
}
