package ee.ivxv.processor.util;

import ee.ivxv.common.model.LName;
import ee.ivxv.common.service.i18n.MessageException;
import ee.ivxv.common.util.Util;
import ee.ivxv.processor.Msg;
import java.io.BufferedReader;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.Optional;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class DistrictsMapper {

    private static final Logger log = LoggerFactory.getLogger(DistrictsMapper.class);

    private static final String SEPARATOR = "\t";

    /** Map from district id to map from station id to district-station pair. */
    private final Map<String, Map<String, LocationPair>> mapping = new LinkedHashMap<>();

    public DistrictsMapper() {
        // Default constructor, no re-mapping
    }

    public DistrictsMapper(InputStream in) throws Exception {
        try (BufferedReader br = new BufferedReader(new InputStreamReader(in, Util.CHARSET))) {
            br.lines() //
                    .filter(s -> !s.isEmpty()) //
                    .forEach(s -> {
                        String[] r = s.split(SEPARATOR);

                        if (r.length != 8) {
                            throw new MessageException(Msg.e_dist_mapping_invalid_row, s);
                        }
                        LName fromStat = new LName(r[0], r[1]);
                        LName fromDist = new LName(r[2], r[3]);

                        LName toStat = new LName(r[4], r[5]);
                        LName toDist = new LName(r[6], r[7]);

                        mapping.computeIfAbsent(fromDist.getId(), x -> new LinkedHashMap<>())
                                .put(fromStat.getId(), new LocationPair(toDist, toStat));
                    });
        }
        mapping.forEach((d, smap) -> smap
                .forEach((s, res) -> log.info("Mapping district / station {} {} -> {} {}", d, s,
                        res.district.getId(), res.station.getId())));
    }

    public LocationPair get(String districtId, String stationId) {
        return Optional.ofNullable(mapping.get(districtId)) //
                .map(d -> d.get(stationId))
                .orElse(new LocationPair(new LName(districtId), new LName(stationId)));
    }

    public static class LocationPair {
        public final LName district;
        public final LName station;

        public LocationPair(LName district, LName station) {
            this.district = district;
            this.station = station;
        }
    }

}
