package ee.ivxv.processor.tool;

import ee.ivxv.common.M;
import ee.ivxv.common.cli.Arg;
import ee.ivxv.common.cli.Arg.TreeList;
import ee.ivxv.common.cli.Args;
import ee.ivxv.common.cli.Tool;
import ee.ivxv.common.crypto.CryptoUtil.PublicKeyHolder;
import ee.ivxv.common.model.BallotBox;
import ee.ivxv.common.model.District;
import ee.ivxv.common.model.DistrictList;
import ee.ivxv.common.model.LName;
import ee.ivxv.common.model.Voter;
import ee.ivxv.common.model.VoterList;
import ee.ivxv.common.service.bbox.BboxHelper;
import ee.ivxv.common.service.bbox.BboxHelper.VoterProvider;
import ee.ivxv.common.service.bbox.InvalidBboxException;
import ee.ivxv.common.service.bbox.Result;
import ee.ivxv.common.service.i18n.MessageException;
import ee.ivxv.common.util.I18nConsole;
import ee.ivxv.common.util.ToolHelper;
import ee.ivxv.common.util.log.PerformanceLog;
import ee.ivxv.processor.Msg;
import ee.ivxv.processor.ProcessorContext;
import ee.ivxv.processor.tool.CheckTool.CheckArgs;
import ee.ivxv.processor.util.ReportHelper;
import ee.ivxv.processor.util.VotersUtil;
import java.nio.file.Path;
import java.time.Instant;
import java.util.List;
import java.util.Map;
import java.util.function.Supplier;
import org.bouncycastle.asn1.ASN1ObjectIdentifier;
import org.bouncycastle.asn1.pkcs.PKCSObjectIdentifiers;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class CheckTool implements Tool.Runner<CheckArgs> {

    private static final Logger log = LoggerFactory.getLogger(CheckTool.class);

    static final ASN1ObjectIdentifier TS_KEY_ALG_ID = PKCSObjectIdentifiers.rsaEncryption;
    static final ASN1ObjectIdentifier VL_KEY_ALG_ID = PKCSObjectIdentifiers.rsaEncryption;

    private static final String OUT_BB = "bb-1.json";

    private final ProcessorContext ctx;
    private final I18nConsole console;
    private final ReportHelper reporter;
    private final ToolHelper tool;

    public CheckTool(ProcessorContext ctx) {
        this.ctx = ctx;
        console = new I18nConsole(ctx.i.console, ctx.i.i18n);
        reporter = new ReportHelper(ctx, console);
        tool = new ToolHelper(console, ctx.container, ctx.bbox);
    }

    @Override
    public boolean run(CheckArgs args) throws Exception {
        DistrictList dl = tool.readJsonDistricts(args.districts.value());
        VoterProvider vp = getVoterProvider(args, dl);
        reporter.writeVlErrors(args.out.value());

        BallotBox bb = readBallotBox(args, vp, dl.getElection());
        reporter.writeBbErrors(args.out.value());

        reporter.writeLog1(args.out.value(), bb);

        tool.writeJsonBb(bb, args.out.value().resolve(OUT_BB));

        return true;
    }

    private VoterProvider getVoterProvider(CheckArgs args, DistrictList dl) {
        if (args.voterLists.isSet()) {
            VoterList vl = readVoterLists(args, dl);
            return vl::find;
        }

        // Voter lists not given - using fictive voter list that accepts all

        if (dl.getDistricts().size() != 1) {
            throw new MessageException(Msg.e_vl_fictive_single_district_and_station_required);
        }
        Map.Entry<String, District> de = dl.getDistricts().entrySet().iterator().next();
        if (de.getValue().getStations().size() != 1) {
            throw new MessageException(Msg.e_vl_fictive_single_district_and_station_required);
        }
        String stationId = de.getValue().getStations().iterator().next();

        console.println(Msg.m_vl_fictive_warning, Msg.m_vl_fictive_voter_name);

        String voterName = ctx.i.i18n.get(Msg.m_vl_fictive_voter_name);
        LName district = new LName(de.getKey());
        LName station = new LName(stationId);

        return (vid, version) -> new Voter(vid, voterName, null, district, station, null);
    }

    private VoterList readVoterLists(CheckArgs args, DistrictList dl) {
        if (!args.vlKey.isSet()) {
            throw new MessageException(Msg.e_vl_vlkey_missing);
        }

        console.println();
        console.println(Msg.m_vl_reading);
        VotersUtil.Loader loader =
                VotersUtil.getLoader(args.vlKey.value(), dl, reporter::reportVlErrors);

        console.println();
        console.println(Msg.m_read);
        // NB! Must process voter lists in certain order. Using the order of input values.
        args.voterLists.value().forEach(vl -> {
            VoterList list = loader.load(vl.path.value(), vl.signature.value());
            console.println(Msg.m_vl, list.getName());
            console.println(Msg.m_vl_type, list.getType());
            console.println(Msg.m_vl_total_added, list.getAdded().size());
            console.println(Msg.m_vl_total_removed, list.getRemoved().size());
            console.println();
        });

        return loader.getCurrent();
    }

    private BallotBox readBallotBox(CheckArgs args, VoterProvider vp, String eid) throws Exception {
        try {
            tool.checkBbChecksum(args.bb.value(), args.bbChecksum.value());
            tool.checkRegChecksum(args.rl.value(), args.rlChecksum.value());

            int tc = ctx.args.threads.value();
            BboxHelper.Loader<?> loader =
                    ctx.bbox.getLoader(args.bb.value(), console::startProgress, tc);

            PerformanceLog.log
                    .info("Starting b-box loader. Num-of threads for container service: {}; "
                            + "num-of threads for BallotBox service: {}", ctx.args.ct.value(), tc);
            BallotBox bb = load(args, vp, eid, loader);

            console.println(M.m_bb_total_checked_ballots, bb.getNumberOfBallots());

            return bb;
        } catch (InvalidBboxException e) {
            throw new MessageException(e, Msg.e_bb_read_error, e.path, e);
        }
    }

    private <T> BallotBox load(CheckArgs args, VoterProvider vp, String eid,
            BboxHelper.Loader<T> loader) {
        BboxHelper.BallotsChecked<T> bc = getCheckedBallots(args.bb.value(), vp, args.tsKey.value(),
                args.elStart.value(), loader);
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
            PublicKeyHolder tsKey, Instant elStart, BboxHelper.Loader<T> l) {
        console.println();
        console.println(M.m_bb_loading, path);
        BboxHelper.BboxLoader<T> loader =
                exec(() -> l.getBboxLoader(path, reporter::reportBbError));
        console.println(M.m_bb_loaded);
        console.println(M.m_bb_checking_type);
        // If no error has occurred so far, the file structure must be correct and type UNORGANIZED
        console.println(M.m_bb_type, BallotBox.Type.UNORGANIZED);

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

    public static class CheckArgs extends Args {

        Arg<Path> bb = Arg.aPath(Msg.arg_ballotbox, true, false);
        Arg<Path> bbChecksum = Arg.aPath(Msg.arg_ballotbox_checksum, true, false);
        Arg<Path> districts = Arg.aPath(Msg.arg_districts, true, false);
        Arg<Path> rl = Arg.aPath(Msg.arg_registrationlist, true, false);
        Arg<Path> rlChecksum = Arg.aPath(Msg.arg_registrationlist_checksum, true, false);
        Arg<PublicKeyHolder> tsKey = Arg.aPublicKey(Msg.arg_tskey, TS_KEY_ALG_ID);
        Arg<PublicKeyHolder> vlKey = Arg.aPublicKey(Msg.arg_vlkey, VL_KEY_ALG_ID).setOptional();
        Arg<List<VoterListEntry>> voterLists =
                new TreeList<>(Msg.arg_voterlists, VoterListEntry::new).setOptional();
        Arg<Instant> elStart = Arg.anInstant(Msg.arg_election_start);

        Arg<Path> out = Arg.aPath(Msg.arg_out, false, null);

        public CheckArgs() {
            args.add(bb);
            args.add(bbChecksum);
            args.add(districts);
            args.add(rl);
            args.add(rlChecksum);
            args.add(tsKey);
            args.add(vlKey);
            args.add(voterLists);
            args.add(elStart);
            args.add(out);
        }

    }

    static class VoterListEntry extends Args {
        Arg<Path> path = Arg.aPath(Msg.arg_path, true, false);
        Arg<Path> signature = Arg.aPath(Msg.arg_signature, true, false);

        VoterListEntry() {
            args.add(path);
            args.add(signature);
        }
    }

}
