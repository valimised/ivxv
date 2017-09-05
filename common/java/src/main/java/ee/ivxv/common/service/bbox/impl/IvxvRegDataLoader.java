package ee.ivxv.common.service.bbox.impl;

import ee.ivxv.common.service.bbox.BboxHelper.RegDataIntegrityChecked;
import ee.ivxv.common.service.bbox.BboxHelper.RegDataLoader;
import ee.ivxv.common.service.bbox.BboxHelper.RegDataLoaderResult;
import ee.ivxv.common.service.bbox.BboxHelper.RegDataRef;
import ee.ivxv.common.service.bbox.BboxHelper.Reporter;
import ee.ivxv.common.service.bbox.InvalidBboxException;
import ee.ivxv.common.service.bbox.Ref;
import ee.ivxv.common.service.bbox.Ref.RegRef;
import ee.ivxv.common.service.bbox.Result;
import ee.ivxv.common.service.console.Progress;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.function.Predicate;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class IvxvRegDataLoader<T extends Record<?>, U extends Record<?>, RT extends Keyable, RU extends Keyable>
        extends AbstractStage implements RegDataLoader<RU> {

    static final Logger log = LoggerFactory.getLogger(IvxvRegDataLoader.class);

    final Profile<T, U, RT, RU> profile;
    final LoaderHelper<RegRef> helper;

    IvxvRegDataLoader(Profile<T, U, RT, RU> profile, FileSource source, Progress.Factory pf,
            Reporter<RegRef> reporter) throws InvalidBboxException {
        this(profile, new LoaderHelper<>(source, RegRef::new, pf, reporter));
    }

    private IvxvRegDataLoader(Profile<T, U, RT, RU> profile, LoaderHelper<RegRef> helper)
            throws InvalidBboxException {
        super(helper.getAllRefs().size());
        this.profile = profile;
        this.helper = helper;
    }

    @Override
    public RegDataIntegrityChecked<RU> checkIntegrity() {
        int n = getNumberOfValidBallots();
        return new RegDataIntegrityCheckedImpl(helper.checkIntegrity(profile::createRegRecord, n),
                n);
    }

    class RegDataIntegrityCheckedImpl extends AbstractStage implements RegDataIntegrityChecked<RU> {

        private final Map<Ref.RegRef, U> voters;

        RegDataIntegrityCheckedImpl(Map<Ref.RegRef, U> voters, int oldValid) {
            super(voters.size(), oldValid);
            this.voters = voters;
        }

        @Override
        public RegDataLoaderResult<RU> getRegData() {
            Map<Object, RegDataRef<RU>> regData = new LinkedHashMap<>();
            Predicate<FileName<RegRef>> filter = name -> voters.containsKey(name.ref);

            helper.processRecords(filter, profile::createRegRecord, (name, record) -> {
                try {
                    RU request = profile.getRequest(record);
                    log.info("REG-NONCE reg-data: {} nonce: {}", name.ref.ref, request.getKey());

                    // Check request uniqueness
                    if (regData.put(request.getKey(),
                            new RegDataRef<>(name.ref, request)) != null) {
                        helper.report(name.ref, Result.REG_REQ_NOT_UNIQUE);
                        return;
                    }
                } catch (Exception e) {
                    helper.handleTechnicalError(name, e);
                }
            });

            return new RegDataLoaderResultImpl(regData, getNumberOfValidBallots());
        }

    } // class RegDataIntegrityCheckedImpl

    class RegDataLoaderResultImpl extends AbstractStage implements RegDataLoaderResult<RU> {

        private final Map<Object, RegDataRef<RU>> regData;

        RegDataLoaderResultImpl(Map<Object, RegDataRef<RU>> regData, int oldValid) {
            super(regData.size(), oldValid);
            this.regData = Collections.unmodifiableMap(regData);
        }

        @Override
        public Map<Object, RegDataRef<RU>> getRegData() {
            return regData;
        }

        @Override
        public void report(RegRef ref, Result res, Object... args) {
            helper.report(ref, res, args);
        }

    }

}
