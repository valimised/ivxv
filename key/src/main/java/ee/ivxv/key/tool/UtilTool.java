package ee.ivxv.key.tool;

import ee.ivxv.common.cli.Arg;
import ee.ivxv.common.cli.Args;
import ee.ivxv.common.cli.Tool;
import ee.ivxv.common.service.smartcard.CardInfo;
import ee.ivxv.common.service.smartcard.Cards;
import ee.ivxv.common.service.smartcard.SmartCardException;
import ee.ivxv.common.service.smartcard.pkcs15.PKCS15Card;
import ee.ivxv.common.util.I18nConsole;
import ee.ivxv.key.KeyContext;
import ee.ivxv.key.Msg;
import ee.ivxv.key.protocol.ThresholdParameters;
import ee.ivxv.key.tool.UtilTool.UtilArgs;
import java.util.List;
import javax.smartcardio.CardException;
import javax.smartcardio.CardTerminal;
import javax.smartcardio.TerminalFactory;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * UtilTool is a tool for running various helper methods within a key application.
 * <p>
 * Currently the tool supports listing smart card readers, the cards inserted to the readers and
 * reading the card information files from the cards.
 * <p>
 * Also, it is possible to test the functionality of the private key shares stored on the card
 * tokens.
 */
public class UtilTool implements Tool.Runner<UtilArgs> {
    public static class UtilArgs extends Args {
        Arg<Boolean> listReaders = Arg.aFlag(Msg.u_listreaders);

        public UtilArgs() {
            super();
            args.add(listReaders);
        }
    }

    private static final Logger log = LoggerFactory.getLogger(UtilTool.class);

    static Cards listCards(KeyContext ctx, ThresholdParameters tparams) {
        Cards cards = ctx.card.createCards();
        for (int i = 0; i < tparams.getParties(); i++) {
            cards.addCard(String.valueOf(i));
        }
        return cards;
    }

    private final I18nConsole console;

    private final KeyContext ctx;

    public UtilTool(KeyContext ctx) {
        this.ctx = ctx;
        this.console = new I18nConsole(ctx.i.console, ctx.i.i18n);
    }

    private void listReaders() throws CardException {
        TerminalFactory factory = TerminalFactory.getDefault();
        try {
            List<CardTerminal> terminals = factory.terminals().list();
            if (terminals.size() == 0) {
                console.println(Msg.e_no_cardterminals_found);
                return;
            }
            String idStr = console.i18n.get(Msg.m_id);
            String nameStr = console.i18n.get(Msg.m_name);
            String withCardStr = console.i18n.get(Msg.m_with_card);
            String cardIdStr = console.i18n.get(Msg.m_card_id);
            String yesStr = console.i18n.get(Msg.m_yes);
            String noStr = console.i18n.get(Msg.m_no);
            int maxLen = nameStr.length();
            for (CardTerminal ct : terminals) {
                maxLen = Math.max(maxLen, ct.getName().length());
            }
            console.console.println(
                    "%-" + idStr.length() + "s | %-" + maxLen + "s | %-" + withCardStr.length()
                            + "s | %-" + cardIdStr.length() + "s",
                    idStr, nameStr, withCardStr, cardIdStr);
            for (int i = 0; i < terminals.size(); i++) {
                CardTerminal ct = terminals.get(i);
                String cardId = "-";
                if (ct.isCardPresent()) {
                    PKCS15Card card = new PKCS15Card("", console);
                    card.setTerminal(i);
                    try {
                        card.initialize();
                        CardInfo info = card.getCardInfo();
                        cardId = info == null ? "-" : info.getId();
                    } catch (SmartCardException e) {
                        log.debug("Couldn't get cardInfo", e);
                        cardId = "error";
                    }
                }
                console.console.println(
                        "%-" + idStr.length() + "d | %-" + maxLen + "s | %-" + withCardStr.length()
                                + "s | %s",
                        i, ct.getName(), ct.isCardPresent() ? yesStr : noStr, cardId);
            }
        } catch (CardException e) {
            if (e.getCause().getMessage().equals("SCARD_E_NO_READERS_AVAILABLE")) {
                console.println(Msg.e_no_cardterminals_found);
            } else {
                throw e;
            }
        }
    }

    @Override
    public boolean run(UtilArgs args) throws Exception {
        if (args.listReaders.value()) {
            listReaders();
        }
        return true;
    }
}
