package ee.ivxv.common.util;

import ee.ivxv.common.M;
import ee.ivxv.common.model.CandidateList;
import ee.ivxv.common.model.DistrictList;
import ee.ivxv.common.service.i18n.MessageException;
import java.io.InputStream;
import java.util.HashSet;
import java.util.Set;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class CandidatesUtil {

    private static final Logger log = LoggerFactory.getLogger(CandidatesUtil.class);

    public static CandidateList readCandidates(InputStream in, DistrictList dl) {
        try {
            CandidateList cl = Json.read(in, CandidateList.class);

            validateCandidates(cl, dl);

            return cl;
        } catch (Exception e) {
            throw new MessageException(e, M.e_cand_read_error, e);
        }
    }

    static void validateCandidates(CandidateList cl, DistrictList dl) {
        Set<String> districtIds = dl.getDistricts().keySet();
        cl.getCandidates().keySet().forEach(id -> {
            if (!districtIds.contains(id)) {
                log.error("Candidate district {} is not in district list", id);
                throw new MessageException(M.e_cand_invalid_dist, id);
            }
        });

        Set<String> candidateIds = new HashSet<>();
        cl.getCandidates().values().forEach(d -> d.values().forEach(p -> p.keySet().forEach(c -> {
            if (!candidateIds.add(c)) {
                log.error("Duplicate candidate id {} in candidate list", c);
                throw new MessageException(M.e_cand_duplicate_id, c);
            }
        })));
    }
}
