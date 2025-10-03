package ee.ivxv.common.zip;

public class ZipFileCommentException extends Exception {
    public ZipFileCommentException() {
        super("invalid .ZIP file comment length");
    }
}