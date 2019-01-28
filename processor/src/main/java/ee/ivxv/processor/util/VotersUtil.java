package ee.ivxv.processor.util;

import ee.ivxv.common.crypto.CryptoUtil.PublicKeyHolder;
import ee.ivxv.common.crypto.hash.HashFunction;
import ee.ivxv.common.crypto.hash.HashType;
import ee.ivxv.common.model.District;
import ee.ivxv.common.model.DistrictList;
import ee.ivxv.common.model.LName;
import ee.ivxv.common.model.Voter;
import ee.ivxv.common.model.VoterList;
import ee.ivxv.common.service.i18n.MessageException;
import ee.ivxv.common.util.Util;
import ee.ivxv.processor.Msg;
import ee.ivxv.processor.util.DistrictsMapper.LocationPair;
import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Base64;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import java.util.stream.Collectors;
import org.bouncycastle.asn1.pkcs.PKCSObjectIdentifiers;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class VotersUtil {

    static final Logger log = LoggerFactory.getLogger(VotersUtil.class);

    private static final String SEPARATOR = "\t";

    public static Loader getLoader(PublicKeyHolder key, DistrictList dl, DistrictsMapper mapper,
            Reporter reporter) {
        return new Loader(key, dl, mapper, reporter);
    }

    /**
     * Calculate the voter list hash according to the following formula:
     * 
     * <pre>
     * v_0 = ""
     * v_n = base64(sha256(v_{n-1} | base64(sha256(nk_n))))
     * </pre>
     * 
     * where {@code nk_n} is the <i>n</i>th voter list, {@code v_n} is its version number and {@code
     * |} is the string concatenation operation.
     * 
     * @param parentHash
     * @param in
     * @return
     * @throws IOException
     */
    static String getVoterListHash(String parentHash, InputStream in) throws IOException {
        StringBuffer sb = new StringBuffer();
        HashFunction sha256 = HashType.SHA256.getFunction();
        Base64.Encoder base64 = Base64.getEncoder();

        // Parent hash...
        sb.append(parentHash);

        // ... concatenated with base64-encoded sha256-digest of the current voter list ...
        sb.append(Util.toString(base64.encode(sha256.digest(in))));

        // ... and result is base64-encoded sha256-digest
        return Util.toString(base64.encode(sha256.digest(Util.toBytes(sb.toString()))));
    }

    static String readHeaderRow(String s) {
        if (s == null || s.contains(SEPARATOR)) {
            throw new MessageException(Msg.e_vl_invalid_header);
        }
        return s;
    }

    public static class Loader {

        private final PublicKeyHolder key;
        private final DistrictList dl;
        private final DistrictsMapper mapper;
        private final Reporter rep;
        private VoterList current;

        Loader(PublicKeyHolder key, DistrictList dl, DistrictsMapper mapper, Reporter rep) {
            this.key = key;
            this.dl = dl;
            this.mapper = mapper;
            this.rep = rep;
        }

        public VoterList load(Path path, Path signature) {
            verifySignature(path, signature);

            current = readVoterList(path);
            validate(current);

            return current;
        }

        public VoterList getCurrent() {
            return current;
        }

        /* For testing only. */
        void setCurrent(VoterList current) {
            this.current = current;
        }

        void verifySignature(Path path, Path signature) {
            try {
                byte[] vl = Files.readAllBytes(path);
                byte[] s = Files.readAllBytes(signature);

                if (!key.verify(vl, s, PKCSObjectIdentifiers.sha256WithRSAEncryption)) {
                    throw new MessageException(Msg.e_vl_signature_error, signature);
                }
            } catch (Exception e) {
                throw new MessageException(e, Msg.e_vl_read_error, path, e);
            }
        }

        VoterList readVoterList(Path path) {
            try (InputStream in = Files.newInputStream(path);
                    InputStream hashIn = Files.newInputStream(path)) {
                String hash = getVoterListHash(current != null ? current.getHash() : "", hashIn);
                String name = path.getFileName().toString();

                return readVoterList(hash, name, in);
            } catch (Exception e) {
                throw new MessageException(e, Msg.e_vl_read_error, path, e);
            }
        }

        /*-
        valijate-nimekiri = versiooninumber LF valimise-identifikaator LF tüüp LF *valija
        
        versiooninumber = 1*2DIGIT
        valimise-identifikaator = 1*28CHAR
        tüüp = "algne" | "muudatused"
         */
        VoterList readVoterList(String hash, String name, InputStream in) throws Exception {
            try (BufferedReader br = new BufferedReader(new InputStreamReader(in, Util.CHARSET))) {
                String version = readHeaderRow(br.readLine());
                String electionId = readHeaderRow(br.readLine());
                String type = readHeaderRow(br.readLine());
                List<Voter> voters =
                        br.lines().map(s -> parseVoter(s)).collect(Collectors.toList());

                validateVoters(voters, name);

                VoterList vl =
                        new VoterList(current, hash, name, version, electionId, type, voters);

                return vl;
            }
        }

        /*-
        valija = isikukood TAB nimi TAB tegevus TAB jaoskond TAB rea-number-voi-tyhi TAB pohjus-voi-tyhi LF
        
        isikukood = 11*11DIGIT
        nimi = 1*100UTF-8-CHAR
        tegevus = "lisamine" | "kustutamine"
        
        jaoskond = jaoskonna-ehak-kood TAB jaoskonna-number-omavalitsuses TAB ringkond
        jaoskonna-ehak-kood = ehak-kood
        jaoskonna-number-omavalitsuses = 1*10 DIGIT
        
        ringkond = ringkonna-ehak-kood TAB ringkonna-number-omavalitsuses
        ringkonna-ehak-kood = ehak-kood
        ringkonna-number-omavalitsuses = 1*10DIGIT
        
        ehak-kood = 1*10DIGIT
        
        rea-number-voi-tyhi = "" | rea-number
        rea-number = 1*11DIGIT
        pohjus-voi-tyhi = “” | pohjus
        pohjus = “tokend” | “jaoskonna vahetus” | “muu”
         */
        Voter parseVoter(String csv) {
            String[] r = csv.split(SEPARATOR, -1);

            if (r.length != 9) {
                throw new MessageException(Msg.e_vl_invalid_voter_row, csv);
            }
            LName district = new LName(r[5], r[6]);
            LName station = new LName(r[3], r[4]);
            LocationPair res = mapper.get(district.getId(), station.getId());

            Long rowNumber = null;
            try {
                rowNumber = r[7] == null || r[7].isEmpty() ? null : Long.parseLong(r[7]);
            } catch (Exception e) {
                throw new MessageException(Msg.e_vl_invalid_row_number, r[7], csv);
            }

            // Ignoring r[8] ("pohjus")
            return new Voter(r[0], r[1], r[2], res.district, res.station, rowNumber);
        }

        void validate(VoterList vl) {
            if (vl.getParent() == null && !vl.isInitial()) {
                log.error("The first voter list {} is not initial list", vl.getName());
                throw new MessageException(Msg.e_vl_first_not_initial, vl.getName());
            }
            if (vl.getParent() != null && vl.isInitial()) {
                log.error("Initial voter list {} is not the first", vl.getName());
                throw new MessageException(Msg.e_vl_initial_not_first, vl.getName());
            }

            if (vl.getElectionId() == null || !vl.getElectionId().equals(dl.getElection())) {
                throw new MessageException(Msg.e_vl_election_id, vl.getName(), vl.getElectionId(),
                        dl.getElection());
            }
        }

        /**
         * Removes invalid voters from the list and reports about errors.
         * 
         * <p>
         * The processing logic:
         * <ul>
         * <li>If there is an error with a record of a voter X, all records of voter X in this voter
         * list are removed.
         * <li>Consider that moving voter from one district/station to another requires 1 removal
         * and 1 addition record NOT necessarily in correct order.
         * <li>The list of voters is checked in 3 rounds: districts/stations, removals, additions.
         * <li>For each record: if the district or station is invalid, report error.
         * <li>For each removal record X: if there are multiple removal records for X or the voter X
         * does not have valid record, report error.
         * <li>For each addition record X: if there are multiple addition records for X or the voter
         * already has valid record, report error.
         * </ul>
         * 
         * @param voters
         * @param vlName The voter list display name for reporting.
         */
        void validateVoters(List<Voter> voters, String vlName) {
            Set<String> removed = new HashSet<>();
            Set<String> added = new HashSet<>();
            Set<String> invalid = new HashSet<>();

            voters.forEach(v -> {
                District d = dl.getDistricts().get(v.getDistrict().getId());
                if (d == null) {
                    rep.report(Msg.e_vl_invalid_district, vlName, v.getCode(), v.getName(),
                            v.getDistrict().getId());
                    invalid.add(v.getCode());
                } else if (!d.getStations().contains(v.getStation().getId())) {
                    rep.report(Msg.e_vl_invalid_station, vlName, v.getCode(), v.getName(),
                            v.getStation().getId());
                    invalid.add(v.getCode());
                }
            });

            // Process all removals before additions
            voters.forEach(v -> {
                if (v.isAddition() || invalid.contains(v.getCode())) {
                    return;
                }
                if (!removed.add(v.getCode())) {
                    rep.report(Msg.e_vl_voter_already_removed, vlName, v.getCode(), v.getName());
                    invalid.add(v.getCode());
                }
                if (current == null || current.find(v.getCode()) == null) {
                    rep.report(Msg.e_vl_removed_voter_missing, vlName, v.getCode(), v.getName());
                    invalid.add(v.getCode());
                }
            });

            voters.forEach(v -> {
                if (!v.isAddition() || invalid.contains(v.getCode())) {
                    return;
                }
                if (!added.add(v.getCode())) {
                    rep.report(Msg.e_vl_voter_already_added, vlName, v.getCode(), v.getName());
                    invalid.add(v.getCode());
                }
                if (!removed.contains(v.getCode()) && current != null
                        && current.find(v.getCode()) != null) {
                    rep.report(Msg.e_vl_added_voter_exists, vlName, v.getCode(), v.getName());
                    invalid.add(v.getCode());
                }
            });

            voters.removeIf(v -> invalid.contains(v.getCode()));
        }

    } // class Loader

    public interface Reporter {
        void report(Enum<?> key, Object... args);
    }

}
