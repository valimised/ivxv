package ee.ivxv.common.service.container;

import java.util.Collections;
import java.util.List;

/**
 * Container is a generic signed file container.
 */
public class Container {

    private final List<DataFile> files;
    private final List<Signature> signatures;

    public Container(List<DataFile> files, List<Signature> signatures) {
        this.files = Collections.unmodifiableList(files);
        this.signatures = Collections.unmodifiableList(signatures);
    }

    /**
     * @return Returns an unmodifiable not-null list of data files.
     */
    public List<DataFile> getFiles() {
        return files;
    }

    /**
     * @return Returns an unmodifiable not-null list of signatures.
     */
    public List<Signature> getSignatures() {
        return signatures;
    }

}
