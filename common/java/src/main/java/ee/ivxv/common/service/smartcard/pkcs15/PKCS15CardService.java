package ee.ivxv.common.service.smartcard.pkcs15;

import ee.ivxv.common.service.console.Console;
import ee.ivxv.common.service.i18n.I18n;
import ee.ivxv.common.service.smartcard.Card;
import ee.ivxv.common.service.smartcard.CardService;
import ee.ivxv.common.service.smartcard.Cards;
import ee.ivxv.common.util.I18nConsole;

public class PKCS15CardService implements CardService {

    private final I18nConsole console;

    public PKCS15CardService(Console console, I18n i18n) {
        this.console = new I18nConsole(console, i18n);
    }

    @Override
    public Card createCard(String id) {
        return new PKCS15Card(id, console);
    }

    @Override
    public Cards createCards() {
        return new Cards(this, console) {
            // Empty block
        };
    }

}
