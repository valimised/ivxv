package ee.ivxv.audit.tools;

import ee.ivxv.audit.AuditContext;
import ee.ivxv.audit.Msg;
import ee.ivxv.audit.shuffle.DataParser;
import ee.ivxv.audit.shuffle.ShuffleProof;
import ee.ivxv.audit.tools.ConvertTool.ConvertArgs;
import ee.ivxv.common.cli.Arg;
import ee.ivxv.common.cli.Args;
import ee.ivxv.common.cli.Tool;
import ee.ivxv.common.crypto.Plaintext;
import ee.ivxv.common.crypto.elgamal.ElGamalCiphertext;
import ee.ivxv.common.crypto.elgamal.ElGamalPublicKey;
import ee.ivxv.common.math.Group;
import ee.ivxv.common.math.GroupElement;
import ee.ivxv.common.math.MathException;
import ee.ivxv.common.math.ProductGroup;
import ee.ivxv.common.math.ProductGroupElement;
import ee.ivxv.common.model.AnonymousBallotBox;
import ee.ivxv.common.util.I18nConsole;
import ee.ivxv.common.util.Json;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.slf4j.helpers.NOPLogger;

/**
 * ConvertTool verifies the correctness of file format conversions.
 */
public class ConvertTool implements Tool.Runner<ConvertArgs> {
    public static class ConvertArgs extends Args {
        Arg<Path> inputBbox = Arg.aPath(Msg.arg_input_bb);
        Arg<Path> outputBbox = Arg.aPath(Msg.arg_output_bb);
        Arg<Path> pubPath = Arg.aPath(Msg.arg_pub);
        Arg<Path> protPath = Arg.aPath(Msg.arg_protinfo, true, false);
        Arg<Path> proofPath = Arg.aPath(Msg.arg_proofdir, true, true);

        public ConvertArgs() {
            super();
            args.add(inputBbox);
            args.add(outputBbox);
            args.add(pubPath);
            args.add(protPath);
            args.add(proofPath);
        }
    }

    // we use static NOP logger to be able to run tests without logging
    // the logger is patched in non-test context
    private static Logger log = NOPLogger.NOP_LOGGER;
    private I18nConsole console;

    public ConvertTool(AuditContext ctx) {
        this.console = new I18nConsole(ctx.i.console, ctx.i.i18n);
        // patch the static logger. It is a single-instance class, so there are not any race
        // conditions.
        ConvertTool.log = LoggerFactory.getLogger(ConvertTool.class);
    }

    @Override
    public boolean run(ConvertArgs args) throws Exception {
        boolean ret = true;
        log.debug("Reading shuffle proof");
        ShuffleProof proof = new ShuffleProof(args.protPath.value(), args.proofPath.value());
        log.debug("Reading pre-shuffle ballot box");
        AnonymousBallotBox bb = Json.read(args.inputBbox.value(), AnonymousBallotBox.class);
        log.debug("Reading post-shuffle ballot box");
        AnonymousBallotBox sbb = Json.read(args.outputBbox.value(), AnonymousBallotBox.class);
        log.debug("Reading public key");
        ElGamalPublicKey pk = new ElGamalPublicKey(args.pubPath.value());
        if (!verifyPublickey(pk, proof)) {
            console.println(Msg.m_convert_publickey_failed);
            ret = false;
        } else {
            console.println(Msg.m_convert_publickey_succ);
        }
        if (!verifyBallotboxToBytetree(pk, bb, proof)) {
            console.println(Msg.m_convert_bb_to_bt_failed);
            ret = false;
        } else {
            console.println(Msg.m_convert_bb_to_bt_succ);
        }
        if (!verifyBytetreeToBallotbox(pk, proof, sbb)) {
            console.println(Msg.m_convert_bt_to_bb_failed);
            ret = false;
        } else {
            console.println(Msg.m_convert_bt_to_bb_succ);
        }
        return ret;
    }

    public static boolean verifyPublickey(ElGamalPublicKey pk, ShuffleProof proof) {
        log.debug("Verifying public key correct converting");
        log.debug("Getting shuffle proof public key");
        GroupElement btpk = proof.get_publickey();
        log.debug("Converting public key to bytetree format");
        GroupElement gppk = convertPublickeyToBytetree(pk);
        boolean res = btpk.equals(gppk);
        log.debug("Shuffle proof public key == converted public key: {}", res);
        return res;
    }

    public static boolean verifyBallotboxToBytetree(ElGamalPublicKey pk, AnonymousBallotBox bb,
            ShuffleProof proof) {
        log.debug("Verifying ballot box to bytetree converting");
        log.debug("Getting shuffle ciphertext list");
        GroupElement[] proofciphs = proof.get_ciphertexts();
        log.debug("Converting ballot box to byte tree format");
        GroupElement[] bbciphs;
        try {
            bbciphs = convertBallotboxToBytetree(pk, bb);
        } catch (IllegalArgumentException e) {
            log.debug("Converting ballot box to bytetree cipherexts failed: {}", e);
            return false;
        }
        if (bbciphs.length != proofciphs.length) {
            log.debug("Ballot box and shuffle ciphertext list size differ");
            return false;
        }
        for (int i = 0; i < bbciphs.length; i++) {
            if (!bbciphs[i].equals(proofciphs[i])) {
                log.debug("Ballot box element and shuffle ciphertext item '{}' differ", i);
                return false;
            }
        }
        log.debug("Shuffle proof ciphertexts == ballot box: true");
        return true;
    }

    public static boolean verifyBytetreeToBallotbox(ElGamalPublicKey pk, ShuffleProof proof,
            AnonymousBallotBox bb) {
        log.debug("Verifying bytetree to ballot box converting");
        log.debug("Converting shuffle ciphertext list to ballot box");
        AnonymousBallotBox proofbb;
        try {
            proofbb = convertBytetreeToBallotbox(pk, proof.get_shuffled_ciphertexts());
        } catch (IllegalArgumentException e) {
            log.debug("Converting bytetree to ballot box failed: {}", e);
            return false;
        }
        if (!bb.getElection().equals(proofbb.getElection())) {
            log.debug("Converted ballot box and ballot box election identifier differ. "
                    + "Expected '{}', got '{}'", bb.getElection(), proofbb.getElection());
            return false;
        }
        Map<String, Map<String, Map<String, List<byte[]>>>> proofd = proofbb.getDistricts();
        Map<String, Map<String, Map<String, List<byte[]>>>> bbd = bb.getDistricts();
        if (!proofd.keySet().equals(bbd.keySet())) {
            log.debug("Converted ballot box and ballot box district identifier set differ");
            return false;
        }
        for (String d : proofd.keySet()) {
            Map<String, Map<String, List<byte[]>>> proofs = proofd.get(d);
            Map<String, Map<String, List<byte[]>>> bbs = bbd.get(d);
            if (!proofs.keySet().equals(bbs.keySet())) {
                log.debug("Converted ballot box and ballot box station identifier set differ "
                        + "for district '{}'", d);
                return false;
            }
            for (String s : proofs.keySet()) {
                Map<String, List<byte[]>> proofq = proofs.get(s);
                Map<String, List<byte[]>> bbq = bbs.get(s);
                if (!proofq.keySet().equals(bbq.keySet())) {
                    log.debug("Converted ballot box and ballot box question identifier set differ "
                            + "for district '{}' and station '{}'", d, s);
                    return false;
                }
                for (String q : proofq.keySet()) {
                    List<byte[]> proofc = proofq.get(q);
                    List<byte[]> bbc = bbq.get(q);
                    if (proofc.size() != bbc.size()) {
                        log.debug(
                                "Converted ballot box and ballot box ciphertext list differ "
                                        + "for district '{}', station '{}' and question '{}'",
                                d, s, q);
                        return false;
                    }
                    for (int i = 0; i < proofc.size(); i++) {
                        if (!Arrays.equals(proofc.get(i), bbc.get(i))) {
                            log.debug("Converted ballot box and ballot box ciphertext differ for "
                                    + "district '{}', station '{}', question '{}' and "
                                    + "index '{}'", d, s, q, i);
                            return false;
                        }
                    }
                }
            }
        }
        return true;
    }

    private static GroupElement[] convertBallotboxToBytetree(ElGamalPublicKey pk,
            AnonymousBallotBox bb) throws IllegalArgumentException {
        List<ProductGroupElement> res = new ArrayList<ProductGroupElement>();
        ProductGroupElement ege = convertStringToProductGroupElement(pk, bb.getElection());
        bb.getDistricts().forEach((d, smap) -> {
            ProductGroupElement dge = convertStringToProductGroupElement(pk, d);
            smap.forEach((s, qmap) -> {
                ProductGroupElement sge = convertStringToProductGroupElement(pk, s);
                qmap.forEach((q, clist) -> {
                    ProductGroupElement qge = convertStringToProductGroupElement(pk, q);
                    clist.forEach(c -> {
                        ElGamalCiphertext ct = new ElGamalCiphertext(pk.getParameters(), c);
                        ProductGroupElement cge = ct.getAsProductGroupElement();
                        ProductGroupElement multict = DataParser
                                .toArray(new ProductGroupElement[] {ege, dge, sge, qge, cge});
                        res.add(multict);
                    });
                });
            });
        });
        return res.toArray(new ProductGroupElement[0]);
    }

    private static ProductGroupElement convertStringToProductGroupElement(ElGamalPublicKey pk,
            String msg) throws IllegalArgumentException {
        Group G = pk.getParameters().getGroup();
        Plaintext padded = G.pad(new Plaintext(msg));
        GroupElement ge;
        try {
            ge = G.encode(padded);
        } catch (MathException e) {
            throw new IllegalArgumentException("Encoding failed", e);
        }
        ProductGroup pG = new ProductGroup(G, 2);
        ProductGroupElement pge = new ProductGroupElement(pG, G.getIdentity(), ge);
        return pge;
    }

    private static AnonymousBallotBox convertBytetreeToBallotbox(ElGamalPublicKey pk,
            GroupElement[] cts) throws IllegalArgumentException {
        Map<String, Map<String, Map<String, List<byte[]>>>> res =
                new LinkedHashMap<String, Map<String, Map<String, List<byte[]>>>>();
        String election = null;
        for (int i = 0; i < cts.length; i++) {
            if (!(cts[i] instanceof ProductGroupElement)) {
                throw new IllegalArgumentException(
                        String.format("Ciphertext %d not ProductGroupElement", i));
            }
            ProductGroupElement multict = (ProductGroupElement) cts[i];
            if (multict.getElements().length != 2) {
                throw new IllegalArgumentException(String.format("Ciphertext %d length not 2", i));
            }
            if (!(multict.getElements()[0] instanceof ProductGroupElement)) {
                throw new IllegalArgumentException(
                        String.format("Ciphertext %d left side not PGE", i));
            }
            if (((ProductGroupElement) multict.getElements()[0]).getElements().length != 5) {
                throw new IllegalArgumentException(
                        String.format("Ciphertext %d left side length not 5", i));
            }
            if (!(multict.getElements()[1] instanceof ProductGroupElement)) {
                throw new IllegalArgumentException(
                        String.format("Ciphertext %d right side not PGE", i));
            }
            if (((ProductGroupElement) multict.getElements()[1]).getElements().length != 5) {
                throw new IllegalArgumentException(
                        String.format("Ciphertext %d right side length not 5", i));
            }
            String thiselection = convertProductGroupElementToString(pk, multict, 0);
            if (election == null) {
                election = thiselection;
            }
            if (!election.equals(thiselection)) {
                throw new IllegalArgumentException(String
                        .format("Ciphertext %d election differs. Expected '%s', got '%s'", i));
            }
            String district = convertProductGroupElementToString(pk, multict, 1);
            String station = convertProductGroupElementToString(pk, multict, 2);
            String question = convertProductGroupElementToString(pk, multict, 3);
            GroupElement blind = ((ProductGroupElement) multict.getElements()[0]).getElements()[4];
            GroupElement blindedMessage =
                    ((ProductGroupElement) multict.getElements()[1]).getElements()[4];
            ElGamalCiphertext cc =
                    new ElGamalCiphertext(blind, blindedMessage, pk.getParameters().getOID());
            res.computeIfAbsent(district, x -> new LinkedHashMap<>())
                    .computeIfAbsent(station, x -> new LinkedHashMap<>())
                    .computeIfAbsent(question, x -> new ArrayList<byte[]>()).add(cc.getBytes());
        }
        AnonymousBallotBox bb = new AnonymousBallotBox(election, res);
        return bb;
    }

    private static String convertProductGroupElementToString(ElGamalPublicKey pk,
            ProductGroupElement pge, int index) throws IllegalArgumentException {
        GroupElement[] els = pge.getElements();
        GroupElement msgenc = ((ProductGroupElement) els[1]).getElements()[index];
        Plaintext padded = pk.getParameters().getGroup().decode(msgenc);
        Plaintext pt = padded.stripPadding();
        String msg = pt.getUTF8DecodedMessage();
        return msg;
    }

    private static GroupElement convertPublickeyToBytetree(ElGamalPublicKey pk) {
        ProductGroupElement pkpg = pk.getAsProductGroupElement();
        ProductGroupElement gpg = new ProductGroupElement((ProductGroup) pkpg.getGroup(),
                pk.getParameters().getGenerator(), pk.getParameters().getGroup().getIdentity());
        ProductGroupElement res =
                DataParser.toArray(new ProductGroupElement[] {gpg, gpg, gpg, gpg, pkpg});
        return res;
    }
}
