package ee.ivxv.key.tool;

import ee.ivxv.common.cli.Arg;
import ee.ivxv.common.cli.Args;
import ee.ivxv.common.cli.Tool;
import ee.ivxv.common.crypto.SignatureUtil;
import ee.ivxv.common.crypto.elgamal.ElGamalParameters;
import ee.ivxv.common.crypto.elgamal.ElGamalPublicKey;
import ee.ivxv.common.crypto.rnd.Rnd;
import ee.ivxv.common.math.ECGroup;
import ee.ivxv.common.math.ECGroupElement;
import ee.ivxv.common.math.Group.Decodable;
import ee.ivxv.common.math.ModPGroup;
import ee.ivxv.common.math.ModPGroupElement;
import ee.ivxv.common.service.smartcard.Card;
import ee.ivxv.common.service.smartcard.Cards;
import ee.ivxv.common.service.smartcard.IndexedBlob;
import ee.ivxv.common.service.smartcard.SmartCardException;
import ee.ivxv.common.util.I18nConsole;
import ee.ivxv.common.util.Util;
import ee.ivxv.key.KeyContext;
import ee.ivxv.key.Msg;
import ee.ivxv.key.RandomSourceArg;
import ee.ivxv.key.protocol.ProtocolException;
import ee.ivxv.key.protocol.ThresholdParameters;
import ee.ivxv.key.protocol.generation.desmedt.DesmedtGeneration;
import ee.ivxv.key.protocol.generation.shoup.ShoupGeneration;
import ee.ivxv.key.protocol.signing.shoup.ShoupSigning;
import ee.ivxv.key.tool.InitTool.InitArgs;
import java.io.IOException;
import java.math.BigInteger;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.security.interfaces.RSAPublicKey;
import java.util.Date;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import javax.smartcardio.CardException;
import org.bouncycastle.asn1.ASN1Sequence;
import org.bouncycastle.asn1.x500.X500Name;
import org.bouncycastle.asn1.x509.SubjectPublicKeyInfo;
import org.bouncycastle.cert.X509CertificateHolder;
import org.bouncycastle.cert.X509v1CertificateBuilder;
import org.bouncycastle.operator.ContentSigner;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * InitTool is a tool for generating ElGamal encryption key pair and RSA signing key pair and
 * storing them on card tokens.
 */
public class InitTool implements Tool.Runner<InitArgs> {
    private static final Logger log = LoggerFactory.getLogger(InitTool.class);

    static final String ENC_CERT_TMPL = "enc.pem";
    static final String ENC_KEY_DER_TMPL = "pub.der";
    static final String ENC_KEY_PEM_TMPL = "pub.pem";
    static final String SIGN_CERT_TMPL = "sign.pem";

    static final byte[] AID = new byte[] {0x01};
    static final byte[] DEC_SHARE_NAME = "DEC".getBytes();
    static final byte[] SIGN_SHARE_NAME = "SIGN".getBytes();

    private final KeyContext ctx;
    private final I18nConsole console;

    public InitTool(KeyContext ctx) {
        this.ctx = ctx;
        this.console = new I18nConsole(ctx.i.console, ctx.i.i18n);
    }

    @Override
    public boolean run(InitArgs args) throws Exception {
        ElGamalParameters params = getParameters(args);
        log.debug("Gen keys with params:{}", params);
        Rnd rnd = null;
        try {
            rnd = RandomSourceArg.combineFromArgument(args.random);
        } catch (IOException e) {
            log.debug("Could not initialize random sources:", e);
            return false;
        }
        try {
            if (args.desmedt.isSet()) {
                desmedtGenProtocol(args, rnd, params);
            }
        } finally {
            rnd.close();
        }
        return true;
    }

    private void desmedtGenProtocol(InitArgs args, Rnd rnd, ElGamalParameters params)
            throws IOException, ProtocolException, CardException, SmartCardException {
        // PREPARE FOR KEYPAIR GENERATION
        ThresholdParameters tparams = new ThresholdParameters(args.dn.value(), args.dm.value());
        byte[][] encshares = new byte[tparams.getParties()][];
        byte[][] signshares = new byte[tparams.getParties()][];
        Cards cards = UtilTool.listCards(ctx, tparams);
        boolean isFastMode =
                args.enableFastMode.value() && cards.enableFastMode(tparams.getParties());

        Msg fastModeStatus = isFastMode ? Msg.m_fastmode_enabled : Msg.m_fastmode_disabled;
        console.println(fastModeStatus);

        // COLLECT ENTROPY
        console.println(Msg.m_collecting_required_randomness, args.requiredRandomness.value());
        byte[] tmp = new byte[args.requiredRandomness.value()];
        rnd.mustRead(tmp, 0, tmp.length);

        // GENERATE ENCRYPTION KEYPAIR
        console.println(Msg.m_generate_decryption_key);
        DesmedtGeneration gen =
                new DesmedtGeneration(cards, params, tparams, rnd, AID, DEC_SHARE_NAME, encshares);
        ElGamalPublicKey pub = new ElGamalPublicKey(gen.generateKey());

        // GENERATE SIGNATURE KEYPAIR
        console.println(Msg.m_generate_signature_key);
        ShoupGeneration shoupGen = new ShoupGeneration(cards, args.slen.value(), tparams, rnd, AID,
                SIGN_SHARE_NAME, signshares);

        RSAPublicKey rsaPub = SignatureUtil.RSA.bytesToRSAPublicKey(shoupGen.generateKey());

        console.println(Msg.m_storing_shares);
        for (int i = 0; i < encshares.length; i++) {
            Card card = cards.getCard(i);
            card.storeIndexedBlob(AID, DEC_SHARE_NAME, encshares[i], i + 1);
            card.storeIndexedBlob(AID, SIGN_SHARE_NAME, signshares[i], i + 1);
        }

        console.println(Msg.m_generating_certificate);
        // GENERATE CERTIFICATES FOR BOTH KEYPAIRS
        Set<IndexedBlob> signBlobs = new HashSet<>();
        for (int i = 0; i < tparams.getThreshold(); i++) {
            Card card;
            if (ctx.card.isPluggableService()) {
                card = ctx.card.createCard("-1");
                cards.initUnprocessedCard(card);
            } else {
                card = cards.getCard(i);
            }
            IndexedBlob ib = card.getIndexedBlob(AID, SIGN_SHARE_NAME);
            if (ib.getIndex() < 1 || ib.getIndex() > tparams.getParties()) {
                throw new ProtocolException("Indexed blob index mismatch");
            }
            signBlobs.add(ib);
        }
        Path signCertPath = Util.prefixedPath(args.identifier.value(), SIGN_CERT_TMPL);
        Path encCertPath = Util.prefixedPath(args.identifier.value(), ENC_CERT_TMPL);
        generateAndWriteCert(args, signBlobs, tparams, rnd, rsaPub.getEncoded(), signCertPath,
                args.signCN.value(), args.signCN.value(), args.signSN.value());
        generateAndWriteCert(args, signBlobs, tparams, rnd, pub.getBytes(), encCertPath,
                args.signCN.value(), args.encCN.value(), args.encSN.value());
        console.println(Msg.m_certificates_generated, signCertPath, encCertPath);

        Path encKeyDerPath = Util.prefixedPath(args.identifier.value(), ENC_KEY_DER_TMPL);
        Path encKeyPemPath = Util.prefixedPath(args.identifier.value(), ENC_KEY_PEM_TMPL);
        writeOutEncryptionKey(pub, args.outputPath.value(), encKeyDerPath, encKeyPemPath);
        console.println(Msg.m_keys_saved, encKeyDerPath, encKeyPemPath);

        if (!args.skipTests.value()) {
            // TEST THAT THE GENERATED KEY SUCCESSFULLY SIGNS
            TestKeyTool.allTests(console, log, ctx.card, tparams, rnd, pub, rsaPub);
        }
    }

    private void generateAndWriteCert(InitArgs args, Set<IndexedBlob> blobs,
            ThresholdParameters tparams,
            Rnd rnd, byte[] publicKey, Path path, String issuerCN, String cn, BigInteger sn)
            throws ProtocolException, IOException {
        ShoupSigning shoupSign =
                new ShoupSigning(blobs, tparams, rnd);
        X509CertificateHolder certHolder = genCert(shoupSign,
                SubjectPublicKeyInfo.getInstance(ASN1Sequence.getInstance(publicKey)), issuerCN, cn,
                sn);
        Path fullPath = args.outputPath.value().resolve(path);
        outputCert(certHolder, fullPath);
    }

    private ElGamalParameters getParameters(InitArgs args) {
        ElGamalParameters params;
        String electionId = args.identifier.value();
        if (args.mod.isSet()) {
            BigInteger p = args.groupP.value();
            BigInteger g = args.groupG.value();

            ModPGroup group = new ModPGroup(p, true);
            ModPGroupElement groupG = new ModPGroupElement(group, g);
            if (group.isDecodable(groupG) != Decodable.VALID) {
                throw new IllegalArgumentException("Invalid generator");
            }
            params = new ElGamalParameters(group, groupG, electionId);
        } else {
            ECGroup group = new ECGroup(args.curveName.value());
            ECGroupElement groupG = group.getBasePoint();
            params = new ElGamalParameters(group, groupG, electionId);
        }
        return params;
    }


    private X509CertificateHolder genCert(ContentSigner signer, SubjectPublicKeyInfo pub,
            String issuerCN, String subjectCN, BigInteger serial) throws IOException {
        Date startDate = new Date();
        Date expiryDate = new Date(startDate.getTime() + 365 * 24 * 60 * 60 * 1000L);
        X509v1CertificateBuilder v1CertGen =
                new X509v1CertificateBuilder(new X500Name("CN=" + issuerCN), serial, startDate,
                        expiryDate, new X500Name("CN=" + subjectCN), pub);

        return v1CertGen.build(signer);
    }


    private void outputCert(X509CertificateHolder cert, Path path) throws IOException {
        Util.createFile(path);
        String out = Util.encodeCertificate(cert.getEncoded());
        Files.write(path, Util.toBytes(out));
    }

    private void writeOutEncryptionKey(ElGamalPublicKey key, Path dir, Path der, Path pem)
            throws IOException {
        Files.write(dir.resolve(pem), Util.encodePublicKey(key.getBytes()).getBytes());
        Files.write(dir.resolve(der), key.getBytes());
    }



    public static class InitArgs extends Args {
        Arg<String> identifier = Arg.aString(Msg.arg_identifier);
        Arg<Path> outputPath = Arg.aPath(Msg.arg_out, false, null);
        Arg<Boolean> skipTests = Arg.aFlag(Msg.i_skiptest);
        Arg<List<RandomSourceArg.RndListEntry>> random = RandomSourceArg.getArgument();
        Arg<Integer> requiredRandomness = Arg.anInt(Msg.i_required_randomness).setDefault(128);
        Arg<Boolean> enableFastMode = Arg.aFlag(Msg.arg_fastmode).setDefault(true);

        Arg<Integer> slen = Arg.anInt(Msg.i_signaturekeylen);
        Arg<String> signCN = Arg.aString(Msg.i_signcn);
        Arg<BigInteger> signSN = Arg.aBigInt(Msg.i_signsn);
        Arg<String> encCN = Arg.aString(Msg.i_enccn);
        Arg<BigInteger> encSN = Arg.aBigInt(Msg.i_encsn);

        // ElGamal parameters
        Arg<BigInteger> groupP = Arg.aBigInt(Msg.i_p);
        Arg<BigInteger> groupG = Arg.aBigInt(Msg.i_g);
        Arg<Args> mod = new Arg.Tree(Msg.arg_mod, groupP, groupG).setOptional();

        Arg<String> curveName = Arg.aChoice(Msg.i_name, ECGroup.P384);
        Arg<Args> ec = new Arg.Tree(Msg.arg_ec, curveName).setOptional();

        Arg.Tree paramType = new Arg.Tree(Msg.arg_paramtype, mod, ec).setExclusive();

        // GEN protocols

        Arg<Integer> dm = Arg.anInt(Msg.arg_threshold);
        Arg<Integer> dn = Arg.anInt(Msg.arg_parties);
        Arg<Args> desmedt = new Arg.Tree(Msg.i_desmedt, dm, dn).setOptional();

        Arg.Tree genProtocol = new Arg.Tree(Msg.i_genprotocol, desmedt).setExclusive();

        public InitArgs() {
            super();
            args.add(identifier);
            args.add(outputPath);
            args.add(skipTests);
            args.add(requiredRandomness);
            args.add(enableFastMode);
            args.add(random);
            args.add(slen);
            args.add(signCN);
            args.add(signSN);
            args.add(encCN);
            args.add(encSN);
            args.add(paramType);
            args.add(genProtocol);
        }
    }
}
