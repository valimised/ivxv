package ee.ivxv.common.service.bbox.impl;

import ee.ivxv.common.model.Ballot;
import ee.ivxv.common.model.Voter;
import ee.ivxv.common.service.bbox.BboxHelper.VoterProvider;
import ee.ivxv.common.service.bbox.Ref;
import ee.ivxv.common.service.bbox.Result;
import ee.ivxv.common.service.bbox.impl.verify.OcspVerifier;
import ee.ivxv.common.service.bbox.impl.verify.TsVerifier;
import ee.ivxv.common.service.container.ContainerReader;
import ee.ivxv.common.service.container.Container;
import ee.ivxv.common.service.container.InvalidContainerException;
import ee.ivxv.common.util.Util;
import java.math.BigInteger;
import java.time.Instant;
import java.util.LinkedHashMap;
import java.util.Map;
import org.apache.xml.security.c14n.Canonicalizer;
import org.bouncycastle.asn1.cms.ContentInfo;
import org.bouncycastle.cert.ocsp.BasicOCSPResp;
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
        implements Profile<T, U, KeyableValue<BigInteger>, KeyableValue<BigInteger>> {

    private static final String EXT_BALLOT = ".ballot";
    // The canonicalization algorithm: http://www.w3.org/2006/12/xml-c14n11
    private static final String TS_C14N_ALG = Canonicalizer.ALGO_ID_C14N11_OMIT_COMMENTS;

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

    Ballot createBallot(FileName<Ref.BbRef> name, Container c, Instant time, String version,
            Voter voter) throws ResultException {
        // Check that ballot has the signature of the voter
        if (!c.getSignatures().stream().anyMatch(s -> s.getSigner() != null
                && name.ref.voter.equals(s.getSigner().getSerialNumber()))) {
            throw new ResultException(Result.MISSING_VOTER_SIGNATURE);
        }

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

        return new Ballot(name.ref.ballot, time, version, voter, votes);
    }

    @Override
    public KeyableValue<BigInteger> getResponse(T record) throws Exception {
        TimeStampToken response = record.getRegResponse();
        return new KeyableValue<>(response.getTimeStampInfo().getNonce());
    }

    @Override
    public KeyableValue<BigInteger> getRequest(U record) throws Exception {
        TimeStampRequest request = record.getRegRequest();
        return new KeyableValue<>(request.getNonce());
    }

    @Override
    public Result checkRegistration(KeyableValue<BigInteger> response,
            KeyableValue<BigInteger> request) {
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

            BasicOCSPResp basicResp = ocspVerifier.verify(record.get(TmType.ocsptm));

            tsv.verify(record.getRegResponse());

            // TO-DO Verify OCSP nonce!
            // Extension ext = basicResp.getExtension(OCSPObjectIdentifiers.id_pkix_ocsp_nonce);
            // ext.getExtnValue();

            Instant time = basicResp.getProducedAt().toInstant();

            return createBallot(bdocName, c, time, version, voter);
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

            ocspVerifier.verify(record.get(TsType.ocsp));

            TimeStampToken tst = tsv.verify(record.getRegResponse());

            Instant time = tst.getTimeStampInfo().getGenTime().toInstant();

            return createBallot(bdocName, c, time, version, voter);
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

    TimeStampToken getRegResponse() throws Exception {
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

    TimeStampRequest getRegRequest() throws Exception {
        try {
            return new TimeStampRequest(get(reqKey));
        } catch (Exception e) {
            throw new ResultException(Result.REG_REQ_INVALID, e);
        }
    }

}


class KeyableValue<T> implements Keyable {
    final T value;
    final Object key;

    /**
     * Constructs instance with the given {@code key} as both key and value.
     * 
     * @param key
     */
    KeyableValue(T key) {
        this(key, key);
    }

    KeyableValue(T value, Object key) {
        this.value = value;
        this.key = key;
    }

    @Override
    public Object getKey() {
        return key;
    }

}
