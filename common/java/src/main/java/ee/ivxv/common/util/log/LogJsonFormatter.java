package ee.ivxv.common.util.log;

import ch.qos.logback.contrib.jackson.JacksonJsonFormatter;
import java.io.IOException;
import java.util.Map;

public class LogJsonFormatter extends JacksonJsonFormatter {

    private static final String NEWLINE = System.getProperty("line.separator", "\n");

    @Override
    @SuppressWarnings("rawtypes")
    public String toJsonString(Map arg0) throws IOException {
        return super.toJsonString(arg0) + NEWLINE;
    }
}
