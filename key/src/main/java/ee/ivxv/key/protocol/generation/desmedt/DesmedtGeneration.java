package ee.ivxv.key.protocol.generation.desmedt;

import ee.ivxv.common.crypto.elgamal.ElGamalParameters;
import ee.ivxv.common.crypto.elgamal.ElGamalPrivateKey;
import ee.ivxv.common.crypto.elgamal.ElGamalPublicKey;
import ee.ivxv.common.crypto.rnd.Rnd;
import ee.ivxv.common.math.Polynomial;
import ee.ivxv.common.service.smartcard.Cards;
import ee.ivxv.key.protocol.GenerationProtocol;
import ee.ivxv.key.protocol.ProtocolException;
import ee.ivxv.key.protocol.ProtocolUtil;
import ee.ivxv.key.protocol.ThresholdParameters;
import java.io.IOException;
import java.math.BigInteger;

/**
 * DesmedtGeneration is a key generation for ElGamal crypto system.
 * <p>
 * In this protocol, the key is shared using Shamir secret sharing as private key shares. The shares
 * are stored on card tokens.
 */
public class DesmedtGeneration implements GenerationProtocol {
    private final ElGamalParameters params;
    private final ThresholdParameters tparams;
    private final Rnd rnd;
    private byte[][] sharestorage;

    /**
     * Initialize the protocol using values.
     * 
     * @param cards List of cards to store the private key shares.
     * @param params ElGamal crypto system parameters to use for key generation.
     * @param tparams Threshold parameters defining the number of share and threshold for
     *        decryption.
     * @param rnd Random source.
     * @param cardShareAID Authentication identifier to apply for private key shares.
     * @param cardShareName Identifier of the private key shares on card tokens.
     * @throws IOException When the number of available cards is less than the number of shares.
     */
    public DesmedtGeneration(Cards cards, ElGamalParameters params, ThresholdParameters tparams,
            Rnd rnd, byte[] cardShareAID, byte[] cardShareName, byte[][] sharestorage)
            throws IOException {
        if (cards.count() < tparams.getParties()) {
            throw new IOException("Fewer cards available than requested number of shareholders");
        }
        this.params = params;
        this.tparams = tparams;
        this.rnd = rnd;
        this.sharestorage = sharestorage;
    }

    /**
     * Generate the key and store the shares on card tokens.
     * <p>
     * The algorithm for generating the key is shortly as follows: {@code
     *  1. generate polynomial of degree t-1 (i.e. f(x) = a0 + a1 x + a2 x^2 + ... a(t-1) x^(t-1).
     *   2. compute shares si = f(i) for every node 1 <= i <= n
     *  3. compute modified share Ki = L(si, i), where L is the Lagrange interpolation coeficcient
     *   4. encode the modified shares as ASN1 DER
     *   5. store the modified shares at cards
     *   6. compute the public key y = g^f(0)
     *   7. encode the public key
     *   8. finally, return the public key
     *   }
     * 
     * @return Serialized {@link ee.ivxv.common.crypto.elgamal.ElGamalPublicKey} instance.
     * @throws ProtocolException When exception occurs during protocol run
     * @throws IOException When exception occurs during card token communication.
     * 
     */
    @Override
    public byte[] generateKey() throws ProtocolException, IOException {
        // 1. generate polynomial of degree t-1 (i.e. f(x) = a0 + a1 x + a2 x^2 + ... a(t-1)
        // x^(t-1).
        // 2. compute shares si = f(i) for every node 1 <= i <= n
        // 3. compute modified share Ki = L(si, i), where L is the Lagrange
        // interpolation coeficcient
        // 4. encode the modified shares as ASN1 DER
        // 5. store the modified shares at cards
        // 6. compute the public key y = g^f(0)
        // 7. encode the public key
        // 8. finally, return the public key

        Polynomial pol = generatePolynomial();
        BigInteger[] shares = ProtocolUtil.generateShares(pol, tparams.getParties());
        ElGamalPrivateKey[] packedShares = packShares(shares);
        storeShares(packedShares);
        ElGamalPublicKey pubKey = generatePublicKey(pol);
        byte[] ret;
        ret = pubKey.getBytes();
        return ret;
    }

    Polynomial generatePolynomial() throws IOException {
        return new Polynomial(tparams.getThreshold() - 1, params.getGeneratorOrder(), rnd);
    }

    ElGamalPrivateKey[] packShares(BigInteger[] shares) {
        ElGamalPrivateKey[] packedShares = new ElGamalPrivateKey[shares.length];
        for (int i = 0; i < shares.length; i++) {
            packedShares[i] = new ElGamalPrivateKey(params, shares[i]);
        }
        return packedShares;
    }

    void storeShares(ElGamalPrivateKey[] packedShares) {
        for (int i = 0; i < packedShares.length; i++) {
            sharestorage[i] = packedShares[i].getBytes();
        }
    }

    ElGamalPublicKey generatePublicKey(Polynomial pol) {
        return new ElGamalPublicKey(params, params.getGenerator().scale(pol.evaluate(0)));
    }
}
