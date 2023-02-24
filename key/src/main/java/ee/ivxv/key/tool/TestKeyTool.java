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
import ee.ivxv.common.service.smartcard.Card;
import ee.ivxv.common.service.smartcard.CardService;
import ee.ivxv.common.service.smartcard.Cards;
import ee.ivxv.common.service.smartcard.IndexedBlob;
import ee.ivxv.common.util.I18nConsole;
import ee.ivxv.common.util.Util;
import ee.ivxv.key.KeyContext;
import ee.ivxv.key.Msg;
import ee.ivxv.key.protocol.ProtocolException;
import ee.ivxv.key.protocol.ThresholdParameters;
import ee.ivxv.key.protocol.decryption.recover.RecoverDecryption;
import ee.ivxv.key.protocol.signing.shoup.ShoupSigning;
import ee.ivxv.key.tool.TestKeyTool.TestKeyArgs;
import ee.ivxv.key.util.QuorumUtil;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.security.KeyException;
import java.security.KeyFactory;
import java.security.NoSuchAlgorithmException;
import java.security.interfaces.RSAPublicKey;
import java.security.spec.InvalidKeySpecException;
import java.security.spec.X509EncodedKeySpec;
import java.util.ArrayList;
import java.util.List;
import java.util.Set;
import javax.smartcardio.CardException;
import org.bouncycastle.asn1.x509.SubjectPublicKeyInfo;
import org.bouncycastle.cert.X509CertificateHolder;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * TestKeyTool is a tool for running various helper methods within a key application.
 * <p>
 * Currently the tool supports listing smart card readers, the cards inserted to the readers and
 * reading the card information files from the cards.
 * <p>
 * Also, it is possible to test the functionality of the private key shares stored on the card
 * tokens.
 */
public class TestKeyTool implements Tool.Runner<TestKeyArgs> {
    public static class TestKeyArgs extends Args {
        Arg<String> identifier = Arg.aString(Msg.arg_identifier);
        Arg<Path> outputPath = Arg.aPath(Msg.arg_out, true, true);
        Arg<Integer> dm = Arg.anInt(Msg.arg_threshold);
        Arg<Integer> dn = Arg.anInt(Msg.arg_parties);
        Arg<Boolean> fm = Arg.aFlag(Msg.arg_fastmode).setDefault(true);

        public TestKeyArgs() {
            super();
            args.add(identifier);
            args.add(outputPath);
            args.add(dm);
            args.add(dn);
            args.add(fm);
        }
    }

    private static final Logger log = LoggerFactory.getLogger(TestKeyTool.class);

    static final Plaintext TEST_MESSAGE = new Plaintext(String.format("%s%s%s%s%s", "123.321",
            Util.UNIT_SEPARATOR, "PARTY", Util.UNIT_SEPARATOR, "NAME SURNAME"));

    static void allRecoverTests(I18nConsole console, Logger log, List<Set<IndexedBlob>> quorums,
            ElGamalPublicKey pub, ThresholdParameters tparams, Rnd rnd, Plaintext message)
            throws ProtocolException, IOException {
        console.println(Msg.m_test_decryption_key);
        ElGamalCiphertext c;
        try {
            c = pub.encrypt(message, rnd);
        } catch (KeyException e) {
            console.println(Msg.e_testencryption_fail);
            log.debug("Encryption failed", e);
            throw new MessageException(Msg.e_testencryption_fail);
        }

        for (Set<IndexedBlob> quorum : quorums) {
            recoverTest(console, log, quorum, c, tparams, message);
        }
    }

    static void allShoupTests(I18nConsole console, Logger log, List<Set<IndexedBlob>> quorums,
            RSAPublicKey rsaPub, ThresholdParameters tparams, Rnd rnd, Plaintext message)
            throws ProtocolException {
        console.println(Msg.m_test_signature_key);
        for (Set<IndexedBlob> quorum : quorums) {
            shoupTest(console, log, quorum, rsaPub, tparams, rnd, message);
        }
    }

    static void allTests(I18nConsole console, Logger log, CardService cardService,
            ThresholdParameters tparams, Rnd rnd, ElGamalPublicKey pub, RSAPublicKey rsaPub)
            throws ProtocolException, IOException {
        try {
            Cards cards = cardService.createCards();
            if (!cardService.isPluggableService()) {
                for (int i = 0; i < tparams.getParties(); i++) {
                    cards.addCard(String.valueOf(i));
                }
            }
            List<IndexedBlob> decList = new ArrayList<>();
            List<IndexedBlob> signList = new ArrayList<>();
            for (int i = 0; i < tparams.getParties(); i++) {
                int retryCount = 0;
                int maxTries = 2;
                while (true) {
                    try {
                        Card card;
                        if (cardService.isPluggableService()) {
                            card = cardService.createCard("-1");
                            cards.initUnprocessedCard(card);
                        } else {
                            card = cards.getCard(i);
                        }
                        IndexedBlob ib = card.getIndexedBlob(InitTool.AID, InitTool.DEC_SHARE_NAME);
                        if (ib.getIndex() < 1 || ib.getIndex() > tparams.getParties()) {
                            throw new ProtocolException("Indexed blob index mismatch");
                        }
                        decList.add(ib);

                        ib = card.getIndexedBlob(InitTool.AID, InitTool.SIGN_SHARE_NAME);
                        if (ib.getIndex() < 1 || ib.getIndex() > tparams.getParties()) {
                            throw new ProtocolException("Indexed blob index mismatch");
                        }
                        signList.add(ib);
                        break;
                    } catch (ProtocolException e) {
                        throw e;
                    } catch (Exception e) {
                        if (++retryCount == maxTries) throw e;
                    }
                }
            }

            List<Set<IndexedBlob>> recoverQuorums =
                    QuorumUtil.getQuorumList(decList, tparams.getThreshold());
            List<Set<IndexedBlob>> shoupQuorums =
                    QuorumUtil.getQuorumList(signList, tparams.getThreshold());

            allRecoverTests(console, log, recoverQuorums, pub, tparams, rnd, TEST_MESSAGE);
            allShoupTests(console, log, shoupQuorums, rsaPub, tparams, rnd, TEST_MESSAGE);
        } catch (Exception e) {
            throw new IOException(e);
        }
    }

    static ElGamalPublicKey readDecryptionKey(TestKeyArgs args) throws IOException {
        Path encCertPath = Util.prefixedPath(args.identifier.value(), InitTool.ENC_CERT_TMPL);
        Path fullpath = args.outputPath.value().resolve(encCertPath);
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

    static RSAPublicKey readSigningKey(TestKeyArgs args)
            throws IOException, InvalidKeySpecException, NoSuchAlgorithmException {
        Path signCertPath = Util.prefixedPath(args.identifier.value(), InitTool.SIGN_CERT_TMPL);
        Path fullpath = args.outputPath.value().resolve(signCertPath);
        byte[] raw = readKey(fullpath);
        RSAPublicKey key = (RSAPublicKey) KeyFactory.getInstance("RSA")
                .generatePublic(new X509EncodedKeySpec(raw));
        return key;
    }

    static void recoverTest(I18nConsole console, Logger log, Set<IndexedBlob> quorum,
            ElGamalCiphertext c, ThresholdParameters tparams, Plaintext message)
            throws ProtocolException, IOException {
        RecoverDecryption dec = new RecoverDecryption(quorum, tparams);
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

    static void shoupTest(I18nConsole console, Logger log, Set<IndexedBlob> quorum,
            RSAPublicKey rsaPub, ThresholdParameters tparams, Rnd rnd, Plaintext message)
            throws ProtocolException {
        ShoupSigning shoupSign = new ShoupSigning(quorum, tparams, rnd);
        byte[] signature = shoupSign.sign(message.getMessage());
        boolean res = SignatureUtil.RSA.RSA_PSS.verifySignature(message.getMessage(), rsaPub, signature);
        if (res) {
            console.println(Msg.m_quorum_test_ok, quorum);
            log.debug("Signature creation test succeeded with quorum: {}", quorum);
        } else {
            log.debug("Signature creation test failed with quorum: {}", quorum);
            throw new MessageException(Msg.e_quorum_test_fail, quorum);
        }
    }

    static boolean testReaders(I18nConsole console, Logger log, KeyContext ctx, TestKeyArgs args)
            throws IOException, ProtocolException, CardException {
        ThresholdParameters tparams = new ThresholdParameters(args.dn.value(), args.dm.value());

        // boolean isFastMode = args.fm.value() && cards.enableFastMode(tparams.getParties());

        // Msg fastModeStatus = isFastMode ? Msg.m_fastmode_enabled : Msg.m_fastmode_disabled;
        // console.println(fastModeStatus);

        NativeRnd rnd = new NativeRnd();
        ElGamalPublicKey pub = readDecryptionKey(args);
        RSAPublicKey rsaPub;
        try {
            rsaPub = readSigningKey(args);
        } catch (InvalidKeySpecException | NoSuchAlgorithmException e) {
            throw new ProtocolException("Invalid RSA encoding", e);
        }
        allTests(console, log, ctx.card, tparams, rnd, pub, rsaPub);
        return true;
    }

    private final I18nConsole console;

    private final KeyContext ctx;


    public TestKeyTool(KeyContext ctx) {
        this.ctx = ctx;
        this.console = new I18nConsole(ctx.i.console, ctx.i.i18n);
    }

    @Override
    public boolean run(TestKeyArgs args) throws Exception {
        testReaders(console, log, ctx, args);
        return true;
    }
}
