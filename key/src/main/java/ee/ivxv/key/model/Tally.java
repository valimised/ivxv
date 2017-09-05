package ee.ivxv.key.model;

import com.fasterxml.jackson.annotation.JsonProperty;
import ee.ivxv.common.model.CandidateList;
import ee.ivxv.common.model.DistrictList;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.Map;

public class Tally {
    public static final String INVALID_VOTE_ID = "invalid";
    private final String election;
    private final Map<String, Map<String, Map<String, Integer>>> byStation = new HashMap<>();

    public Tally(String election, CandidateList candidates, DistrictList districts) {
        this.election = election;
        init(candidates, districts);
    }

    private void init(CandidateList candidates, DistrictList districts) {
        districts.getDistricts().forEach((dId, d) -> {
            Map<String, Map<String, String>> dCands = candidates.getCandidates().get(dId);
            Map<String, Map<String, Integer>> dTally = new LinkedHashMap<>();
            byStation.put(dId, dTally);
            d.getStations().forEach(s -> {
                Map<String, Integer> sTally = new LinkedHashMap<>();
                dTally.put(s, sTally);
                if (dCands != null) {
                    dCands.forEach((pName, pCandMap) -> pCandMap.forEach((cId, cName) -> {
                        sTally.put(cId, 0);
                    }));
                }
                sTally.put(INVALID_VOTE_ID, 0);
            });
        });
    }

    public String getElection() {
        return election;
    }

    /**
     * @return Returns a map from district id to a map from station id to a map from candidate id to
     *         number of received votes.
     */
    @JsonProperty("bystation")
    public Map<String, Map<String, Map<String, Integer>>> getByStation() {
        return byStation;
    }

    /**
     * @return Returns a map from district id to a map from candidate id to number of received
     *         votes.
     */
    @JsonProperty("bydistrict")
    public Map<String, Map<String, Integer>> getByDistrict() {
        Map<String, Map<String, Integer>> res = new LinkedHashMap<>();
        getByStation().forEach((d, sMap) -> {
            Map<String, Integer> ccMap = res.computeIfAbsent(d, tmp -> new LinkedHashMap<>());
            sMap.forEach((s, cMap) -> cMap.forEach((c, count) -> {
                ccMap.compute(c, (cc, ccount) -> ccount == null ? count : ccount + count);
            }));

        });
        return res;
    }
}
