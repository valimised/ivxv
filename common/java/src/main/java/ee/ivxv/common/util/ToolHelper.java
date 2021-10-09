package ee.ivxv.common.util;

import ee.ivxv.common.M;
import ee.ivxv.common.model.AnonymousBallotBox;
import ee.ivxv.common.model.BallotBox;
import ee.ivxv.common.model.CandidateList;
import ee.ivxv.common.model.DistrictList;
import ee.ivxv.common.model.SkipCommand;
import ee.ivxv.common.model.IBallotBox;
import ee.ivxv.common.model.Proof;
import ee.ivxv.common.service.bbox.BboxHelper;
import ee.ivxv.common.service.container.Container;
import ee.ivxv.common.service.container.ContainerReader;
import ee.ivxv.common.service.container.DataFile;
import ee.ivxv.common.service.i18n.Message;
import ee.ivxv.common.service.i18n.MessageException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;

public class ToolHelper {

    private static final String CHECKSUM_SUFFIX = ".sha256sum";

    private final I18nConsole console;
    private final ContainerReader container;
    private final BboxHelper bbox;

    public ToolHelper(I18nConsole console, ContainerReader container, BboxHelper bbox) {
        this.console = console;
        this.container = container;
        this.bbox = bbox;
    }

    public void checkBbChecksum(Path bb, Path checksum) throws Exception {
        checkChecksum(bb, checksum, M.m_bb_arg_for_checksum);
    }

    public void checkRegChecksum(Path bb, Path checksum) throws Exception {
        checkChecksum(bb, checksum, M.m_reg_arg_for_checksum);
    }

    public void checkChecksum(Path bb, Path checksum, Enum<?> name) throws Exception {
        console.println();
        console.println(M.m_checksum_loading, name, checksum);
        container.requireContainer(checksum);
        Container c = container.read(checksum.toString());
        console.println(M.m_checksum_loaded, name);

        ContainerHelper ch = new ContainerHelper(console, c);
        DataFile file = ch.getSingleFileAndReport(new Message(M.m_checksum_arg_for_cont, name));
        byte[] sum1 = Util.toBytes(file.getStream());

        console.println(M.m_checksum_calculate, name, bb);
        byte[] sum2 = bbox.getChecksum(bb);

        if (!bbox.compareChecksum(sum1, sum2)) {
            throw new MessageException(M.e_checksum_mismatch, name, bb, checksum);
        }
        console.println(M.m_checksum_ok, name);
    }

    public BallotBox readJsonBb(Path path, BallotBox.Type requiredType) throws Exception {
        return readJsonBb(path, BallotBox.class, requiredType);
    }

    public AnonymousBallotBox readJsonAbb(Path path, BallotBox.Type requiredType) throws Exception {
        return readJsonBb(path, AnonymousBallotBox.class, requiredType);
    }

    private <T extends IBallotBox> T readJsonBb(Path path, Class<T> clazz,
            BallotBox.Type requiredType) throws Exception {
        console.println();
        console.println(M.m_bb_loading, path);
        T bb = Json.read(path, clazz);
        console.println(M.m_bb_loaded);

        console.println(M.m_bb_checking_type);
        bb.requireType(requiredType);
        console.println(M.m_bb_type, bb.getType());

        console.println(M.m_bb_numof_ballots, bb.getNumberOfBallots());

        return bb;
    }

    public void writeJsonBb(IBallotBox bb, Path out) throws Exception {
        console.println();
        console.println(M.m_bb_saving, bb.getType(), out);
        Json.write(bb, out);
        console.println(M.m_bb_saved, bb.getType());

        Path checksumOut = Paths.get(out.toString() + CHECKSUM_SUFFIX);
        console.println(M.m_bb_checksum_saving, checksumOut);
        byte[] checksum = bbox.getChecksum(out);
        Files.write(checksumOut, checksum);
        console.println(M.m_bb_checksum_saved);
    }

    public CandidateList readJsonCandidates(Path path, DistrictList dl) throws Exception {
        console.println();
        console.println(M.m_cand_loading, path);
        container.requireContainer(path);
        Container c = container.read(path.toString());
        ContainerHelper ch = new ContainerHelper(console, c);
        DataFile file = ch.getSingleFileAndReport(M.m_cand_arg_for_cont);
        CandidateList candidates = CandidatesUtil.readCandidates(file.getStream(), dl);

        console.println(M.m_cand_loaded);
        console.println(M.m_cand_count, candidates.getCount());
        console.println(M.m_election_id, candidates.getElection());

        return candidates;
    }

    public DistrictList readJsonDistricts(Path path) throws Exception {
        console.println();
        console.println(M.m_dist_loading, path);
        container.requireContainer(path);
        Container c = container.read(path.toString());
        ContainerHelper ch = new ContainerHelper(console, c);
        DataFile file = ch.getSingleFileAndReport(M.m_dist_arg_for_cont);
        DistrictList districts = DistrictsUtil.readDistricts(file.getStream());

        console.println(M.m_dist_loaded);
        console.println(M.m_dist_count, districts.getCount());
        console.println(M.m_election_id, districts.getElection());

        return districts;
    }

    public SkipCommand readSkipCommand(Path path) throws Exception {
        console.println();
        console.println(M.m_skip_cmd_loading, path);
        container.requireContainer(path);
        Container c = container.read(path.toString());
        ContainerHelper ch = new ContainerHelper(console, c);
        DataFile file = ch.getSingleFileAndReport(M.m_skip_cmd_arg_for_cont);
        SkipCommand skip = SkipCommandUtil.readSkipCommand(file.getStream());
        console.println(M.m_skip_cmd_loaded);
        return skip;
    }

    public Proof readJsonProofs(Path path) throws Exception {
        console.println();
        console.println(M.m_proof_loading, path);
        Proof proofs = Json.read(path, Proof.class);
        console.println(M.m_proof_loaded);

        console.println(M.m_proof_count, proofs.getCount());

        return proofs;
    }

}
