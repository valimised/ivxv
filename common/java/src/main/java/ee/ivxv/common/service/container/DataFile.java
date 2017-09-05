package ee.ivxv.common.service.container;

import java.io.InputStream;

/**
 * DataFile is interface for data files.
 */
public interface DataFile {

    String getName();

    InputStream getStream();

}
