package ee.ivxv.common.util;

import ee.ivxv.common.M;
import ee.ivxv.common.model.DistrictList;
import ee.ivxv.common.model.LName;
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
        Set<String> stationIds = new HashSet<>();

        dl.getDistricts().forEach((d, sl) -> sl.getStations().forEach(sid -> {
            if (!stationIds.add(sid)) {
                log.error("Voting station id '{}' not unique", sid);
                throw new MessageException(M.e_dist_station_id_not_unique, sid);
            }
            LName s = new LName(sid);
            if (s.getNumber() == null || s.getNumber().isEmpty()) {
                log.error("Voting station number is null, i.e the station code {} is invalid", sid);
                throw new MessageException(M.e_dist_station_id_invalid, sid);
            }
            if (!dl.getRegions().containsKey(s.getRegionCode())) {
                log.error("Voting station '{}' region code '{}' does not conform to any region",
                        sid, s.getRegionCode());
                throw new MessageException(M.e_dist_station_region_unknown, sid, s.getRegionCode());
            }
        }));
    }

}
