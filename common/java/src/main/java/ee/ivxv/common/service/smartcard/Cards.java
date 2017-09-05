package ee.ivxv.common.service.smartcard;

import ee.ivxv.common.util.I18nConsole;
import java.util.ArrayList;
import java.util.List;
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

    public Cards(CardService cardService, I18nConsole console) {
        this.cardService = cardService;
        this.console = console;
    }

    public boolean enableFastMode(int requiredCardCount) throws CardException {
        if (cards.size() != requiredCardCount) {
            return false;
        }
        for (Card card : cards) {
            if (card.getTerminal() != -1) {
                return false;
            }
        }
        if (TerminalUtil.getTerminalCount() != requiredCardCount) {
            return false;
        }
        if (TerminalUtil.getCardCount() != requiredCardCount) {
            return false;
        }
        for (int i = 0; i < requiredCardCount; i++) {
            cards.get(i).setTerminal(i);
        }
        return true;
    }

    public void addCard(String id) {
        cards.add(cardService.createCard(id));
    }

    public List<Card> getCardList() {
        return cards;
    }

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

    private void waitUntilCardRemoved(CardTerminal terminal, Card card, String currentCard, int termNo)
            throws CardException, SmartCardException {
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

    /**
     * <pre>
     * {@literal
     * quorumSize == count():
     *  result list size is 1 and contains the object this was called on
     * quorumSize < count():
     *  creates 3 quorums such that all cards are at least in one quorum
     *  but none are in every quorum.
     *  1st quorum: first n cards among all cards
     *  2nd quorum: last n cards among all cards
     *  3rd quorum: first n-1 cards and the last card among all cards
     *  where n is parameter quorumSize.
     *  }
     * </pre>
     *
     * @param quorumSize number of cards in a single quorum
     * @throws IllegalArgumentException if {@code quorumSize > count()}
     */
    public List<Cards> getQuorumList(int quorumSize) {
        if (quorumSize > count()) {
            throw new IllegalArgumentException("Quorum size has to smaller than total card count");
        }
        List<Cards> res = new ArrayList<>();
        if (quorumSize == count()) {
            res.add(this);
        } else {
            Cards q1 = cardService.createCards();
            q1.cards.addAll(cards.subList(0, quorumSize));

            Cards q2 = cardService.createCards();
            q2.cards.addAll(cards.subList(count() - quorumSize, count()));

            Cards q3 = cardService.createCards();
            q3.cards.addAll(cards.subList(0, quorumSize - 1));
            q3.cards.add(cards.get(cards.size() - 1));

            res.add(q1);
            res.add(q2);
            res.add(q3);
        }
        return res;
    }
}
