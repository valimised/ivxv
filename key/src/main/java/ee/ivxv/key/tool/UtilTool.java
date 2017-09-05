package ee.ivxv.key.tool;

import ee.ivxv.common.cli.Arg;
import ee.ivxv.common.cli.Args;
import ee.ivxv.common.cli.Tool;
import ee.ivxv.common.crypto.Plaintext;
import ee.ivxv.common.crypto.SignatureUtil;
import ee.ivxv.common.crypto.elgamal.ElGamalCiphertext;
import ee.ivxv.common.crypto.elgamal.ElGamalDecryptionProof;
import ee.ivxv.common.crypto.elgamal.ElGamalPublicKey;
import ee.ivxv.common.crypto.rnd.NativeRnd;
import ee.ivxv.common.crypto.rnd.Rnd;
import ee.ivxv.common.service.i18n.MessageException;
import ee.ivxv.common.service.smartcard.CardInfo;
import ee.ivxv.common.service.smartcard.Cards;
import ee.ivxv.common.service.smartcard.SmartCardException;
import ee.ivxv.common.service.smartcard.pkcs15.PKCS15Card;
import ee.ivxv.common.util.I18nConsole;
import ee.ivxv.common.util.Util;
import ee.ivxv.key.KeyContext;
import ee.ivxv.key.Msg;
import ee.ivxv.key.protocol.ProtocolException;
import ee.ivxv.key.protocol.ThresholdParameters;
import ee.ivxv.key.protocol.decryption.recover.RecoverDecryption;
import ee.ivxv.key.protocol.signing.shoup.ShoupSigning;
import ee.ivxv.key.tool.UtilTool.UtilArgs;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.security.KeyException;
import java.security.KeyFactory;
import java.security.NoSuchAlgorithmException;
import java.security.interfaces.RSAPublicKey;
import java.security.spec.InvalidKeySpecException;
import java.security.spec.X509EncodedKeySpec;
import java.util.List;
import javax.smartcardio.CardException;
import javax.smartcardio.CardTerminal;
import javax.smartcardio.TerminalFactory;
import org.bouncycastle.asn1.x509.SubjectPublicKeyInfo;
import org.bouncycastle.cert.X509CertificateHolder;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class UtilTool implements Tool.Runner<UtilArgs> {
    public static class UtilArgs extends Args {
        Arg<Boolean> listReaders = Arg.aFlag(Msg.u_listreaders);
        Arg<Path> outputPath = Arg.aPath(Msg.arg_out, true, true);
        Arg<Integer> dm = Arg.anInt(Msg.arg_threshold);
        Arg<Integer> dn = Arg.anInt(Msg.arg_parties);
        Arg<Args> testkey = new Arg.Tree(Msg.u_testkey, outputPath, dm, dn).setOptional();

        public UtilArgs() {
            super();
            args.add(listReaders);
            args.add(testkey);
        }
    }

    private static final Logger log = LoggerFactory.getLogger(UtilTool.class);

    static final Plaintext TEST_MESSAGE = new Plaintext(String.format("%s%s%s%s%s", "123.321",
            Util.UNIT_SEPARATOR, "PARTY", Util.UNIT_SEPARATOR, "NAME SURNAME"));

    static void allRecoverTests(I18nConsole console, Logger log, List<Cards> quorums,
            ElGamalPublicKey pub, ThresholdParameters tparams, Rnd rnd, byte[] aid,
            byte[] shareName, Plaintext message) throws ProtocolException, IOException {
        console.println(Msg.m_test_decryption_key);
        ElGamalCiphertext c;
        try {
            c = pub.encrypt(message, rnd);
        } catch (KeyException e) {
            console.println(Msg.e_testencryption_fail);
            log.debug("Encryption failed", e);
            throw new MessageException(Msg.e_testencryption_fail);
        }

        for (Cards quorum : quorums) {
            recoverTest(console, log, quorum, c, tparams, aid, shareName, message);
        }
    }

    static void allShoupTests(I18nConsole console, Logger log, List<Cards> quorums,
            RSAPublicKey rsaPub, ThresholdParameters tparams, Rnd rnd, byte[] aid, byte[] shareName,
            Plaintext message) throws ProtocolException {
        console.println(Msg.m_test_signature_key);
        for (Cards quorum : quorums) {
            shoupTest(console, log, quorum, rsaPub, tparams, rnd, aid, shareName, message);
        }
    }

    static void allTests(I18nConsole console, Logger log, Cards cards, ThresholdParameters tparams,
            Rnd rnd, ElGamalPublicKey pub, RSAPublicKey rsaPub)
            throws ProtocolException, IOException {
        List<Cards> quorums = cards.getQuorumList(tparams.getThreshold());
        allRecoverTests(console, log, quorums, pub, tparams, rnd, InitTool.AID,
                InitTool.DEC_SHARE_NAME, TEST_MESSAGE);
        allShoupTests(console, log, quorums, rsaPub, tparams, rnd, InitTool.AID,
                InitTool.SIGN_SHARE_NAME, TEST_MESSAGE);
    }

    static Cards listCards(KeyContext ctx, ThresholdParameters tparams) {
        Cards cards = ctx.card.createCards();
        for (int i = 0; i < tparams.getParties(); i++) {
            cards.addCard(String.valueOf(i));
        }
        return cards;
    }

    static ElGamalPublicKey readDecryptionKey(UtilArgs args) throws IOException {
        Path fullpath = args.outputPath.value().resolve(InitTool.ENC_CERT_PATH);
        byte[] raw = readKey(fullpath);
        ElGamalPublicKey key = new ElGamalPublicKey(raw);
        return key;
    }

    private static byte[] readKey(Path fullpath) throws IOException {
        byte[] content = Files.readAllBytes(fullpath);
        String certString = new String(content, Util.CHARSET);
        byte[] certBytes = Util.decodeCertificate(certString);
        X509CertificateHolder cert = new X509CertificateHolder(certBytes);
        SubjectPublicKeyInfo key = cert.getSubjectPublicKeyInfo();
        byte[] ret = key.getEncoded();
        return ret;
    }

    static RSAPublicKey readSigningKey(UtilArgs args)
            throws IOException, InvalidKeySpecException, NoSuchAlgorithmException {
        Path fullpath = args.outputPath.value().resolve(InitTool.SIGN_CERT_PATH);
        byte[] raw = readKey(fullpath);
        RSAPublicKey key = (RSAPublicKey) KeyFactory.getInstance("RSA")
                .generatePublic(new X509EncodedKeySpec(raw));
        return key;
    }

    static void recoverTest(I18nConsole console, Logger log, Cards quorum, ElGamalCiphertext c,
            ThresholdParameters tparams, byte[] aid, byte[] shareName, Plaintext message)
            throws ProtocolException, IOException {
        RecoverDecryption dec = new RecoverDecryption(quorum, tparams, aid, shareName);
        ElGamalDecryptionProof decProof = dec.decryptMessage(c.getBytes());
        boolean res = false;
        try {
            res = decProof.getDecrypted().equals(message);
        } catch (IllegalArgumentException e) {
            e.printStackTrace();
        }
        if (res) {
            console.println(Msg.m_quorum_test_ok, quorum);
            log.debug("Key usage test succeeded with quorum: {}", quorum);
        } else {
            log.debug("Key usage test failed with quorum: {}", quorum);
            throw new MessageException(Msg.e_quorum_test_fail, quorum);
        }
    }

    static void shoupTest(I18nConsole console, Logger log, Cards quorum, RSAPublicKey rsaPub,
            ThresholdParameters tparams, Rnd rnd, byte[] aid, byte[] shareName, Plaintext message)
            throws ProtocolException {
        ShoupSigning shoupSign = new ShoupSigning(quorum, tparams, aid, shareName, rnd);
        byte[] signature = shoupSign.sign(message.getMessage());
        boolean res = SignatureUtil.RSA.verifyPSSSignature(message.getMessage(), rsaPub, signature);
        if (res) {
            console.println(Msg.m_quorum_test_ok, quorum);
            log.debug("Signature creation test succeeded with quorum: {}", quorum);
        } else {
            log.debug("Signature creation test failed with quorum: {}", quorum);
            throw new MessageException(Msg.e_quorum_test_fail, quorum);
        }
    }

    static boolean testReaders(I18nConsole console, Logger log, KeyContext ctx, UtilArgs args)
            throws IOException, ProtocolException {
        ThresholdParameters tparams = new ThresholdParameters(args.dn.value(), args.dm.value());
        Cards cards = listCards(ctx, tparams);
        NativeRnd rnd = new NativeRnd();
        ElGamalPublicKey pub = readDecryptionKey(args);
        RSAPublicKey rsaPub;
        try {
            rsaPub = readSigningKey(args);
        } catch (InvalidKeySpecException | NoSuchAlgorithmException e) {
            throw new ProtocolException("Invalid RSA encoding", e);
        }
        allTests(console, log, cards, tparams, rnd, pub, rsaPub);
        return true;
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
        if (args.testkey.isSet()) {
            testReaders(console, log, ctx, args);
        }
        return true;
    }
}
