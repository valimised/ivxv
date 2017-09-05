package ee.ivxv.common.util.log;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * The published logger of this class is configured to log performance messages in dedicated log,
 * but reflect everything in the main log. The dedicated log file is not rolled, since it is
 * supposed to gather information over longer periods of time and the log should be rather small.
 */
public class PerformanceLog {

    public static final Logger log = LoggerFactory.getLogger(PerformanceLog.class);

}
