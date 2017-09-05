package ee.ivxv.common.service.smartcard;

/**
 * CardService is an interface for creating smartcards and card sets.
 */
public interface CardService {

    /**
     * Creates new instance of uninitialized card with no terminal attached.
     * 
     * @param id
     * @return
     */
    Card createCard(String id);

    /**
     * Creates an empty instance of {@code Cards}.
     * 
     * @return
     */
    Cards createCards();

}
