package ee.ivxv.common.service.smartcard.dummy;

import ee.ivxv.common.service.console.Console;
import ee.ivxv.common.service.i18n.I18n;
import ee.ivxv.common.service.smartcard.Card;
import ee.ivxv.common.service.smartcard.CardService;
import ee.ivxv.common.service.smartcard.Cards;
import ee.ivxv.common.service.smartcard.SmartCardException;
import ee.ivxv.common.util.I18nConsole;
import java.io.IOException;
import java.nio.file.FileAlreadyExistsException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;

public class DummyCardService implements CardService {
    private static final String CARD_DIR_NAME = "dummy_cards";

    private final I18nConsole console;
    /** The directory where to store dummy card data, or null if not using persistent service. */
    private final Path cardDirPath;

    public DummyCardService(Console console, I18n i18n) {
        this(console, i18n, false);
    }

    public DummyCardService(Console console, I18n i18n, boolean persistent) {
        this.console = new I18nConsole(console, i18n);
        cardDirPath = persistent ? Paths.get(CARD_DIR_NAME) : null;
    }

    @Override
    public Card createCard(String id) {
        if (cardDirPath != null && !Files.exists(cardDirPath)) {
            try {
                Files.createDirectory(cardDirPath);
            } catch (FileAlreadyExistsException ignore) {
                // Allowed
            } catch (IOException e) {
                throw new RuntimeException(e);
            }
        }
        return new DummyPKCS15Card(id, console, cardDirPath);
    }

    @Override
    public Cards createCards() {
        return new Cards(this, console) {
            @Override
            public Card getCard(int index) throws SmartCardException {
                return cards.get(index);
            }
        };
    }

}
