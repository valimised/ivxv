package ee.ivxv.common.util;

import ee.ivxv.common.M;
import ee.ivxv.common.model.DistrictList;
import ee.ivxv.common.service.i18n.MessageException;
import java.io.InputStream;
import java.util.HashSet;
import java.util.Set;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class DistrictsUtil {

    private static final Logger log = LoggerFactory.getLogger(DistrictsUtil.class);

    public static DistrictList readDistricts(InputStream in) {
        try {
            DistrictList dl = Json.read(in, DistrictList.class);

            validateDistricts(dl);

            return dl;
        } catch (Exception e) {
            throw new MessageException(M.e_dist_read_error, e);
        }
    }

    static void validateDistricts(DistrictList dl) {
        Set<String> parishIds = new HashSet<>();
        dl.getDistricts().forEach((d, sl) -> sl.getParish().forEach(pid -> {
            if (pid.equals("FOREIGN")){
                pid = "0000";
            }
            if (!parishIds.add(d + "|" + pid)) {
                log.error("Voting station id '{}' not unique", pid);
                throw new MessageException(M.e_dist_parish_id_not_unique, pid);
            }
            if (pid == null || pid.isEmpty()) {
                log.error("Voting parish {} is invalid", pid);
                throw new MessageException(M.e_dist_parish_id_invalid, pid);
            }
            if (!dl.getRegions().containsKey(pid)) {
                log.error("Voting parish '{}' does not conform to any region", pid);
                throw new MessageException(M.e_dist_parish_region_unknown, pid);
            }
        }));
    }

}
