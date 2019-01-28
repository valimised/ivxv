package ee.ivxv.common.service.smartcard;

import ee.ivxv.common.util.I18nConsole;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.Iterator;
import java.util.List;
import java.util.Set;
import javax.smartcardio.CardException;
import javax.smartcardio.CardTerminal;

/**
 * Cards is a collection of cards. Class is abstract only to emphasize that instances should be
 * created using card service.
 */
public abstract class Cards {
    private static final int RETRY_COUNT = 3;

    private final CardService cardService;
    private final I18nConsole console;
    protected final List<Card> cards = new ArrayList<>();
    private final Set<String> processedCardIds = new HashSet<>();

    /**
     * Initialize using CardService and console.
     * 
     * @param cardService
     * @param console
     */
    public Cards(CardService cardService, I18nConsole console) {
        this.cardService = cardService;
        this.console = console;
    }

    /**
     * Enable fast mode if possible.
     * <p>
     * In fast mode, the user is not asked to insert the card terminal identifiers and they are
     * assigned automatically.
     * 
     * @param requiredCardCount Number of required cards
     * @return Boolean indicating if fast mode was enabled.
     * @throws CardException
     */
    public boolean enableFastMode(int requiredCardCount) throws CardException {
        if (!cardService.isPluggableService()) {
            return true;
        }
        if (cards.size() != requiredCardCount) {
            return false;
        }
        for (Card card : cards) {
            if (card.getTerminal() != -1) {
                return false;
            }
        }
        if (TerminalUtil.getCardCount() != requiredCardCount) {
            return false;
        }
        Iterator<Card> cardIterator = cards.iterator();
        List<CardTerminal> terminalList = TerminalUtil.getTerminals();
        for (int i = 0; i < terminalList.size(); i++) {
            if (terminalList.get(i).isCardPresent()) {
                cardIterator.next().setTerminal(i);
            }
        }
        return true;
    }

    /**
     * Add card.
     * 
     * @param id
     */
    public void addCard(String id) {
        cards.add(cardService.createCard(id));
    }

    /**
     * Get list of all cards
     * 
     * @return
     */
    public List<Card> getCardList() {
        return cards;
    }

    public void initUnprocessedCard(Card card) throws SmartCardException, CardException {
        List<CardTerminal> terminalList = TerminalUtil.getTerminals();
        while (true) {
            for (int i = 0; i < terminalList.size(); i++) {
                CardTerminal terminal = terminalList.get(i);
                if (!terminal.isCardPresent()) {
                    continue;
                }
                card.setTerminal(i);
                card.initialize();
                String id = card.getCardInfo().getId();
                if (!processedCardIds.contains(id)) {
                    processedCardIds.add(id);
                    console.println(Msg.inserted_card_id, id);
                    return;
                } else {
                    card.close();
                }
            }
            console.println(Msg.insert_unprocessed_card);
            console.console.readln();
        }
    }

    /**
     * Get card with specified index
     * 
     * @param index
     * @return
     * @throws SmartCardException
     */
    public Card getCard(int index) throws SmartCardException {
        Card card = cards.get(index);
        int termNo = card.getTerminal();
        if (termNo == -1) {
            termNo = askTermNo(index);
            card.setTerminal(termNo);
        }
        CardTerminal terminal = null;
        for (int i = RETRY_COUNT; i >= 0; i--) {
            try {
                terminal = TerminalUtil.getTerminal(termNo);
                break;
            } catch (CardException e) {
                if (i == 0) {
                    throw new SmartCardException("Failed to get terminal", e);
                }
            } catch (SmartCardException e) {
                if (i == 0) {
                    throw e;
                }
            }
        }
        for (int i = RETRY_COUNT; i >= 0; i--) {
            try {
                waitForSpecificCard(terminal, card, termNo, 0);
                break;
            } catch (CardException e) {
                if (i == 0) {
                    throw new SmartCardException("Failed to wait for smart card", e);
                }
            } catch (SmartCardException e) {
                if (i == 0) {
                    throw e;
                }
            }
        }
        return card;
    }

    private void waitForSpecificCard(CardTerminal terminal, Card card, int termNo, int tries)
            throws CardException, SmartCardException {
        tries = tries > 0 ? tries : Integer.MAX_VALUE;
        for (int i = 0; i < tries; i++) {
            waitUntilCardPresent(terminal, card, termNo);
            String currentCard = checkCardIndex(terminal, card, termNo);
            if (!currentCard.equals(card.getId())) {
                waitUntilCardRemoved(terminal, card, currentCard, termNo);
            } else {
                return;
            }
        }
    }

    private void waitUntilCardPresent(CardTerminal terminal, Card card, int termNo)
            throws CardException, SmartCardException {
        if (!terminal.isCardPresent()) {
            console.println(Msg.insert_card_indexed, card.getId(), termNo);
        }
        terminal.waitForCardPresent(0);
        if (!card.isInitialized()) {
            card.initialize();
        }
    }

    private void waitUntilCardRemoved(CardTerminal terminal, Card card, String currentCard,
            int termNo) throws CardException, SmartCardException {
        if (terminal.isCardPresent()) {
            console.println(Msg.remove_card_indexed, currentCard, termNo);
        }
        card.close();
        terminal.waitForCardAbsent(0);
    }

    private String checkCardIndex(CardTerminal terminal, Card card, int termNo)
            throws SmartCardException {
        CardInfo ci = card.getCardInfo();
        if (ci == null) {
            card.storeCardInfo(new CardInfo(card.getId()));
            return card.getId();
        } else {
            return ci.getId();
        }

    }

    private int askTermNo(int cardNo) {
        while (true) {
            console.println(Msg.enter_terminal_id, cardNo);
            try {
                // may return -1 which indicates in Card interface that the terminal is not set
                return Integer.parseInt(console.console.readln());
            } catch (NumberFormatException e) {
                // Ignored
            }
        }
    }

    @Override
    public String toString() {
        if (cards.size() == 0) {
            return "[]";
        }
        StringBuilder str = new StringBuilder("[" + cards.get(0).getId());

        for (int i = 1; i < cards.size(); i++) {
            str.append(", ").append(cards.get(i).getId());
        }
        return str.append("]").toString();
    }

    public int count() {
        return cards.size();
    }
}
