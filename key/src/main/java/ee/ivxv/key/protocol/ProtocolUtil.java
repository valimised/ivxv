package ee.ivxv.key.protocol;

import ee.ivxv.common.math.Polynomial;
import ee.ivxv.common.service.smartcard.Card;
import ee.ivxv.common.service.smartcard.Cards;
import ee.ivxv.common.service.smartcard.IndexedBlob;
import ee.ivxv.common.service.smartcard.SmartCardException;
import java.math.BigInteger;

public class ProtocolUtil {

    public static byte[][] getAllBlobs(ThresholdParameters tparams, Cards cards,
            byte[] cardShareAID, byte[] cardShareName) throws ProtocolException {
        if (tparams.getParties() < cards.count()) {
            throw new ProtocolException("Number of cards larger than number of allowed parties");
        }
        byte[][] recovered = new byte[tparams.getParties()][];
        IndexedBlob ib;
        for (int i = 0; i < cards.count(); i++) {
            Card card = null;
            try {
                card = cards.getCard(i);
            } catch (SmartCardException e) {
                throw new ProtocolException("Can not get card", e);
            }
            try {
                ib = card.getIndexedBlob(cardShareAID, cardShareName);
            } catch (SmartCardException e) {
                throw new ProtocolException("Exception while reading secret share", e);
            }
            if (ib.getIndex() < 1 || ib.getIndex() > recovered.length) {
                throw new ProtocolException("Indexed blob index mismatch");
            }
            if (recovered[ib.getIndex() - 1] == null) {
                recovered[ib.getIndex() - 1] = ib.getBlob();
            } else {
                throw new ProtocolException("Duplicate indexed blob");
            }
        }
        return recovered;
    }

    public static BigInteger[] generateShares(Polynomial pol, int amount) {
        BigInteger[] shares = new BigInteger[amount];
        for (int i = 0; i < amount; i++) {
            shares[i] = pol.evaluate(i + 1);
        }
        return shares;
    }
}
