package ee.ivxv.processor.util;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonPropertyOrder;
import ee.ivxv.common.model.DistrictList;
import ee.ivxv.common.model.Voter;
import ee.ivxv.common.service.i18n.MessageException;
import ee.ivxv.common.util.Json;
import ee.ivxv.processor.Msg;
import java.io.IOException;
import java.io.OutputStreamWriter;
import java.io.Writer;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.time.LocalDate;
import java.time.Period;
import java.time.format.DateTimeFormatter;
import java.time.format.DateTimeParseException;
import java.time.format.ResolverStyle;
import java.util.Arrays;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Map;
import java.util.Optional;
import java.util.Set;
import java.util.SortedMap;
import java.util.TreeMap;
import java.util.stream.Collectors;

@JsonIgnoreProperties({"meta"})
public class Statistics {

    public static final String TOTAL_DISTRICT = "TOTAL";

    private static final DateTimeFormatter ESTONIAN_DOB_FMT =
            DateTimeFormatter.ofPattern("uuuuMMdd").withResolverStyle(ResolverStyle.STRICT);
    private static final String CSV_SEPARATOR = ",";
    private static final String CSV_QUOTE = "\"";

    private static final String AGE_GROUP_16_17 = "age_group_16-17";
    private static final String AGE_GROUP_18_24 = "age_group_18-24";
    private static final String AGE_GROUP_25_34 = "age_group_25-34";
    private static final String AGE_GROUP_35_44 = "age_group_35-44";
    private static final String AGE_GROUP_45_54 = "age_group_45-54";
    private static final String AGE_GROUP_55_64 = "age_group_55-64";
    private static final String AGE_GROUP_65_74 = "age_group_65-74";
    private static final String AGE_GROUP_75PLUS = "age_group_75plus";
    private static final String REVOTERS_2_TIMES = "revoters-2-times";
    private static final String REVOTERS_3_TIMES = "revoters-3-times";
    private static final String REVOTERS_MORE_THAN_3_TIMES = "revoters-more-than-3-times";
    private static final String REVOTERS_TOTAL = "revoters-total";
    private static final String TOTAL_VOTERS = "total-voters";
    private static final String TOTAL_VOTES_COLLECTED = "total-votes-collected";
    private static final String VOTERS_FEMALES = "voters-females";
    private static final String VOTERS_MALES = "voters-males";

    private final boolean readOnly;
    private final LocalDate ageDate;
    private final SortedMap<String, Statistics.Block> districts;
    private final boolean withDistricts;

    public Statistics(LocalDate ageDate, DistrictList dl) {
        this.readOnly = false;
        this.ageDate = ageDate;
        this.districts = new TreeMap<>();
        this.districts.put(TOTAL_DISTRICT, new Block());
        this.withDistricts = dl != null;
        if (withDistricts) {
            dl.getDistricts().keySet().forEach(district -> districts.put(district, new Block()));
        }
    }

    /**
     * Statistics created via this constructor are read-only and calls to countVoteFrom are no-ops.
     */
    @JsonCreator
    public Statistics(@JsonProperty("data") Map<String, Statistics.Block> districts) {
        this.readOnly = true;
        this.ageDate = null; // Unused in read-only mode.
        this.districts = new TreeMap<>(districts);
        this.withDistricts = false; // Unused in read-only mode.
    }

    /**
     * Constructs a new read-only Statistics which represents the difference between two Statistics.
     * <p>
     * Each district is the result of subtracting values of the <tt>to</tt> Block from the values of
     * the corresponding <tt>compare</tt> Block. If a district is missing from either parameter,
     * then it is counted as a Block with all zero values.
     * 
     * @param compare the base for the comparison
     * @param to the Statistics to compare to
     */
    public Statistics(Statistics compare, Statistics to) {
        this.readOnly = true;
        this.ageDate = null; // Unused in read-only mode.
        this.districts = new TreeMap<>();
        this.withDistricts = false; // Unused in read-only mode.

        Set<String> allDistricts = new HashSet<>(compare.districts.keySet());
        allDistricts.addAll(to.districts.keySet());

        for (String district : allDistricts) {
            districts.put(district, new Block(//
                    compare.districts.getOrDefault(district, new Block()),
                    to.districts.getOrDefault(district, new Block())));
        }
    }

    public void countVoteFrom(Voter voter) {
        if (readOnly) {
            return;
        }
        String code = voter.getCode();
        int age = getAgeFromEstonianCode(code, ageDate);
        boolean female = isFemaleFromEstonianCode(code);
        districts.get(TOTAL_DISTRICT).countVoteFrom(code, age, female);
        if (withDistricts) {
            districts.get(voter.getDistrict().getId()).countVoteFrom(code, age, female);
        }
    }

    /**
     * @param code Estonian national person number
     * @param date
     * @return the age of the person relative to date
     */
    private static int getAgeFromEstonianCode(String code, LocalDate date) {
        if (code.length() != 11) {
            throw new MessageException(Msg.e_stats_code_not_estonian, code);
        }
        String yearPrefix;
        switch (code.codePointAt(0)) {
            case '1':
            case '2':
                yearPrefix = "18";
                break;
            case '3':
            case '4':
                yearPrefix = "19";
                break;
            case '5':
            case '6':
                yearPrefix = "20";
                break;
            case '7':
            case '8':
                yearPrefix = "21";
                break;
            default:
                throw new MessageException(Msg.e_stats_code_not_estonian, code);
        }
        try {
            LocalDate dob = LocalDate.parse(yearPrefix + code.substring(1, 7), ESTONIAN_DOB_FMT);
            return Period.between(dob, date).getYears();
        } catch (DateTimeParseException e) {
            throw new MessageException(e, Msg.e_stats_code_not_estonian, code);
        }
    }

    /**
     * @param code Estonian national person number
     * @return if the person is female
     */
    private static boolean isFemaleFromEstonianCode(String code) {
        if (code.length() != 11) {
            throw new MessageException(Msg.e_stats_code_not_estonian, code);
        }
        switch (code.codePointAt(0)) {
            case '1':
            case '3':
            case '5':
            case '7':
                return false;
            case '2':
            case '4':
            case '6':
            case '8':
                return true;
            default:
                throw new MessageException(Msg.e_stats_code_not_estonian, code);
        }
    }

    /**
     * @return number of total votes collected, or empty Optional if no "TOTAL" district.
     */
    @JsonIgnore
    public Optional<Integer> getTotalCount() {
        return Optional.ofNullable(districts.get(TOTAL_DISTRICT))
                .map(Block::getTotalVotesCollected);
    }

    @JsonProperty("data")
    public Map<String, Statistics.Block> getDistricts() {
        return Collections.unmodifiableMap(districts);
    }

    public void writeJSON(Path path) throws Exception {
        Json.write(this, path);
    }

    public void writeCSV(Path path) throws Exception {
        if (path.getParent() != null) {
            Files.createDirectories(path.getParent());
        }
        try (Writer writer =
                new OutputStreamWriter(Files.newOutputStream(path), StandardCharsets.UTF_8)) {
            writeCSVHeader(writer);
            for (String district : districts.keySet()) {
                writeCSVRecord(writer, district, districts.get(district));
            }
        }
    }

    private static void writeCSVHeader(Writer writer) throws IOException {
        writer.append(CSV_QUOTE)
                .append(String.join(CSV_QUOTE + CSV_SEPARATOR + CSV_QUOTE, "district",
                        AGE_GROUP_16_17, AGE_GROUP_18_24, AGE_GROUP_25_34, AGE_GROUP_35_44,
                        AGE_GROUP_45_54, AGE_GROUP_55_64, AGE_GROUP_65_74, AGE_GROUP_75PLUS,
                        REVOTERS_2_TIMES, REVOTERS_3_TIMES, REVOTERS_MORE_THAN_3_TIMES,
                        REVOTERS_TOTAL, TOTAL_VOTERS, TOTAL_VOTES_COLLECTED, VOTERS_FEMALES,
                        VOTERS_MALES))
                .append(CSV_QUOTE).append('\n');
    }

    private static void writeCSVRecord(Writer writer, String district, Block block)
            throws IOException {
        writer.append(CSV_QUOTE).append(district).append(CSV_QUOTE).append(CSV_SEPARATOR)
                .append(Arrays.asList(block.getAgeGroup1617(), block.getAgeGroup1824(),
                        block.getAgeGroup2534(), block.getAgeGroup3544(), block.getAgeGroup4554(),
                        block.getAgeGroup5564(), block.getAgeGroup6574(), block.getAgeGroup75Plus(),
                        block.getRevoters2Times(), block.getRevoters3Times(),
                        block.getRevotersMoreThan3Times(), block.getRevotersTotal(),
                        block.getTotalVoters(), block.getTotalVotesCollected(),
                        block.getVotersFemales(), block.getVotersMales()) //
                        .stream().map(String::valueOf).collect(Collectors.joining(CSV_SEPARATOR)))
                .append('\n');
    }

    @JsonPropertyOrder({AGE_GROUP_16_17, AGE_GROUP_18_24, AGE_GROUP_25_34, AGE_GROUP_35_44,
            AGE_GROUP_45_54, AGE_GROUP_55_64, AGE_GROUP_65_74, AGE_GROUP_75PLUS, REVOTERS_2_TIMES,
            REVOTERS_3_TIMES, REVOTERS_MORE_THAN_3_TIMES, REVOTERS_TOTAL, TOTAL_VOTERS,
            TOTAL_VOTES_COLLECTED, VOTERS_FEMALES, VOTERS_MALES})
    @JsonIgnoreProperties(ignoreUnknown = true) // Allow for Blocks with extra data.
    public static class Block {

        private final boolean readOnly;
        private final Map<String, Integer> voters;
        private final Map<Integer, Integer> ages;
        private int revoters2Times;
        private int revoters3Times;
        private int revotersMoreThan3Times;
        private int revotersTotal;
        private int totalVoters;
        private int totalVotesCollected;
        private int votersFemales;
        private int votersMales;

        private Block() {
            this.readOnly = false;
            this.voters = new HashMap<>();
            this.ages = new HashMap<>();
        }

        @JsonCreator
        public Block(//
                @JsonProperty(AGE_GROUP_16_17) int ageGroup1617,
                @JsonProperty(AGE_GROUP_18_24) int ageGroup1824,
                @JsonProperty(AGE_GROUP_25_34) int ageGroup2534,
                @JsonProperty(AGE_GROUP_35_44) int ageGroup3544,
                @JsonProperty(AGE_GROUP_45_54) int ageGroup4554,
                @JsonProperty(AGE_GROUP_55_64) int ageGroup5564,
                @JsonProperty(AGE_GROUP_65_74) int ageGroup6574,
                @JsonProperty(AGE_GROUP_75PLUS) int ageGroup75plus,
                @JsonProperty(REVOTERS_2_TIMES) int revoters2Times,
                @JsonProperty(REVOTERS_3_TIMES) int revoters3Times,
                @JsonProperty(REVOTERS_MORE_THAN_3_TIMES) int revotersMoreThan3Times,
                @JsonProperty(REVOTERS_TOTAL) int revotersTotal,
                @JsonProperty(TOTAL_VOTERS) int totalVoters,
                @JsonProperty(TOTAL_VOTES_COLLECTED) int totalVotesCollected,
                @JsonProperty(VOTERS_FEMALES) int votersFemales,
                @JsonProperty(VOTERS_MALES) int votersMales) {
            this.readOnly = true;
            this.voters = null; // Unused in read-only mode.
            this.ages = new HashMap<>();
            this.ages.put(16, ageGroup1617);
            this.ages.put(18, ageGroup1824);
            this.ages.put(25, ageGroup2534);
            this.ages.put(35, ageGroup3544);
            this.ages.put(45, ageGroup4554);
            this.ages.put(55, ageGroup5564);
            this.ages.put(65, ageGroup6574);
            this.ages.put(75, ageGroup75plus);
            this.revoters2Times = revoters2Times;
            this.revoters3Times = revoters3Times;
            this.revotersMoreThan3Times = revotersMoreThan3Times;
            this.revotersTotal = revotersTotal;
            this.totalVoters = totalVoters;
            this.totalVotesCollected = totalVotesCollected;
            this.votersFemales = votersFemales;
            this.votersMales = votersMales;
        }

        private Block(Block compare, Block to) {
            this.readOnly = true;
            this.voters = null; // Unused in read-only mode.
            this.ages = new HashMap<>(compare.ages);
            for (Map.Entry<Integer, Integer> age : to.ages.entrySet()) {
                this.ages.merge(age.getKey(), -age.getValue(), Integer::sum);
            }
            this.revoters2Times = compare.revoters2Times - to.revoters2Times;
            this.revoters3Times = compare.revoters3Times - to.revoters3Times;
            this.revotersMoreThan3Times =
                    compare.revotersMoreThan3Times - to.revotersMoreThan3Times;
            this.revotersTotal = compare.revotersTotal - to.revotersTotal;
            this.totalVoters = compare.totalVoters - to.totalVoters;
            this.totalVotesCollected = compare.totalVotesCollected - to.totalVotesCollected;
            this.votersFemales = compare.votersFemales - to.votersFemales;
            this.votersMales = compare.votersMales - to.votersMales;
        }

        private void countVoteFrom(String code, int age, boolean female) {
            if (readOnly) {
                return;
            }
            switch (voters.merge(code, 1, Integer::sum)) {
                case 1:
                    ages.merge(age, 1, Integer::sum);
                    totalVoters++;
                    if (female) {
                        votersFemales++;
                    } else {
                        votersMales++;
                    }
                    break;
                case 2:
                    revoters2Times++;
                    revotersTotal++;
                    break;
                case 3:
                    revoters2Times--;
                    revoters3Times++;
                    break;
                case 4:
                    revoters3Times--;
                    revotersMoreThan3Times++;
                    break;
            }
            totalVotesCollected++;
        }

        private int getAgeGroup(Integer min, Integer max) {
            return ages.entrySet().stream() //
                    .filter(entry -> min == null || min <= entry.getKey())
                    .filter(entry -> max == null || max >= entry.getKey())
                    .mapToInt(entry -> entry.getValue()).sum();
        }

        /*
         * Fields to include in the statistics output.
         */

        @JsonProperty(AGE_GROUP_16_17)
        public int getAgeGroup1617() {
            return getAgeGroup(16, 17);
        }

        @JsonProperty(AGE_GROUP_18_24)
        public int getAgeGroup1824() {
            return getAgeGroup(18, 24);
        }

        @JsonProperty(AGE_GROUP_25_34)
        public int getAgeGroup2534() {
            return getAgeGroup(25, 34);
        }

        @JsonProperty(AGE_GROUP_35_44)
        public int getAgeGroup3544() {
            return getAgeGroup(35, 44);
        }

        @JsonProperty(AGE_GROUP_45_54)
        public int getAgeGroup4554() {
            return getAgeGroup(45, 54);
        }

        @JsonProperty(AGE_GROUP_55_64)
        public int getAgeGroup5564() {
            return getAgeGroup(55, 64);
        }

        @JsonProperty(AGE_GROUP_65_74)
        public int getAgeGroup6574() {
            return getAgeGroup(65, 74);
        }

        @JsonProperty(AGE_GROUP_75PLUS)
        public int getAgeGroup75Plus() {
            return getAgeGroup(75, null);
        }

        @JsonProperty(REVOTERS_2_TIMES)
        public int getRevoters2Times() {
            return revoters2Times;
        }

        @JsonProperty(REVOTERS_3_TIMES)
        public int getRevoters3Times() {
            return revoters3Times;
        }

        @JsonProperty(REVOTERS_MORE_THAN_3_TIMES)
        public int getRevotersMoreThan3Times() {
            return revotersMoreThan3Times;
        }

        @JsonProperty(REVOTERS_TOTAL)
        public int getRevotersTotal() {
            return revotersTotal;
        }

        @JsonProperty(TOTAL_VOTERS)
        public int getTotalVoters() {
            return totalVoters;
        }

        @JsonProperty(TOTAL_VOTES_COLLECTED)
        public int getTotalVotesCollected() {
            return totalVotesCollected;
        }

        @JsonProperty(VOTERS_FEMALES)
        public int getVotersFemales() {
            return votersFemales;
        }

        @JsonProperty(VOTERS_MALES)
        public int getVotersMales() {
            return votersMales;
        }

    }

}
