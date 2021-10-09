package ee.ivxv.key.model;

import com.fasterxml.jackson.annotation.JsonProperty;
import ee.ivxv.common.model.CandidateList;
import ee.ivxv.common.model.DistrictList;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.Map;

/**
 * JSON serializable structure for holding the tally of the votes.
 */
public class Tally {
    public static final String INVALID_VOTE_ID = "invalid";
    private final String election;
    private final Map<String, Map<String, Map<String, Integer>>> byParish = new HashMap<>();

    /**
     * Initialize using values.
     *
     * @param election Election identifier.
     * @param candidates List of candidates.
     * @param districts List of districts.
     */
    public Tally(String election, CandidateList candidates, DistrictList districts) {
        this.election = election;
        init(candidates, districts);
    }

    private void init(CandidateList candidates, DistrictList districts) {
        districts.getDistricts().forEach((dId, d) -> {
            Map<String, Map<String, String>> dCands = candidates.getCandidates().get(dId);
            Map<String, Map<String, Integer>> dTally = new LinkedHashMap<>();
            byParish.put(dId, dTally);
            d.getParish().forEach(p -> {
                Map<String, Integer> pTally = new LinkedHashMap<>();
                dTally.put(p, pTally);
                if (dCands != null) {
                    dCands.forEach((pName, pCandMap) -> pCandMap.forEach((cId, cName) -> {
                        pTally.put(cId, 0);
                    }));
                }
                pTally.put(INVALID_VOTE_ID, 0);
            });
        });
    }

    /**
     * Get the election identifier.
     *
     * @return
     */
    public String getElection() {
        return election;
    }

    /**
     * @return Returns a map from district id to a map from station id to a map from candidate id to
     *         number of received votes.
     */
    @JsonProperty("byparish")
    public Map<String, Map<String, Map<String, Integer>>> getByParish() {
        return byParish;
    }

    /**
     * @return Returns a map from district id to a map from candidate id to number of received
     *         votes.
     */
    @JsonProperty("bydistrict")
    public Map<String, Map<String, Integer>> getByDistrict() {
        Map<String, Map<String, Integer>> res = new LinkedHashMap<>();
        getByParish().forEach((d, sMap) -> {
            Map<String, Integer> ccMap = res.computeIfAbsent(d, tmp -> new LinkedHashMap<>());
            sMap.forEach((s, cMap) -> cMap.forEach((c, count) -> {
                ccMap.compute(c, (cc, ccount) -> ccount == null ? count : ccount + count);
            }));

        });
        return res;
    }
}
