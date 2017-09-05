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
import ee.ivxv.common.math.ModPGroup;
import ee.ivxv.common.math.ModPGroupElement;
import ee.ivxv.common.service.smartcard.Cards;
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
import java.util.List;
import javax.smartcardio.CardException;
import org.bouncycastle.asn1.ASN1Sequence;
import org.bouncycastle.asn1.x500.X500Name;
import org.bouncycastle.asn1.x509.SubjectPublicKeyInfo;
import org.bouncycastle.cert.X509CertificateHolder;
import org.bouncycastle.cert.X509v1CertificateBuilder;
import org.bouncycastle.operator.ContentSigner;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class InitTool implements Tool.Runner<InitArgs> {

    private static final Logger log = LoggerFactory.getLogger(InitTool.class);

    static final Path ENC_CERT_PATH = Paths.get("enc.pem");
    static final Path ENC_KEY_DER_PATH = Paths.get("pub.der");
    static final Path ENC_KEY_PEM_PATH = Paths.get("pub.pem");
    static final Path SIGN_CERT_PATH = Paths.get("sign.pem");
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
            if (args.desmedt.isSet()) {
                desmedtGenProtocol(args, rnd, params);
            }
        } finally {
            rnd.close();
        }
        return true;
    }

    private void desmedtGenProtocol(InitArgs args, Rnd rnd, ElGamalParameters params)
            throws IOException, ProtocolException, CardException {
        // PREPARE FOR KEYPAIR GENERATION
        ThresholdParameters tparams = new ThresholdParameters(args.dn.value(), args.dm.value());
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
                new DesmedtGeneration(cards, params, tparams, rnd, AID, DEC_SHARE_NAME);
        ElGamalPublicKey pub = new ElGamalPublicKey(gen.generateKey());

        // GENERATE SIGNATURE KEYPAIR
        console.println(Msg.m_generate_signature_key);
        ShoupGeneration shoupGen =
                new ShoupGeneration(cards, args.slen.value(), tparams, rnd, AID, SIGN_SHARE_NAME);
        RSAPublicKey rsaPub = SignatureUtil.RSA.bytesToRSAPublicKey(shoupGen.generateKey());

        // GENERATE CERTIFICATES FOR BOTH KEYPAIRS
        generateAndWriteCert(args, cards, tparams, rnd, rsaPub.getEncoded(), SIGN_CERT_PATH,
                args.issuerCN.value(), args.signCN.value(), args.signSN.value());
        generateAndWriteCert(args, cards, tparams, rnd, pub.getBytes(), ENC_CERT_PATH,
                args.issuerCN.value(), args.signCN.value(), args.signSN.value());
        console.println(Msg.m_certificates_generated, SIGN_CERT_PATH, ENC_CERT_PATH);

        writeOutEncryptionKey(pub, args.outputPath.value());
        console.println(Msg.m_keys_saved, ENC_KEY_DER_PATH, ENC_KEY_PEM_PATH);

        if (!args.skipTests.value()) {
            // TEST THAT THE GENERATED KEY SUCCESSFULLY SIGNS
            UtilTool.allTests(console, log, cards, tparams, rnd, pub, rsaPub);
        }
    }

    private void generateAndWriteCert(InitArgs args, Cards cards, ThresholdParameters tparams,
            Rnd rnd, byte[] publicKey, Path path, String issuerCN, String cn, BigInteger sn)
            throws ProtocolException, IOException {
        List<Cards> quorums = cards.getQuorumList(tparams.getThreshold());
        ShoupSigning shoupSign =
                new ShoupSigning(quorums.get(0), tparams, AID, SIGN_SHARE_NAME, rnd);
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

            ModPGroup group = new ModPGroup(p);
            ModPGroupElement groupG = new ModPGroupElement(group, g);
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

    private void writeOutEncryptionKey(ElGamalPublicKey key, Path path) throws IOException {
        Files.write(path.resolve(ENC_KEY_PEM_PATH),
                Util.encodePublicKey(key.getBytes()).getBytes());
        Files.write(path.resolve(ENC_KEY_DER_PATH), key.getBytes());
    }

    public static class InitArgs extends Args {
        Arg<String> identifier = Arg.aString(Msg.arg_identifier);
        Arg<Path> outputPath = Arg.aPath(Msg.arg_out, false, null);
        Arg<Boolean> skipTests = Arg.aFlag(Msg.i_skiptest);
        Arg<List<RandomSourceArg.RndListEntry>> random = RandomSourceArg.getArgument();
        Arg<Integer> requiredRandomness = Arg.anInt(Msg.i_required_randomness).setDefault(128);
        Arg<Boolean> enableFastMode = Arg.aFlag(Msg.i_fastmode).setDefault(true);

        Arg<Integer> slen = Arg.anInt(Msg.i_signaturekeylen);
        Arg<String> issuerCN = Arg.aString(Msg.i_issuercn);
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
            args.add(issuerCN);
            args.add(signCN);
            args.add(signSN);
            args.add(encCN);
            args.add(encSN);
            args.add(paramType);
            args.add(genProtocol);
        }
    }
}
