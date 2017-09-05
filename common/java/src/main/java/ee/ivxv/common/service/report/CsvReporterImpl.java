package ee.ivxv.common.service.report;

import ee.ivxv.common.service.i18n.I18n;
import ee.ivxv.common.util.Util;
import java.io.BufferedWriter;
import java.io.IOException;
import java.io.UncheckedIOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.List;
import java.util.stream.Collectors;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * CsvReporterImpl implements methods regarding the output format being used, which is {@code CSV}.
 */
public class CsvReporterImpl extends DefaultReporter implements Reporter {

    static final Logger log = LoggerFactory.getLogger(CsvReporterImpl.class);

    private static final String TAB = "\t";
    private static final String LF = "\n";
    private static final String VERSION_NUMBER = "1";

    public CsvReporterImpl(I18n i18n) {
        super(i18n);
    }

    /**
     * Writes report in {@code CSV} format into the specified file. The report version number is 1.
     * 
     * <p>
     * The format is:
     * </p>
     * 
     * <pre>
     * {@literal
     * report = version-number LF election-id LF *record
     * version-number = 1*2DIGIT
     * election-id = 1*28CHAR
     * record = <tab-separated-content> LF
     * }
     * </pre>
     * 
     * @param out The output file
     * @param eid The election ID
     * @param records The list of log records to be written in
     * @param headers The list of additional headers
     * @throws UncheckedIOException
     */
    @Override
    public <T extends Record> void write(Path out, String eid, List<T> records, String... headers)
            throws UncheckedIOException {
        try {
            Util.createFile(out);
        } catch (IOException e) {
            throw new UncheckedIOException(e);
        }

        try (BufferedWriter writer = Files.newBufferedWriter(out, Util.CHARSET)) {
            writer.write(VERSION_NUMBER);
            writer.write(LF);
            writer.write(eid);
            writer.write(LF);

            if (headers != null) {
                for (String header : headers) {
                    writer.write(header);
                    writer.write(LF);
                }
            }

            for (Record r : records) {
                writer.write(format(r));
                writer.write(LF);
            }
        } catch (IOException e) {
            throw new UncheckedIOException(e);
        }
    }

    /**
     * Formats a single report record as a {@code CSV} record, using tabs as separators.
     * 
     * @param r
     * @return
     */
    @Override
    public String format(Record r) {
        return r.fields.stream().collect(Collectors.joining(TAB));
    }

}
