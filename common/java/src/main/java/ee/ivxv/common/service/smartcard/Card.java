package ee.ivxv.common.service.smartcard;

// the idea behind this interface is that we provide methods for connecting to
// different types of cards depending the communication interface
public interface Card {
    // should return -1 when terminal not set
    int getTerminal();

    // set the terminal
    void setTerminal(int terminalNo);

    // get all possible cards, even the ones which are not in the terminal. The
    // index of a card in the array should correspond to terminalNo in previous
    // methods
    // public static Card[] getCards();
    // initialize the connection to the cards. it should throw some kind of
    // exception if it is not possible.
    void initialize() throws SmartCardException;

    // return if the card is initialized (i.e. the connection is set up)
    boolean isInitialized();

    // return if the card can be initialized with a call to initialize()
    boolean isAvailable();

    // close the connection to card
    void close() throws SmartCardException;

    // card id that the user recognizes card by
    String getId();

    IndexedBlob getIndexedBlob(byte[] aid, byte[] identifier) throws SmartCardException;

    void storeIndexedBlob(byte[] aid, byte[] identifier, byte[] blob, int index)
            throws SmartCardException;

    CardInfo getCardInfo() throws SmartCardException;

    void storeCardInfo(CardInfo data) throws SmartCardException;
}
