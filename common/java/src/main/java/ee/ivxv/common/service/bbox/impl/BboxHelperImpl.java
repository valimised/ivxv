package ee.ivxv.common.service.bbox.impl;

import ee.ivxv.common.conf.Conf;
import ee.ivxv.common.crypto.hash.HashType;
import ee.ivxv.common.service.bbox.BboxHelper;
import ee.ivxv.common.service.bbox.InvalidBboxException;
import ee.ivxv.common.service.bbox.Ref;
import ee.ivxv.common.service.bbox.impl.TspProfile.TsProfile;
import ee.ivxv.common.service.bbox.impl.verify.OcspVerifier;
import ee.ivxv.common.service.console.Progress;
import ee.ivxv.common.service.container.ContainerReader;
import ee.ivxv.common.util.Util;
import java.nio.file.Files;
import java.nio.file.Path;
import javax.xml.bind.DatatypeConverter;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class BboxHelperImpl implements BboxHelper {

    private static final Logger log = LoggerFactory.getLogger(BboxHelperImpl.class);

    private final OcspVerifier ocspVerifier;
    private final ContainerReader container;

    public BboxHelperImpl(Conf conf, ContainerReader container) {
        ocspVerifier = new OcspVerifier(conf.getCaCerts(), conf.getOcspCerts());
        this.container = container;
    }

    @Override
    public Loader<?> getLoader(Path path, Progress.Factory pf, int nThreads) {
        // TODO Choose profile better!!!
        Profile<?, ?, ?, ?> profile = new TsProfile(container, ocspVerifier);
        return new LoaderImpl<>(profile, pf, nThreads);
    }

    @Override
    public byte[] getChecksum(Path path) throws Exception {
        byte[] bytes = HashType.SHA256.getFunction().digest(Files.newInputStream(path));
        String checksum = DatatypeConverter.printHexBinary(bytes).toLowerCase();
        return Util.toBytes(checksum);
    }

    @Override
    public boolean compareChecksum(byte[] sum1, byte[] sum2) {
        String str1 = Util.toString(sum1).trim();
        String str2 = Util.toString(sum2).trim();
        boolean result = str1.equalsIgnoreCase(str2);
        log.debug("Comparing checksum1 {}", str1);
        log.debug("Comparing checksum2 {}", str2);
        log.debug("Result: {}", result);
        return result;
    }

    static class LoaderImpl<T extends Record<?>, U extends Record<?>, RT extends Keyable, RU extends Keyable>
            implements Loader<RU> {

        private final Profile<T, U, RT, RU> profile;
        private final Progress.Factory pf;
        private final int nThreads;

        LoaderImpl(Profile<T, U, RT, RU> profile, Progress.Factory pf, int nThreads) {
            this.profile = profile;
            this.pf = pf;
            this.nThreads = nThreads;
        }

        @Override
        public BboxLoader<RU> getBboxLoader(Path path, Reporter<Ref.BbRef> r)
                throws InvalidBboxException {
            return new IvxvBboxLoader<>(profile, new ZipSource(path), pf, r, nThreads);
        }

        @Override
        public RegDataLoader<RU> getRegDataLoader(Path path, Reporter<Ref.RegRef> r)
                throws InvalidBboxException {
            return new IvxvRegDataLoader<>(profile, new ZipSource(path), pf, r);
        }
    }

}
