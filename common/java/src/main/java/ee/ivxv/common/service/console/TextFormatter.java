package ee.ivxv.common.service.console;

import java.util.Collections;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class TextFormatter {

    private static final Logger log = LoggerFactory.getLogger(TextFormatter.class);

    private static final String TAG_START = "[[";
    private static final String TAG_END = "]]";
    /** Absolute tabulation: col:2 + col:4 = column 4. */
    private static final String TAG_COL = "col";
    /** Cumulative tabulation: tab:2 + tab:2 = column 4. */
    private static final String TAG_TAB = "tab";

    public String formatLine(String line) {
        if (line == null || line.isEmpty()) {
            return line;
        }
        try {
            StringBuffer result = new StringBuffer();
            int i = 0;
            int lastTab = 0;

            while (true) {
                int j = line.indexOf(TAG_START, i);
                if (j == -1) {
                    break;
                }
                int k = line.indexOf(TAG_END, j);
                if (k == -1) {
                    throw new RuntimeException("Invalid format: tag end not found");
                }

                result.append(line.substring(i, j));
                int resLen = result.length();

                int tabWidth;
                int paddingWidth;
                String tagContent = line.substring(j + TAG_START.length(), k);
                if (tagContent.startsWith(TAG_COL)) {
                    tabWidth = Integer.parseInt(tagContent.substring(TAG_COL.length()).trim());
                    paddingWidth = Math.max(0, tabWidth - resLen);
                } else if (tagContent.startsWith(TAG_TAB)) {
                    tabWidth = Integer.parseInt(tagContent.substring(TAG_TAB.length()).trim());
                    paddingWidth = Math.max(0, lastTab + tabWidth - resLen);
                } else {
                    throw new RuntimeException("Invalid format: unknown tag: " + tagContent);
                }

                String padding = String.join("", Collections.nCopies(paddingWidth, " "));
                result.append(padding);

                i = k + TAG_END.length();
                lastTab += tabWidth;
            }

            result.append(line.substring(i));

            return result.toString();
        } catch (Exception e) {
            log.warn("Exception occurred while processing tags", e);
            return line;
        }
    }

}
