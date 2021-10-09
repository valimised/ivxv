package ee.ivxv.processor.util;

import ee.ivxv.common.crypto.CryptoUtil.PublicKeyHolder;
import ee.ivxv.common.model.District;
import ee.ivxv.common.model.DistrictList;
import ee.ivxv.common.model.LName;
import ee.ivxv.common.model.Voter;
import ee.ivxv.common.model.VoterList;
import ee.ivxv.common.model.SkipCommand;
import ee.ivxv.common.service.i18n.MessageException;
import ee.ivxv.common.util.Util;
import ee.ivxv.processor.Msg;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.nio.file.Files;
import java.nio.file.Path;
import java.time.ZonedDateTime;
import java.time.format.DateTimeParseException;
import java.util.HashSet;
import java.util.ArrayList;
import java.util.List;
import java.util.Set;
import java.util.stream.Collectors;
import java.util.stream.Stream;

import org.bouncycastle.asn1.x9.X9ObjectIdentifiers;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class VotersUtil {

    static final Logger log = LoggerFactory.getLogger(VotersUtil.class);

    private static final String SEPARATOR = "\t";
    private static final String VERSION = "2";
    private static final String FOREIGN = "FOREIGN";
    private static final String DEFAULT_EHAK = "0000";

    public static Loader getLoader(PublicKeyHolder key, DistrictList dl, DistrictsMapper mapper,
                                   Reporter reporter) {
        return new Loader(key, dl, mapper, reporter);
    }


    static String readHeaderRow(String s) {
        if (s == null) {
            throw new MessageException(Msg.e_vl_invalid_header);
        }
        if (s.contains(SEPARATOR)) {
            String[] times = s.split(SEPARATOR);
            if (times.length != 2) {
                throw new MessageException(Msg.e_vl_invalid_header);
            }
            for (String time : times) {
                if (time.contains(" ")) {
                    time = time.replace(" ", "T");
                }
                try {
                    ZonedDateTime.parse(time);
                } catch (DateTimeParseException e) {
                    throw new MessageException(Msg.e_vl_invalid_time);
                }
            }
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

        public VoterList load(Path path, Path signature, Path skippath, SkipCommand skip, String... arg) {
            if (skippath == null) {
                verifySignature(path, signature);
                current = readVoterList(path, arg);
                validate(current);
            }
            else {
                List<Voter> voters = new ArrayList<>();

                String name = path.getFileName().toString() + " / " + skippath.getFileName().toString();

                current = new VoterList(current, name,
                        current.getVersionNumber(), skip.getElection(), skip.getChangeset(), voters);

            }
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

                if (!key.verify(vl, s, X9ObjectIdentifiers.ecdsa_with_SHA256)) {
                    throw new MessageException(Msg.e_vl_signature_error, signature);
                }
            } catch (Exception e) {
                throw new MessageException(e, Msg.e_vl_read_error, path, e);
            }
        }

        VoterList readVoterList(Path path, String... arg) {
            try (InputStream in = Files.newInputStream(path)) {
                String name = path.getFileName().toString();

                return readVoterList(name, in, arg);
            } catch (Exception e) {
                throw new MessageException(e, Msg.e_vl_read_error, path, e);
            }
        }

        /*-
        valijate-nimekiri = versiooninumber LF valimise-identifikaator LF tüüp LF period LF *valija

        versiooninumber = 1*2DIGIT
        valimise-identifikaator = 1*28CHAR
        tüüp = DIGIT
        rfc3339_from = RFC3339 time
        rfc3339_to = RFC3339 time
        period = rfc3339_from TAB rfc3339_to
         */
        VoterList readVoterList(String name, InputStream in, String... arg) throws Exception {
            try (BufferedReader br = new BufferedReader(new InputStreamReader(in, Util.CHARSET))) {
                String version = readHeaderRow(br.readLine());
                if (!version.equals(VERSION)) {
                    throw new MessageException(Msg.e_vl_invalid_version, version);
                }
                String electionId = readHeaderRow(br.readLine());
                String type = readHeaderRow(br.readLine());
                String times = readHeaderRow(br.readLine());
                List<Voter> voters =
                        br.lines().map(s -> parseVoter(s, arg)).collect(Collectors.toList());

                validateVoters(voters, name);

                VoterList vl =
                        new VoterList(current, name, version, electionId, type, voters);

                return vl;
            }
        }

        /*-
        valija = isikukood TAB nimi TAB tegevus TAB valijaringkond LF

        isikukood = 11*11DIGIT
        nimi = 1*100UTF-8-CHAR
        tegevus = "lisamine" | "kustutamine"

        valijaringkond = omavalitsuse-ehak-kood TAB ringkond
        omavalitsuse-ehak-kood = ehak-kood
        jaoskonna-number-omavalitsuses = 1*10 DIGIT

        ringkond =  ringkonna-number-omavalitsuses
        ringkonna-number-omavalitsuses = 1*10DIGIT

        ehak-kood = 1*10DIGIT

         */
        Voter parseVoter(String csv, String... arg) {
            String[] r = csv.split(SEPARATOR, -1);
            if (r.length != 5) {
                throw new MessageException(Msg.e_vl_invalid_voter_row, csv);
            }
            LName district = new LName("", "");
            String ehak = r[3];
            if (ehak.equals(FOREIGN)) {
                if (arg.length == 1 && arg[0] != null) {
                    ehak = arg[0];
                } else {
                    ehak = DEFAULT_EHAK;
                }
            }

            List<String> dists = dl.getDistricts().keySet().stream()
                    .filter(x -> x.endsWith("." + r[4]))
                    .collect(Collectors.toList());

            for (String key : dists) {
                List<String> parishes = dl.getDistricts().get(key).getParish();
                if (parishes.contains(ehak)) {
                    String[] d = key.split("\\.");
                    district = new LName(d[0], r[4]);
                    break;
                }
            }

            return new Voter(r[0], r[1], r[2], ehak, district);
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

            if (vl.getParent() != null) {
                if (Integer.parseInt(vl.getChangeset()) <= Integer.parseInt(vl.getParent().getChangeset())) {
                    log.error("Voter list {} order is wrong", vl.getName());
                    throw new MessageException(Msg.e_vl_invalid_changeset);
                }
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
         * <li>Consider that moving voter from one district/parish to another requires 1 removal
         * and 1 addition record NOT necessarily in correct order.
         * <li>The list of voters is checked in 3 rounds: districts/parish, removals, additions.
         * <li>For each record: if the district or parish is invalid, report error.
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
                } else if (!d.getParish().contains(v.getParish())) {
                    rep.report(Msg.e_vl_invalid_parish, vlName, v.getCode(), v.getName(),
                            v.getParish());
                    invalid.add(v.getCode());
                }
            });

            voters.forEach(v -> {
                if (invalid.contains(v.getCode())) {
                    return;
                }
                if (added.contains(v.getCode()) && removed.contains(v.getCode())) {
                    removed.remove(v.getCode());
                    added.remove(v.getCode());
                }
                if (v.isAddition()) {
                    if (!added.add(v.getCode())) {
                        rep.report(Msg.e_vl_voter_already_added, vlName, v.getCode(), v.getName());
                        invalid.add(v.getCode());
                    } else {
                        if (current != null && current.find(v.getCode()) != null && !removed.contains(v.getCode())) {
                            rep.report(Msg.e_vl_added_voter_exists, vlName, v.getCode(), v.getName());
                            invalid.add(v.getCode());
                        }
                    }
                } else {
                    if (!removed.add(v.getCode())) {
                        rep.report(Msg.e_vl_voter_already_removed, vlName, v.getCode(), v.getName());
                        invalid.add(v.getCode());
                    } else {
                        if ((current == null || current.find(v.getCode()) == null) && (!added.contains(v.getCode()))) {
                            rep.report(Msg.e_vl_removed_voter_missing, vlName, v.getCode(), v.getName());
                            invalid.add(v.getCode());
                        }
                    }
                }
            });

            voters.removeIf(v -> invalid.contains(v.getCode()));
        }

    } // class Loader

    public interface Reporter {
        void report(Enum<?> key, Object... args);
    }

}
