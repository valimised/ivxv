package ee.ivxv.common.service.report;

import ee.ivxv.common.M;
import ee.ivxv.common.crypto.hash.HashType;
import ee.ivxv.common.model.Ballot;
import ee.ivxv.common.model.BallotBox;
import ee.ivxv.common.model.DistrictList;
import ee.ivxv.common.model.LName;
import ee.ivxv.common.model.Region;
import ee.ivxv.common.service.i18n.I18n;
import ee.ivxv.common.service.i18n.MessageException;
import ee.ivxv.common.util.Json;
import ee.ivxv.common.util.PdfDoc;
import ee.ivxv.common.util.PdfDoc.Alignment;
import java.io.OutputStream;
import java.io.UncheckedIOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.time.Instant;
import java.time.LocalDateTime;
import java.time.ZoneId;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.Base64;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.Set;
import java.util.SortedMap;
import java.util.SortedSet;
import java.util.TreeMap;
import java.util.TreeSet;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.stream.Collectors;
import java.util.stream.Stream;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * DefaultReporter implements methods that do not depend on output format, but only the
 * specification.
 */
public abstract class DefaultReporter implements Reporter {

    private static final Logger log = LoggerFactory.getLogger(DefaultReporter.class);

    private static final DateTimeFormatter TIME_FMT = DateTimeFormatter.ofPattern("yyyyMMddHHmmss");

    private final I18n i18n;

    public DefaultReporter(I18n i18n) {
        this.i18n = i18n;
    }

    @Override
    public String getCurrentTime() {
        return TIME_FMT.format(LocalDateTime.now());
    }

    @Override
    public LogNRecord newLog123Record(String voterId, Ballot b) {
        Map<String, Record> records = new LinkedHashMap<>();
        b.getVotes().forEach((qid, bytes) -> records.put(qid, newLog123Record(voterId, b, qid)));
        return new LogNRecord(records);
    }

    @Override
    public Record newLog123Record(String voterId, Ballot b, String qid) {
        String time = getCurrentTime();
        LName d = b.getDistrict();
        LName s = b.getStation();
        return new Record(time, getVoteHash(b.getVotes().get(qid)), d.getRegionCode(),
                d.getNumber(), s.getRegionCode(), s.getNumber(), voterId);
    }

    private String getVoteHash(byte[] bytes) {
        return Base64.getEncoder().encodeToString(HashType.SHA256.getFunction().digest(bytes));
    }

    @Override
    public Record newRevocationRecordForRecurrentVote(String voterId, Ballot b) {
        return newRevocationRecord(RevokeAction.RECURRENT, voterId, b, "");
    }

    @Override
    public Record newRevocationRecord(RevokeAction action, String voterId, Ballot b,
            String operator) {
        String time = getCurrentTime();
        return new Record(action.value, voterId, b.getName(), format(b.getTime()), time, operator);
    }

    @Override
    public Map<String, Path> writeLogN(Path dir, String eid, LogType type,
            Stream<LogNRecord> records) throws UncheckedIOException {
        Map<String, List<Record>> rmap = new LinkedHashMap<>();

        records.forEach(r -> r.records.forEach(
                (qid, record) -> rmap.computeIfAbsent(qid, x -> new ArrayList<>()).add(record)));

        return writeRecords(dir, eid, type, rmap);
    }

    @Override
    public Map<String, Path> writeRecords(Path dir, String eid, LogType type,
            Map<String, List<Record>> rmap) {
        Map<String, Path> paths = new TreeMap<>();
        rmap.forEach((qid, rs) -> write(paths.computeIfAbsent(qid, x -> logNName(dir, type, qid)),
                eid, rs, type.value));
        return paths;
    }

    private Path logNName(Path dir, LogType type, String qid) {
        String name = String.format("%s.%s", qid, type.name().toLowerCase())
                .replaceAll("[^a-zA-Z0-9.-]", "_");
        return dir.resolve(name);
    }

    @Override
    public void writeIVoterList(Path jsonOut, Path pdfOut, BallotBox bb, DistrictList dl)
            throws Exception {
        SortedIvlData data = new SortedIvlData(bb);

        IVoterList jsonIvl = createIVoterList(data, bb.getElection());
        Json.write(jsonIvl, jsonOut);

        if (pdfOut != null) {
            Set<StationBallots> pdfIvl = createIvlDataForPdf(dl, data);
            writeIVoterListPdf(pdfIvl, bb.getElection(), pdfOut);
        }
    }

    private String format(Instant i) {
        return TIME_FMT.format(i.atZone(ZoneId.systemDefault()));
    }

    private IVoterList createIVoterList(SortedIvlData data, String election) {
        Map<String, Map<String, Map<String, List<?>>>> onlinevoters = new LinkedHashMap<>();

        data.districts.forEach((did, s) -> s.forEach((sid, ballots) -> ballots.forEach(vb -> {
            List<Object> record = new ArrayList<>();
            record.add(vb.ballot.getName());

            if (vb.ballot.getRowNumber() != null) {
                record.add(vb.ballot.getRowNumber());
            }

            onlinevoters.computeIfAbsent(did, x -> new LinkedHashMap<>())
                    .computeIfAbsent(sid, x -> new LinkedHashMap<>()).put(vb.voterId, record);
        })));

        return new IVoterList(election, onlinevoters);
    }

    private Set<StationBallots> createIvlDataForPdf(DistrictList dl, SortedIvlData data) {
        Set<StationBallots> result = new TreeSet<>();
        Set<String> stations = data.districts.values().stream().flatMap(d -> d.keySet().stream())
                .collect(Collectors.toSet());

        dl.getDistricts().forEach((did, district) -> district.getStations().forEach(sid -> {
            LName station = new LName(sid);
            String regionName = getRegionName(dl.getRegions().get(station.getRegionCode()));
            String stationName = i18n.get(M.r_ivl_station_name, station.getNumber(), regionName);
            String districtName = i18n.get(M.r_ivl_district_name, district.getName());
            SortedSet<VoterBallot> ballots = Optional.ofNullable(data.districts.get(did))
                    .map(sb -> sb.get(sid)).orElse(new TreeSet<>());
            stations.remove(sid);
            result.add(new StationBallots(did + "|" + sid, stationName, districtName, ballots));
        }));

        if (!stations.isEmpty()) {
            throw new MessageException(M.e_dist_bb_station_missing, stations);
        }

        return result;
    }

    private String getRegionName(Region r) {
        return Stream.of(r.getCounty(), r.getParish()) // Name parts to use
                .filter(x -> x != null) // Discard nulls
                .collect(Collectors.joining(", ")); // Join by ','
    }

    private void writeIVoterListPdf(Set<StationBallots> stations, String election, Path out) throws Exception {
        try (OutputStream os = Files.newOutputStream(out); PdfDoc doc = new PdfDoc(os)) {
            AtomicBoolean guard = new AtomicBoolean();
            stations.forEach(sb -> {
                try {
                    if (guard.getAndSet(true)) {
                        doc.newPage();
                    }

                    doc.resetPageNumber();
                    doc.addText(i18n.get(M.r_ivl_description, election), -1, Alignment.LEFT);
                    doc.newLine();
                    doc.newLine();
                    doc.addTitle(sb.districtName, -1, Alignment.LEFT);
                    doc.newLine();
                    doc.addTitle(sb.stationName, -1, Alignment.LEFT);
                    doc.newLine();

                    sb.ballots.forEach(vb -> {
                        try {
                            doc.newLine();
                            doc.addText(vb.ballot.getRowNumber(), 30, Alignment.RIGHT);
                            doc.tab(40);
                            doc.addText(vb.ballot.getName(), 240, Alignment.LEFT);
                            doc.tab(250);
                            doc.addText(vb.voterId);
                        } catch (Exception e) {
                            log.error("Exception while writing voter {} ({}) data to PDF",
                                    vb.voterId, vb.ballot.getName(), e);
                            throw new RuntimeException(e);
                        }
                    });
                } catch (Exception e) {
                    log.error("Exception while writing station '{}' to PDF", sb.stationName, e);
                    throw new RuntimeException(e);
                }
            });
        }
    }

    private static class SortedIvlData {
        /**
         * Map from district id to map from station id to set of voter ballots. All sorted
         * naturally.
         */
        final SortedMap<String, SortedMap<String, SortedSet<VoterBallot>>> districts =
                new TreeMap<>();

        SortedIvlData(BallotBox bb) {
            bb.getBallots().forEach((voterId, ballots) -> ballots.getBallots().forEach(b -> {
                districts.computeIfAbsent(b.getDistrictId(), x -> new TreeMap<>())
                        .computeIfAbsent(b.getStationId(), x -> new TreeSet<>())
                        .add(new VoterBallot(voterId, b));
            }));
        }
    }

    private static class VoterBallot implements Comparable<VoterBallot> {
        final String voterId;
        final Ballot ballot;

        VoterBallot(String voterId, Ballot ballot) {
            this.voterId = voterId;
            this.ballot = ballot;
        }

        @Override
        public int compareTo(VoterBallot o) {
            // Sort first by row number (null value in the end) then by voterId
            boolean hasRn1 = ballot.getRowNumber() != null;
            boolean hasRn2 = o.ballot.getRowNumber() != null;
            // Both have row numbers
            if (hasRn1 && hasRn2) {
                if (ballot.getRowNumber().equals(o.ballot.getRowNumber())) {
                    return voterId.compareTo(o.voterId);
                }
                return ballot.getRowNumber().compareTo(o.ballot.getRowNumber());
            }
            // One has row number
            if (hasRn1 || hasRn2) {
                return hasRn1 ? -1 : 1;
            }
            // Neither have row numbers
            return voterId.compareTo(o.voterId);
        }
    }

    private static class StationBallots implements Comparable<StationBallots> {
        final String districtStationId;
        final String stationName;
        final String districtName;
        final Set<VoterBallot> ballots;

        StationBallots(String districtStationId, String stationName, String districtName, Set<VoterBallot> ballots) {
            this.districtStationId = districtStationId;
            this.stationName = stationName;
            this.districtName = districtName;
            this.ballots = ballots;
        }

        @Override
        public int compareTo(StationBallots o) {
            return districtStationId.compareTo(o.districtStationId);
        }
    }

    /**
     * <pre>
     * {
     *   "election": "TESTKOV2017",
     *   "onlinevoters": {
     *     "164.1": {
     *       "164.1": {
     *         "23074322661": [ "Nimi Nimeste", 1 ],
     *         "31633238606": [ "Nimi Nimeste", 2 ]
     *       }
     *     },
     *     "296.1": {
     *       "296.1": {
     *         "43421413240": [ "Nimi Nimeste", 1 ]
     *       },
     *       "296.2": {
     *         "17368648225": [ "Nimi Nimeste", 1 ],
     *       }
     *     }
     *   }
     * }
     * </pre>
     */
    static class IVoterList {

        private final String election;
        private final Map<String, Map<String, Map<String, List<?>>>> onlinevoters =
                new LinkedHashMap<>();

        IVoterList(String election, Map<String, Map<String, Map<String, List<?>>>> onlinevoters) {
            this.election = election;
            this.onlinevoters.putAll(onlinevoters);
        }

        public String getElection() {
            return election;
        }

        public Map<String, Map<String, Map<String, List<?>>>> getOnlinevoters() {
            return onlinevoters;
        }
    }

}
