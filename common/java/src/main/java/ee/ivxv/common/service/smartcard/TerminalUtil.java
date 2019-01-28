package ee.ivxv.common.service.smartcard;

import java.util.List;
import javax.smartcardio.CardException;
import javax.smartcardio.CardTerminal;
import javax.smartcardio.TerminalFactory;

/**
 * Helper methods for managing card terminals.
 */
public class TerminalUtil {
    private static final TerminalFactory factory = TerminalFactory.getDefault();

    /**
     * Get all available terminals.
     * 
     * @return
     * @throws CardException
     */
    public static List<CardTerminal> getTerminals() throws CardException {
        return factory.terminals().list();
    }

    /**
     * Get specific terminal.
     * 
     * @param termNo
     * @return
     * @throws CardException
     * @throws SmartCardException
     */
    public static CardTerminal getTerminal(int termNo) throws CardException, SmartCardException {
        List<CardTerminal> terminals = TerminalUtil.getTerminals();
        try {
            return terminals.get(termNo);
        } catch (IndexOutOfBoundsException e) {
            throw new SmartCardException(String.format("Terminal with index %d not found", termNo),
                    e);
        }
    }

    /**
     * Get the number of available terminals.
     * 
     * @return
     * @throws CardException
     */
    public static int getTerminalCount() throws CardException {
        return getTerminals().size();
    }

    /**
     * Get the number of cards attached to the terminals.
     * 
     * @return
     * @throws CardException
     */
    public static int getCardCount() throws CardException {
        int cardCount = 0;
        for (CardTerminal terminal : getTerminals()) {
            if (terminal.isCardPresent()) {
                cardCount++;
            }
        }
        return cardCount;
    }
}
