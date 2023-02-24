package ee.ivxv.processor.tool;

import ee.ivxv.common.M;
import ee.ivxv.common.cli.Arg;
import ee.ivxv.common.cli.Arg.TreeList;
import ee.ivxv.common.cli.Args;
import ee.ivxv.common.cli.Tool;
import ee.ivxv.common.crypto.CorrectnessUtil;
import ee.ivxv.common.crypto.CryptoUtil.PublicKeyHolder;
import ee.ivxv.common.crypto.elgamal.ElGamalPublicKey;
import ee.ivxv.common.model.*;
import ee.ivxv.common.service.bbox.BboxHelper;
import ee.ivxv.common.service.bbox.BboxHelper.RegDataRef;
import ee.ivxv.common.service.bbox.BboxHelper.VoterProvider;
import ee.ivxv.common.service.bbox.InvalidBboxException;
import ee.ivxv.common.service.bbox.Ref;
import ee.ivxv.common.service.bbox.Ref.RegRef;
import ee.ivxv.common.service.bbox.Result;
import ee.ivxv.common.service.console.Progress;
import ee.ivxv.common.service.container.Container;
import ee.ivxv.common.service.container.DataFile;
import ee.ivxv.common.service.i18n.MessageException;
import ee.ivxv.common.service.report.Reporter;
import ee.ivxv.common.util.ContainerHelper;
import ee.ivxv.common.util.I18nConsole;
import ee.ivxv.common.util.ToolHelper;
import ee.ivxv.common.util.Util;
import ee.ivxv.common.util.log.PerformanceLog;
import ee.ivxv.processor.Msg;
import ee.ivxv.processor.ProcessorContext;
import ee.ivxv.processor.util.DistrictsMapper;
import ee.ivxv.processor.util.ReportHelper;
import ee.ivxv.processor.util.VotersUtil;

import java.io.Closeable;
import java.nio.file.Path;
import java.time.Instant;
import java.util.*;
import java.util.function.Supplier;

import org.bouncycastle.asn1.ASN1ObjectIdentifier;
import org.bouncycastle.asn1.pkcs.PKCSObjectIdentifiers;
import org.bouncycastle.asn1.x9.X9ObjectIdentifiers;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Class is used to combine Check and Squash tools functionality.
 * What have been done:
 * a) removed WRITE out-1 directory
 * b) reuse input for "check" tool in "squash" tool
 * c) add key processing
 * d) WRITE out-2 as "squash" tool output
 */
public class CheckAndSquashTool implements Tool.Runner<CheckAndSquashTool.CheckAndSquashArgs> {

    private static final Logger log = LoggerFactory.getLogger(CheckAndSquashTool.class);

    static final ASN1ObjectIdentifier TS_KEY_ALG_ID = PKCSObjectIdentifiers.rsaEncryption;
    static final ASN1ObjectIdentifier VL_KEY_ALG_ID = X9ObjectIdentifiers.id_ecPublicKey;
    private static final String OUT_BB_TMPL = "bb-2.json";
    private static final String OUT_IVLJSON_TMPL = "ivoterlist.json";
    private static final String OUT_IVLPDF_TMPL = "ivoterlist.pdf";
    private static final String OUT_RR_TMPL = "revocation-report.csv";
    private final ProcessorContext ctx;
    private final I18nConsole console;
    private final ReportHelper reporter;
    private final ToolHelper tool;

    private final List<Reporter.Record> revocationRecords = new ArrayList<>();
    private final List<Reporter.LogNRecord> log2Records = new ArrayList<>();

    public CheckAndSquashTool(ProcessorContext ctx) {
        this.ctx = ctx;
        console = new I18nConsole(ctx.i.console, ctx.i.i18n);
        reporter = new ReportHelper(ctx, console);
        tool = new ToolHelper(console, ctx.container, ctx.bbox);
    }

    @Override
    public boolean run(CheckAndSquashArgs args) throws Exception {
        DistrictList dl = tool.readJsonDistricts(args.districts.value());
        VoterProvider vp = getVoterProvider(args, dl);
        reporter.writeVlErrors(args.out.value());

        boolean signed = args.bbChecksum.isSet();
        if (!signed) {
            console.println(Msg.m_bb_unsigned_skipping_output);
        }


        BallotBox bb = readBallotBox(args, vp, dl.getElection());
        reporter.writeBbErrors(args.out.value());

        reporter.writeLog1(args.out.value(), bb);

        ElGamalPublicKey pub = new ElGamalPublicKey(args.encKey.value());

        removeRecurrentVotes(bb);
        removeInvalidCiphertexts(bb, pub);

        Path OUT_IVLJSON = Util.prefixedPath(bb.getElection(), OUT_IVLJSON_TMPL);
        Path OUT_IVLPDF = Util.prefixedPath(bb.getElection(), OUT_IVLPDF_TMPL);
        Path OUT_RR = Util.prefixedPath(bb.getElection(), OUT_RR_TMPL);
        Path OUT_RR_ANONYMOUS = Util.prefixedPath(bb.getElection(), OUT_RR_TMPL + ".anonymous");
        Path OUT_BB = Util.prefixedPath(bb.getElection(), OUT_BB_TMPL);

        reporter.writeIVoterList(args.out.value().resolve(OUT_IVLJSON), args.out.value().resolve(OUT_IVLPDF), bb, dl);
        reporter.writeRevocationReport(args.out.value().resolve(OUT_RR), bb.getElection(), revocationRecords,
                Reporter.AnonymousFormatter.NOT_ANONYMOUS);
        reporter.writeRevocationReport(args.out.value().resolve(OUT_RR_ANONYMOUS), bb.getElection(), revocationRecords,
                Reporter.AnonymousFormatter.REVOCATION_REPORT_CSV);
        reporter.writeLog2(args.out.value(), bb.getElection(), log2Records);

        tool.writeJsonBb(bb, args.out.value().resolve(OUT_BB));

        return true;
    }

    private VoterProvider getVoterProvider(CheckAndSquashArgs args, DistrictList dl) throws Exception {
        if (args.voterLists.isSet()) {
            VoterList vl = readVoterLists(args, dl, getDistrictsMapper(args.distMapping.value()));
            return vl::find;
        }

        // Voter lists not given - using fictive voter list that accepts all

        if (dl.getDistricts().size() != 1) {
            throw new MessageException(Msg.e_vl_fictive_single_district_and_parish_required);
        }
        Map.Entry<String, District> de = dl.getDistricts().entrySet().iterator().next();
        if (de.getValue().getParish().size() != 1) {
            throw new MessageException(Msg.e_vl_fictive_single_district_and_parish_required);
        }
        String parish = de.getValue().getParish().iterator().next();

        console.println(Msg.m_vl_fictive_warning, Msg.m_vl_fictive_voter_name);

        String voterName = ctx.i.i18n.get(Msg.m_vl_fictive_voter_name);
        LName district = new LName(de.getKey());
        return (vid, version) -> new Voter(vid, voterName, null, parish, district);
    }

    private DistrictsMapper getDistrictsMapper(Path path) throws Exception {
        if (path == null) {
            return new DistrictsMapper();
        }
        console.println();
        console.println(Msg.m_dist_mapping_loading, path);
        ctx.container.requireContainer(path);
        Container c = ctx.container.read(path.toString());
        ContainerHelper ch = new ContainerHelper(console, c);
        DataFile file = ch.getSingleFileAndReport(Msg.m_dist_mapping_arg_for_cont);
        console.println(Msg.m_dist_mapping_loaded, path);

        return new DistrictsMapper(file.getStream());
    }

    private VoterList readVoterLists(CheckAndSquashArgs args, DistrictList dl, DistrictsMapper mapper) {
        if (!args.vlKey.isSet()) {
            throw new MessageException(Msg.e_vl_vlkey_missing);
        }

        console.println();
        console.println(Msg.m_vl_reading);
        VotersUtil.Loader loader =
                VotersUtil.getLoader(args.vlKey.value(), dl, mapper, reporter::reportVlErrors);

        console.println();
        console.println(Msg.m_read);
        // NB! Must process voter lists in certain order. Using the order of input values.
        args.voterLists.value().forEach(vl -> {
            SkipCommand skip_cmd = null;
            if (vl.skip_cmd.value() != null) {
                try {
                    skip_cmd = tool.readSkipCommand(vl.skip_cmd.value());
                } catch (Exception e) {
                    throw new MessageException(Msg.e_skip_cmd_loading);
                }
            }

            VoterList list = loader.load(vl.path.value(), vl.signature.value(),
                    vl.skip_cmd.value(), skip_cmd, args.foreignEHAK.value());
            console.println(Msg.m_vl, list.getName());
            console.println(Msg.m_vl_type, list.getChangeset());
            if (skip_cmd != null) {
                console.println(Msg.m_vl_skipped);
            }
            console.println(Msg.m_vl_total_added, list.getAdded().size());
            console.println(Msg.m_vl_total_removed, list.getRemoved().size());
            console.println();
        });

        return loader.getCurrent();
    }

    private BallotBox readBallotBox(CheckAndSquashArgs args, VoterProvider vp, String eid) throws Exception {
        try {
            if (args.bbChecksum.isSet()) {
                tool.checkBbChecksum(args.bb.value(), args.bbChecksum.value());
            }
            if (args.rl.isSet()) {
                if (!args.rlChecksum.isSet()) {
                    throw new MessageException(Msg.e_reg_checksum_missing);
                }
                tool.checkRegChecksum(args.rl.value(), args.rlChecksum.value());
            }

            int tc = ctx.args.threads.value();
            BboxHelper.Loader<?> loader =
                    ctx.bbox.getLoader(args.bb.value(), console::startProgress, tc);

            BallotBox bb = load(args, vp, eid, loader);

            console.println(M.m_bb_total_checked_ballots, bb.getNumberOfBallots());

            return bb;
        } catch (InvalidBboxException e) {
            throw new MessageException(e, Msg.e_bb_read_error, e.path, e);
        }
    }

    private <T> BallotBox load(CheckAndSquashArgs args, VoterProvider vp, String eid,
                               BboxHelper.Loader<T> loader) {
        boolean haveRegData = args.rl.isSet();
        BboxHelper.BallotsChecked<T> bc = getCheckedBallots(args.bb.value(), vp, args.tsKey.value(),
                args.elStart.value(), loader, (ref, res, va) -> {
                    // Ignore BALLOT_WITHOUT_REG_REQ if there is no registration data.
                    if (haveRegData || res != Result.BALLOT_WITHOUT_REG_REQ) {
                        reporter.reportBbError(ref, res, va);
                    }
                });

        if (!haveRegData) {
            console.println();
            console.println(Msg.m_reg_skipping_compare);

            // Even if there is no reg data, we still need to run checkRegData,
            // because we cannot skip ballotbox stages.
            console.println();
            console.println(Msg.m_bb_grouping_votes_by_voter);
            return exec(() -> bc.checkRegData(new EmptyRegDataLoaderResult<T>())).getBallotBox(eid);
        }

        BboxHelper.RegDataLoaderResult<T> rdlr = getRegData(args.rl.value(), loader);

        console.println();
        console.println(M.m_bb_compare_with_reg);
        BboxHelper.BboxLoaderResult res = exec(() -> bc.checkRegData(rdlr));

        long bwr = reporter.countBbErrors(Result.BALLOT_WITHOUT_REG_REQ);
        console.println(M.m_bb_ballot_missing_reg, bwr);
        if (bwr == 0) {
            console.println(M.m_bb_in_compliance_with_reg);
        }
        long rwb = reporter.countBbErrors(Result.REG_REQ_WITHOUT_BALLOT);
        console.println(M.m_bb_reg_missing_ballot, rwb);
        if (rwb == 0) {
            console.println(M.m_reg_in_compliance_with_bb);
        }

        return res.getBallotBox(eid);
    }

    private <T> BboxHelper.BallotsChecked<T> getCheckedBallots(Path path, VoterProvider vp,
                                                               PublicKeyHolder tsKey, Instant elStart, BboxHelper.Loader<T> l,
                                                               BboxHelper.Reporter<Ref.BbRef> reporter) {
        console.println();
        console.println(M.m_bb_loading, path);
        BboxHelper.BboxLoader<T> loader = exec(() -> l.getBboxLoader(path, reporter));
        console.println(M.m_bb_loaded);
        console.println(M.m_bb_checking_type);
        // If no error has occurred so far, the file structure must be correct and type UNORGANIZED
        console.println(M.m_bb_type, BallotBox.Type.UNORGANIZED);
        console.println(M.m_bb_numof_collector_ballots, loader.getNumberOfValidBallots());

        console.println(M.m_bb_checking_integrity);
        BboxHelper.IntegrityChecked<T> ic = exec(() -> loader.checkIntegrity());
        console.println(M.m_bb_data_is_integrous);
        console.println(M.m_bb_numof_ballots, ic.getNumberOfValidBallots());

        console.println(M.m_bb_checking_ballot_sig);
        BboxHelper.BallotsChecked<T> bc = exec(() -> ic.checkBallots(vp, tsKey, elStart));
        console.println(M.m_bb_total_ballots, ic.getNumberOfValidBallots());
        console.println(M.m_bb_numof_ballots_sig_valid, bc.getNumberOfValidBallots());
        console.println(M.m_bb_numof_ballots_sig_invalid, bc.getNumberOfInvalidBallots());
        if (bc.getNumberOfInvalidBallots() == 0) {
            console.println(M.m_bb_all_ballots_sig_valid);
        }

        return bc;
    }

    private <T> BboxHelper.RegDataLoaderResult<T> getRegData(Path path, BboxHelper.Loader<T> l) {
        console.println();
        console.println(M.m_reg_loading, path);
        BboxHelper.RegDataLoader<T> rdl =
                exec(() -> l.getRegDataLoader(path, reporter::reportRegError));
        console.println(M.m_reg_loaded);

        console.println(M.m_reg_checking_integrity);
        BboxHelper.RegDataIntegrityChecked<T> rdic = exec(() -> rdl.checkIntegrity());
        console.println(M.m_reg_data_is_integrous);
        console.println(M.m_reg_numof_records, rdic.getNumberOfValidBallots());

        return exec(() -> rdic.getRegData());
    }

    private <T extends BboxHelper.Stage> T exec(Supplier<T> task) {
        long t = System.currentTimeMillis();
        T result = task.get();

        log(result, t);

        return result;
    }

    private void log(BboxHelper.Stage stage, long startTime) {
        long t = System.currentTimeMillis() - startTime;
        String name = stage.getClass().getSimpleName();

        PerformanceLog.log.info("{} #BALLOTS: {}", name, stage.getNumberOfValidBallots());
        PerformanceLog.log.info("{}     TIME: {} ms", name, t);
        log.info("{} #Ballots: {}", name, stage.getNumberOfValidBallots());
        log.info("{} #Invalid: {}", name, stage.getNumberOfInvalidBallots());
    }

    static class EmptyRegDataLoaderResult<T> implements BboxHelper.RegDataLoaderResult<T> {
        @Override
        public int getNumberOfValidBallots() {
            return 0;
        }

        @Override
        public int getNumberOfInvalidBallots() {
            return 0;
        }

        @Override
        public void report(RegRef ref, Result res, Object... args) {
            // Ignore any errors.
        }

        @Override
        public Map<Object, RegDataRef<T>> getRegData() {
            return new LinkedHashMap<>();
        }
    }

    public static class CheckAndSquashArgs extends Args {

        Arg<Path> bb = Arg.aPath(Msg.arg_ballotbox, true, false);
        Arg<Path> bbChecksum = Arg.aPath(Msg.arg_ballotbox_checksum, true, false).setOptional();
        Arg<Path> districts = Arg.aPath(Msg.arg_districts, true, false);
        Arg<Path> rl = Arg.aPath(Msg.arg_registrationlist, true, false).setOptional();
        Arg<Path> rlChecksum =
                Arg.aPath(Msg.arg_registrationlist_checksum, true, false).setOptional();
        Arg<PublicKeyHolder> tsKey = Arg.aPublicKey(Msg.arg_tskey, TS_KEY_ALG_ID);
        Arg<PublicKeyHolder> vlKey = Arg.aPublicKey(Msg.arg_vlkey, VL_KEY_ALG_ID).setOptional();
        Arg<List<VoterListEntry>> voterLists =
                new TreeList<>(Msg.arg_voterlists, VoterListEntry::new).setOptional();
        Arg<Path> distMapping = Arg.aPath(Msg.arg_districts_mapping, true, false).setOptional();
        Arg<Instant> elStart = Arg.anInstant(Msg.arg_election_start);
        Arg<String> foreignEHAK = Arg.aString(Msg.arg_voterforeignehak).setOptional();
        Arg<Path> encKey = Arg.aPath(Msg.arg_enckey, true, false);
        Arg<Path> out = Arg.aPath(Msg.arg_out, false, null);

        public CheckAndSquashArgs() {
            args.add(bb);
            args.add(bbChecksum);
            args.add(districts);
            args.add(rl);
            args.add(rlChecksum);
            args.add(tsKey);
            args.add(vlKey);
            args.add(voterLists);
            args.add(distMapping);
            args.add(elStart);
            args.add(foreignEHAK);
            args.add(encKey);
            args.add(out);
        }

    }

    static class VoterListEntry extends Args {
        Arg<Path> path = Arg.aPath(Msg.arg_path, true, false);
        Arg<Path> signature = Arg.aPath(Msg.arg_signature, true, false);
        Arg<Path> skip_cmd = Arg.aPath(Msg.arg_skip_cmd, true, false).setOptional();

        VoterListEntry() {
            args.add(path);
            args.add(signature);
            args.add(skip_cmd);
        }
    }

    private void removeRecurrentVotes(BallotBox bb) {
        console.println();
        console.println(Msg.m_removing_recurrent_votes);

        bb.removeRecurrentVotes(this::collect);

        console.println();
        console.println(M.m_bb_type, bb.getType());
        console.println(M.m_bb_numof_ballots, bb.getNumberOfBallots());
    }

    private void removeInvalidCiphertexts(BallotBox bb, ElGamalPublicKey pk) {
        console.println();
        console.println(Msg.m_removing_invalid_ciphertexts);

        try (CiphertextFilter filter = new CiphertextFilter(pk, bb.getNumberOfBallots())) {
            bb.removeInvalidCiphertexts(filter::accept);
        }
        console.println();
        console.println(M.m_bb_type, bb.getType());
        console.println(M.m_bb_numof_ballots, bb.getNumberOfBallots());
    }

    private void collect(String vid, Ballot b) {
        revocationRecords.add(ctx.reporter.newRevocationRecordForRecurrentVote(vid, b));
        log2Records.add(ctx.reporter.newLog123Record(vid, b));
    }

    private void collectinvalid(String vid, Ballot b) {
        revocationRecords.add(ctx.reporter.newRevocationRecordForInvalidVote(vid, b));
        log2Records.add(ctx.reporter.newLog123Record(vid, b));
    }

    private class CiphertextFilter implements BallotBox.VoteFilter, Closeable {

        final ElGamalPublicKey pk;
        final Progress p;

        public CiphertextFilter(ElGamalPublicKey pk, int total) {
            this.pk = pk;
            this.p = console.startProgress(total);
        }

        @Override
        public boolean accept(String voterId, Ballot b, String qid, byte[] vote) {
            CorrectnessUtil.CiphertextCorrectness res = CorrectnessUtil.isValidCiphertext(pk, vote);
            boolean isValid = res == CorrectnessUtil.CiphertextCorrectness.VALID;

            p.increase(1);

            if (!isValid) {
                collectinvalid(voterId, b);
            }

            return isValid;
        }

        @Override
        public void close() {
            p.finish();
        }

    }
}
