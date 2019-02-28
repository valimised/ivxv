package ee.ivxv.processor.tool;

import ee.ivxv.common.M;
import ee.ivxv.common.cli.Arg;
import ee.ivxv.common.cli.Arg.TreeList;
import ee.ivxv.common.cli.Args;
import ee.ivxv.common.cli.Tool;
import ee.ivxv.common.crypto.CryptoUtil.PublicKeyHolder;
import ee.ivxv.common.model.BallotBox;
import ee.ivxv.common.model.DistrictList;
import ee.ivxv.common.model.LName;
import ee.ivxv.common.model.Voter;
import ee.ivxv.common.model.VoterList;
import ee.ivxv.common.service.bbox.BboxHelper;
import ee.ivxv.common.service.bbox.BboxHelper.VoterProvider;
import ee.ivxv.common.service.bbox.InvalidBboxException;
import ee.ivxv.common.service.i18n.MessageException;
import ee.ivxv.common.util.I18nConsole;
import ee.ivxv.common.util.Json;
import ee.ivxv.common.util.ToolHelper;
import ee.ivxv.processor.Msg;
import ee.ivxv.processor.ProcessorContext;
import ee.ivxv.processor.tool.StatsTool.StatsArgs;
import ee.ivxv.processor.util.DistrictsMapper;
import ee.ivxv.processor.util.ReportHelper;
import ee.ivxv.processor.util.Statistics;
import ee.ivxv.processor.util.VotersUtil;
import java.nio.file.Path;
import java.time.Instant;
import java.time.LocalDate;
import java.util.List;
import org.bouncycastle.asn1.ASN1ObjectIdentifier;
import org.bouncycastle.asn1.pkcs.PKCSObjectIdentifiers;

public class StatsTool implements Tool.Runner<StatsArgs> {

    private static final ASN1ObjectIdentifier VL_KEY_ALG_ID = PKCSObjectIdentifiers.rsaEncryption;
    private static final LName DUMMY_LNAME = new LName("0.0");

    private static final String OUT_JSON = "stats.json";
    private static final String OUT_CSV = "stats.csv";

    private final ProcessorContext ctx;
    private final I18nConsole console;
    private final ReportHelper reporter;
    private final ToolHelper tool;

    public StatsTool(ProcessorContext ctx) {
        this.ctx = ctx;
        console = new I18nConsole(ctx.i.console, ctx.i.i18n);
        reporter = new ReportHelper(ctx, console);
        tool = new ToolHelper(console, ctx.container, ctx.bbox);
    }

    @Override
    public boolean run(StatsArgs args) throws Exception {
        // If there is no district list, then only report total statistics. Otherwise statistics are
        // reported per district.
        DistrictList dl = null;
        if (args.districts.isSet()) {
            dl = tool.readJsonDistricts(args.districts.value());
        }

        Statistics stats;
        if (args.bb.value().toString().endsWith(".json")) {
            // Do not use tool.readJsonBb, since it forces us to specify a ballot box type,
            // but we want to be able to compute statistics from any type.
            BallotBox bb = readJsonBallotBox(args.bb.value());
            stats = generateStatistics(args, dl, bb);
        } else {
            VoterProvider vp = getVoterProvider(args, dl);
            reporter.writeVlErrors(args.out.value());

            BboxHelper.IntegrityChecked<?> bb = checkBallotBox(args.bb.value());
            reporter.writeBbErrors(args.out.value());
            stats = generateStatistics(args, vp, dl, bb);
        }

        Path path = args.out.value().resolve(OUT_JSON);
        stats.writeJSON(path);
        console.println(Msg.m_stats_json_saved, path);

        path = args.out.value().resolve(OUT_CSV);
        stats.writeCSV(path);
        console.println(Msg.m_stats_csv_saved, path);

        return true;
    }

    private BallotBox readJsonBallotBox(Path path) throws Exception {
        console.println();
        console.println(M.m_bb_loading, path);
        BallotBox bb = Json.read(path, BallotBox.class);
        console.println(M.m_bb_loaded);
        console.println(M.m_bb_total_ballots, bb.getNumberOfBallots());
        return bb;
    }

    private Statistics generateStatistics(StatsArgs args, DistrictList dl, BallotBox bb) {
        console.println();
        console.println(Msg.m_stats_generating);
        Statistics stats = new Statistics(args.elDay.value(), dl);
        bb.getBallots().forEach((vid, vb) -> vb.getBallots().forEach(ballot -> {
            if (args.start.isSet() && ballot.getTime().isBefore(args.start.value())
                    || args.end.isSet() && ballot.getTime().isAfter(args.end.value())) {
                return;
            }

            // Ballot does not contain voter code so assemble new voter. Cannot reuse for
            // all ballots, since district or station may change.
            Voter v = new Voter(vid, ballot.getName(), null, ballot.getDistrict(),
                    ballot.getStation(), ballot.getRowNumber());
            stats.countVoteFrom(v);
        }));
        console.println(Msg.m_stats_generated);

        return stats;
    }

    private VoterProvider getVoterProvider(StatsArgs args, DistrictList dl) {
        if (!args.voterLists.isSet()) {
            // Use dummy VoterProvider for reporting total statistics only.
            return (voter, version) -> new Voter(voter, "", "", DUMMY_LNAME, DUMMY_LNAME, 0l);
        }

        if (!args.vlKey.isSet()) {
            throw new MessageException(Msg.e_vl_vlkey_missing);
        }

        if (dl == null) {
            throw new MessageException(Msg.e_vl_districts_missing);
        }

        console.println();
        console.println(Msg.m_vl_reading);
        VotersUtil.Loader loader = VotersUtil.getLoader(args.vlKey.value(), dl,
                new DistrictsMapper(), reporter::reportVlErrors);

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

        return loader.getCurrent()::find;
    }

    private BboxHelper.IntegrityChecked<?> checkBallotBox(Path path) {
        try {
            int tc = ctx.args.threads.value();
            BboxHelper.Loader<?> loader = ctx.bbox.getLoader(path, console::startProgress, tc);

            console.println();
            console.println(M.m_bb_loading, path);
            BboxHelper.BboxLoader<?> bbLoader = loader.getBboxLoader(path, reporter::reportBbError);
            console.println(M.m_bb_loaded);
            console.println(M.m_bb_numof_collector_ballots, bbLoader.getNumberOfValidBallots());

            console.println(M.m_bb_checking_integrity);
            BboxHelper.IntegrityChecked<?> bb = bbLoader.checkIntegrity();
            console.println(M.m_bb_data_is_integrous);
            console.println(M.m_bb_total_ballots, bb.getNumberOfValidBallots());

            return bb;
        } catch (InvalidBboxException e) {
            throw new MessageException(e, Msg.e_bb_read_error, e.path, e);
        }
    }

    private Statistics generateStatistics(StatsArgs args, VoterProvider vp, DistrictList dl,
            BboxHelper.IntegrityChecked<?> bbox) {
        console.println();
        console.println(Msg.m_stats_generating);
        Statistics stats = new Statistics(args.elDay.value(), dl);
        bbox.listVoters(args.start.value(), args.end.value(), vp, stats::countVoteFrom);
        console.println(Msg.m_stats_generated);
        console.println(Msg.m_stats_ballot_errors, reporter.countBbErrors());
        console.println(Msg.m_stats_valid_ballots, stats.getTotalCount().get());

        return stats;
    }

    public static class StatsArgs extends Args {

        Arg<Path> bb = Arg.aPath(Msg.arg_ballotbox, true, false);
        Arg<LocalDate> elDay = Arg.aLocalDate(Msg.arg_election_day);
        Arg<Instant> start = Arg.anInstant(Msg.arg_period_start).setOptional();
        Arg<Instant> end = Arg.anInstant(Msg.arg_period_end).setOptional();
        Arg<Path> districts = Arg.aPath(Msg.arg_districts, true, false).setOptional();
        Arg<PublicKeyHolder> vlKey = Arg.aPublicKey(Msg.arg_vlkey, VL_KEY_ALG_ID).setOptional();
        Arg<List<VoterListEntry>> voterLists =
                new TreeList<>(Msg.arg_voterlists, VoterListEntry::new).setOptional();

        Arg<Path> out = Arg.aPath(Msg.arg_out, false, null);

        public StatsArgs() {
            args.add(bb);
            args.add(elDay);
            args.add(start);
            args.add(end);
            args.add(districts);
            args.add(vlKey);
            args.add(voterLists);
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
