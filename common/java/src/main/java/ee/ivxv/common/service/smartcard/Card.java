package ee.ivxv.common.service.smartcard;

/**
 * Interface for abstract card token communication
 */
public interface Card {
    /**
     * Get the index of the terminal which is used.
     * <p>
     * 
     * @return Terminal identifier or -1 when terminal is not set.
     */
    int getTerminal();

    /**
     * Set the terminal for the card
     * 
     * @param terminalNo
     */
    void setTerminal(int terminalNo);

    /**
     * Initialize the connections to the card.
     * 
     * @throws SmartCardException
     */
    void initialize() throws SmartCardException;

    /**
     * Return if the card is initialized
     * 
     * @return
     */
    boolean isInitialized();

    /**
     * Close the connection to the card.
     * 
     * @throws SmartCardException
     */
    void close() throws SmartCardException;

    /**
     * Get the card identifier.
     * 
     * @return
     */
    String getId();

    /**
     * Get an indexed file from the card.
     * 
     * @param aid Authentication identifier if needed
     * @param identifier File identifier
     * @return
     * @throws SmartCardException
     */
    IndexedBlob getIndexedBlob(byte[] aid, byte[] identifier) throws SmartCardException;

    /**
     * Store an indexed blob with authentication identifier
     * 
     * @param aid Authentication identifier if needed
     * @param identifier File identifier to store at
     * @param blob File to store
     * @param index File index
     * @throws SmartCardException
     */
    void storeIndexedBlob(byte[] aid, byte[] identifier, byte[] blob, int index)
            throws SmartCardException;

    /**
     * Get card information
     * 
     * @return
     * @throws SmartCardException
     */
    CardInfo getCardInfo() throws SmartCardException;

    /**
     * Store card information.
     * 
     * @param data
     * @throws SmartCardException
     */
    void storeCardInfo(CardInfo data) throws SmartCardException;
}
