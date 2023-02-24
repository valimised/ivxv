package ee.ivxv.common.service.report;

import ee.ivxv.common.model.Ballot;
import ee.ivxv.common.model.BallotBox;
import ee.ivxv.common.model.DistrictList;
import java.io.UncheckedIOException;
import java.nio.file.Path;
import java.util.Arrays;
import java.util.List;
import java.util.Map;
import java.util.stream.Stream;

public interface Reporter {

    /**
     * @return Returns the properly formatted current time for various types of log records.
     */
    String getCurrentTime();

    /**
     * Creates and returns new record for {@code LOG1..LOG3} log.
     *
     * @param voterId
     * @param b
     * @return
     */
    LogNRecord newLog123Record(String voterId, Ballot b);

    /**
     * Creates and returns new record for {@code LOG1..LOG3} log.
     *
     * @param voterId
     * @param b
     * @param qid The question id
     * @return
     */
    Record newLog123Record(String voterId, Ballot b, String qid);

    /**
     * Creates and returns new record of revocation report for recurrent vote.
     *
     * @param voterId
     * @param b
     * @return
     */
    Record newRevocationRecordForRecurrentVote(String voterId, Ballot b);

    /**
     * Creates and returns new record of revocation report for invalid vote.
     *
     * @param voterId
     * @param b
     * @return
     */
    Record newRevocationRecordForInvalidVote(String voterId, Ballot b);

    /**
     * Creates and returns new record for revocation report.
     *
     * @param action
     * @param voterId
     * @param b
     * @param operator
     * @return
     */
    Record newRevocationRecord(RevokeAction action, String voterId, Ballot b, String operator);

    /**
     * Writes the log records of the specified type into files in the specified directory and
     * returns the map from question id to the corresponding log file path.
     *
     * @param dir
     * @param eid
     * @param type
     * @param records
     * @return The map from question id to the corresponding log file path used.
     * @throws UncheckedIOException
     */
    Map<String, Path> writeLogN(Path dir, String eid, LogType type, Stream<LogNRecord> records)
            throws UncheckedIOException;

    /**
     * Writes the log records grouped by question id into files in the specified directory and
     * returns the map from question id to the corresponding log file path.
     *
     * @param dir
     * @param eid
     * @param type
     * @param rmap
     * @return
     */
    Map<String, Path> writeRecords(Path dir, String eid, LogType type,
            Map<String, List<Record>> rmap);

    /**
     * Writes the i-voter list of the specified ballot box into the specified file.
     *
     * @param jsonOut
     * @param pdfOut
     * @param bb
     * @throws Exception
     */
    void writeIVoterList(Path jsonOut, Path pdfOut, BallotBox bb, DistrictList dl) throws Exception;

    /**
     * Writes the report on the specified path.
     *
     * @param out
     * @param eid
     * @param records
     * @param headers Additional headers
     * @throws UncheckedIOException
     */
    <T extends Record> void write(Path out, String eid, List<T> records, AnonymousFormatter formatter, String... headers)
            throws UncheckedIOException;

    /**
     * Formats single report record according to the current implementation rules.
     *
     * @param r
     * @return
     */
    String format(Record r,AnonymousFormatter formatter);

    /**
     * Gives information about anonymization type of the output files
     */
    enum AnonymousFormatter {
        NOT_ANONYMOUS,
        REVOCATION_REPORT_CSV,
    }
    /**
     * Generic report record - just a list of strings.
     */
    class Record {

        final List<String> fields;

        public Record(String... fields) {
            this.fields = Arrays.asList(fields);
        }
    }

    /**
     * Set of log 1..5 records for single ballot.
     */
    class LogNRecord {

        /** Record from question id to log record */
        final Map<String, Record> records;

        public LogNRecord(Map<String, Record> records) {
            this.records = records;
        }
    }

    enum RevokeAction {
        RECURRENT("korduv e-hääl"), RESTORED("ennistatud"), REVOKED("tühistatud"), INVALID("kehtetu sedel");

        final String value;

        private RevokeAction(String value) {
            this.value = value;
        }
    }

    enum LogType {
        LOG1("1"), LOG2("2"), LOG3("3");

        public final String value;

        private LogType(String value) {
            this.value = value;
        }
    }

}
