package ee.ivxv.common.service.bbox;

import java.nio.file.Path;

public class InvalidBboxException extends RuntimeException {

    private static final long serialVersionUID = -1888245283165329958L;

    public final Path path;

    public InvalidBboxException(Path path, Throwable e) {
        super(String.format("Invalid container: %s", path), e);
        this.path = path;
    }

}
