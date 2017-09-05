package ee.ivxv.common.service.container;

public class InvalidContainerException extends RuntimeException {

    private static final long serialVersionUID = 4698298758772827643L;

    public final String path;

    public InvalidContainerException(String path) {
        super(String.format("Invalid signed container: %s", path));
        this.path = path;
    }

    public InvalidContainerException(String path, Throwable e) {
        super(String.format("Invalid signed container: %s", path), e);
        this.path = path;
    }

}
