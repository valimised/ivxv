package ee.ivxv.common.service.bbox.impl;

import ee.ivxv.common.crypto.hash.HashType;
import ee.ivxv.common.model.Ballot;
import ee.ivxv.common.model.Voter;
import ee.ivxv.common.service.bbox.BboxHelper.VoterProvider;
import ee.ivxv.common.service.bbox.Ref;
import ee.ivxv.common.service.bbox.Result;
import ee.ivxv.common.service.bbox.impl.verify.OcspVerifier;
import ee.ivxv.common.service.bbox.impl.verify.TsVerifier;
import ee.ivxv.common.service.container.Container;
import ee.ivxv.common.service.container.ContainerReader;
import ee.ivxv.common.service.container.InvalidContainerException;
import ee.ivxv.common.service.container.Signature;
import ee.ivxv.common.service.container.Subject;
import ee.ivxv.common.util.ByteArrayWrapper;
import ee.ivxv.common.util.Util;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.Optional;
import org.apache.xml.security.c14n.Canonicalizer;
import org.bouncycastle.asn1.cms.ContentInfo;
import org.bouncycastle.tsp.TimeStampRequest;
import org.bouncycastle.tsp.TimeStampToken;

/**
 * TspProfile is an abstract implementation of {@code Profile} that uses timestamp provider as the
 * ballot registration service.
 * 
 * @param <T>
 * @param <U>
 */
abstract class TspProfile<T extends TspRespRecord<?>, U extends TspReqRecord<?>>
        implements Profile<T, U, KeyableValue<?>, KeyableValue<?>> {

    private static final String EXT_BALLOT = ".ballot";
    // The canonicalization algorithm: http://www.w3.org/2006/12/xml-c14n11
    private static final String TS_C14N_ALG = Canonicalizer.ALGO_ID_C14N11_OMIT_COMMENTS;
    private static final HashType TS_HASH = HashType.SHA256;
    private static final String SN_PNOEE_PREFIX = "PNOEE-"; // TODO Make country code configurable.

    final ContainerReader container;
    final OcspVerifier ocspVerifier;

    TspProfile(ContainerReader container, OcspVerifier ocspVerifier) {
        this.container = container;
        this.ocspVerifier = ocspVerifier;
    }

    Voter findVoter(VoterProvider vp, String voterId, String version) throws ResultException {
        Voter voter = vp.find(voterId, version);
        if (voter == null) {
            throw new ResultException(Result.VOTER_NOT_FOUND);
        }
        return voter;
    }

    Ballot createBallot(FileName<Ref.BbRef> name, Container c, String version, Voter voter)
            throws ResultException {
        // Find the voter's signature.
        Signature signature = c.getSignatures().stream() //
                .filter(s -> s.getSigner() != null
                        && name.ref.voter.equals(getVoterIdentifier(s.getSigner())))
                .findFirst().orElseThrow(() -> new ResultException(Result.MISSING_VOTER_SIGNATURE));

        Map<String, byte[]> votes = new LinkedHashMap<>();

        c.getFiles().stream().forEach(f -> {
            if (!f.getName().endsWith(EXT_BALLOT)) {
                throw new ResultException(Result.UNKNOWN_FILE_IN_VOTE_CONTAINER, name.path,
                        f.getName());
            }
            // NB! This is not pure <question id>, but <election id>.<question id>
            String voteId = f.getName().substring(0, f.getName().length() - EXT_BALLOT.length());
            votes.put(voteId, Util.toBytes(f.getStream()));
        });

        return new Ballot(name.ref.ballot, signature.getSigningTime(), version, voter, votes);
    }

    private String getVoterIdentifier(Subject s) {
        String serial = s.getSerialNumber();
        if (serial.startsWith(SN_PNOEE_PREFIX)) {
            return serial.substring(SN_PNOEE_PREFIX.length());
        }
        return serial;
    }

    void checkSignatureProfiles(Container c, Signature.Profile profile) {
        c.getSignatures().stream().filter(s -> s.getProfile() != profile).findAny()
                .ifPresent(invalid -> {
                    throw new ResultException(Result.INVALID_SIGNATURE_PROFILE,
                            invalid.getProfile());
                });
    }

    @Override
    public KeyableValue<?> getResponse(T record) {
        return new KeyableValue<>(getResponseKey(record).get());
    }

    @Override
    public KeyableValue<?> getRequest(U record) {
        TimeStampRequest request = record.getRegRequest();
        byte[] bytes = request.getMessageImprintDigest();
        return new KeyableValue<>(createRegKey(bytes));
    }

    @Override
    public Optional<Object> getResponseKey(T record) {
        if (record.hasRegResponse()) {
            TimeStampToken response = record.getRegResponse();
            byte[] bytes = response.getTimeStampInfo().getMessageImprintDigest();
            return Optional.of(createRegKey(bytes));
        }
        byte[] c = getContainer(record);
        if (c == null) {
            return Optional.empty();
        }
        byte[] bytes = TS_HASH.getFunction().digest(container.getTimestampData(c, TS_C14N_ALG));
        return Optional.of(createRegKey(bytes));
    }

    abstract byte[] getContainer(T record);

    private Object createRegKey(byte[] bytes) {
        return new ByteArrayWrapper(bytes);
    }

    @Override
    public Result checkRegistration(KeyableValue<?> response, KeyableValue<?> request) {
        return response.value.equals(request.value) ? Result.OK : Result.REG_RESP_REQ_UNMATCH;
    }

    /**
     * Time-mark profile. Consists of the following files:
     * <ul>
     * <li>bdoc - the BES type vote container</li>
     * <li>ocsptm - the OCSP response, also the timestamp - the source of voting time</li>
     * <li>tspreg - the registration response</li>
     * <li>version - the voterlist version</li>
     * </ul>
     * 
     * <b>NB!</b> Not fully implemented! Add OCSP nonce verification.
     */
    static class TmProfile extends TspProfile<TspRespRecord<TmType>, TspReqRecord<RegType>> {

        TmProfile(ContainerReader container, OcspVerifier ocspVerifier) {
            super(container, ocspVerifier);
        }

        @Override
        byte[] getContainer(TspRespRecord<TmType> record) {
            return record.get(TmType.bdoc);
        }

        @Override
        public TspRespRecord<TmType> createBbRecord() {
            return new TspRespRecord<>(TmType.class, TmType.tspreg);
        }

        @Override
        public TspReqRecord<RegType> createRegRecord() {
            return new TspReqRecord<>(RegType.class, RegType.request);
        }

        @Override
        public byte[] combineBallotContainer(TspRespRecord<TmType> record) {
            byte[] besBdoc = record.get(TmType.bdoc);
            byte[] ocsptm = record.get(TmType.ocsptm);
            // Timestamp is not used in case of TM profile
            return container.combine(besBdoc, ocsptm, null, null);
        }

        @Override
        public Ballot createBallot(FileName<Ref.BbRef> name, TspRespRecord<TmType> record,
                VoterProvider vp, TsVerifier tsv)
                throws Exception, InvalidContainerException, ResultException {
            String version = Util.toString(record.get(TmType.version));
            Voter voter = findVoter(vp, name.ref.voter, version);

            FileName<Ref.BbRef> bdocName = name.forType(TmType.bdoc);
            Container c = container.open(combineBallotContainer(record), bdocName.path);

            checkSignatureProfiles(c, Signature.Profile.BDOC_TM);

            // TODO: Verify OCSP nonce! Are we sure that digidoc4j did not do this already? If so,
            // it should probably happen inside ocspVerifier similarly to tsv.verify.
            // BasicOCSPResp basicResp = ocspVerifier.verify(record.get(TmType.ocsptm));
            // Extension ext = basicResp.getExtension(OCSPObjectIdentifiers.id_pkix_ocsp_nonce);
            // ext.getExtnValue();

            // XXX: TM has no timestamp, so how can we use it to check for registration?
            tsv.verify(record.getRegResponse());

            return createBallot(bdocName, c, version, voter);
        }
    }

    /**
     * Timestamp profile. Consists of the following files:
     * <ul>
     * <li>bdoc - the BES type vote container</li>
     * <li>ocsp - the OCSP response</li>
     * <li>tspreg - the registration response, also the timestamp - the source of voting time</li>
     * <li>version - the voterlist version</li>
     * </ul>
     */
    static class TsProfile extends TspProfile<TspRespRecord<TsType>, TspReqRecord<RegType>> {

        TsProfile(ContainerReader container, OcspVerifier ocspVerifier) {
            super(container, ocspVerifier);
        }

        @Override
        byte[] getContainer(TspRespRecord<TsType> record) {
            return record.get(TsType.bdoc);
        }

        @Override
        public TspRespRecord<TsType> createBbRecord() {
            return new TspRespRecord<>(TsType.class, TsType.tspreg);
        }

        @Override
        public TspReqRecord<RegType> createRegRecord() {
            return new TspReqRecord<>(RegType.class, RegType.request);
        }

        @Override
        public byte[] combineBallotContainer(TspRespRecord<TsType> record) {
            byte[] besBdoc = record.get(TsType.bdoc);
            byte[] ocsp = record.get(TsType.ocsp);
            byte[] ts = record.get(TsType.tspreg);
            return container.combine(besBdoc, ocsp, ts, TS_C14N_ALG);
        }

        @Override
        public Ballot createBallot(FileName<Ref.BbRef> name, TspRespRecord<TsType> record,
                VoterProvider vp, TsVerifier tsv)
                throws Exception, InvalidContainerException, ResultException {
            String version = Util.toString(record.get(TsType.version));
            Voter voter = findVoter(vp, name.ref.voter, version);

            FileName<Ref.BbRef> bdocName = name.forType(TsType.bdoc);
            Container c = container.open(combineBallotContainer(record), bdocName.path);

            checkSignatureProfiles(c, Signature.Profile.BDOC_TS);

            tsv.verify(record.getRegResponse());

            return createBallot(bdocName, c, version, voter);
        }
    }

    enum TmType {
        version, bdoc, ocsptm, tspreg;
    }

    enum TsType {
        version, bdoc, ocsp, tspreg;
    }

    enum RegType {
        request;
    }
}


class TspRespRecord<T extends Enum<T>> extends Record<T> {

    private final T respKey;

    TspRespRecord(Class<T> clazz, T respKey) {
        super(clazz);
        this.respKey = respKey;
    }

    boolean hasRegResponse() {
        return get(respKey) != null;
    }

    TimeStampToken getRegResponse() {
        try {
            return new TimeStampToken(ContentInfo.getInstance(get(respKey)));
        } catch (Exception e) {
            throw new ResultException(Result.REG_RESP_INVALID, e);
        }
    }

}


class TspReqRecord<T extends Enum<T>> extends Record<T> {

    private final T reqKey;

    TspReqRecord(Class<T> clazz, T reqKey) {
        super(clazz);
        this.reqKey = reqKey;
    }

    TimeStampRequest getRegRequest() {
        try {
            return new TimeStampRequest(get(reqKey));
        } catch (Exception e) {
            throw new ResultException(Result.REG_REQ_INVALID, e);
        }
    }

}


class KeyableValue<T> implements Keyable {
    final Object key;
    final T value;

    /**
     * Constructs instance with the given {@code value} as both key and value.
     * 
     * @param value
     */
    KeyableValue(T value) {
        this(value, value);
    }

    private KeyableValue(Object key, T value) {
        this.key = key;
        this.value = value;
    }

    @Override
    public Object getKey() {
        return key;
    }

}
