package ee.ivxv.common.service.smartcard;

import java.util.List;
import javax.smartcardio.CardException;
import javax.smartcardio.CardTerminal;
import javax.smartcardio.TerminalFactory;

public class TerminalUtil {
    private static final TerminalFactory factory = TerminalFactory.getDefault();

    public static List<CardTerminal> getTerminals() throws CardException {
        return factory.terminals().list();
    }

    public static CardTerminal getTerminal(int termNo) throws CardException, SmartCardException {
        List<CardTerminal> terminals = TerminalUtil.getTerminals();
        try {
            return terminals.get(termNo);
        } catch (IndexOutOfBoundsException e) {
            throw new SmartCardException(String.format("Terminal with index %d not found", termNo),
                    e);
        }
    }

    public static int getTerminalCount() throws CardException {
        return getTerminals().size();
    }

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
