package ee.ivxv.common.service.console;

/**
 * Channel to communicate with the user.
 */
public interface Console {

    void println();

    void println(String format, Object... args);

    String readln();

    String readPw();

    Progress startProgress(String format, long total);

    void shutdown();

}
