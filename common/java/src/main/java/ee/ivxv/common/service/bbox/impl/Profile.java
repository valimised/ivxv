package ee.ivxv.common.service.bbox.impl;

import ee.ivxv.common.model.Ballot;
import ee.ivxv.common.service.bbox.BboxHelper.VoterProvider;
import ee.ivxv.common.service.bbox.Ref;
import ee.ivxv.common.service.bbox.Result;
import ee.ivxv.common.service.bbox.impl.verify.TsVerifier;
import ee.ivxv.common.service.container.InvalidContainerException;
import java.util.Optional;

/**
 * Profile abstracts the contents of ballot box and registration data.
 * 
 * <p>
 * The generic concept is:
 * <ul>
 * <li>There are 2 main data types - ballot box and registration data.
 * <li>{@code Profile} class provides method to create records for each.
 * <li>Ballot box records are the primary source for creating instances of {@code Ballot}.
 * <li>Registration data records are needed only to check correlation between the two collections.
 * <li>For correlation a <i>registration response</i> must match a <i>registration request</i>.
 * <li>Registration requests are stored in the registration data and can be created from a
 * registration record.
 * <li>Registration responses are stored in the ballot box and can be created from a ballot box
 * record.
 * <li>Both request and response must have a <u>unique</u> hash key for quick lookup.
 * </ul>
 * 
 * @param <T> Ballot box record type
 * @param <U> Registration data record type
 * @param <RT> Registration response type - bound to ballot box
 * @param <RU> Registration request type - bound to registration data
 */
public interface Profile<T extends Record<?>, U extends Record<?>, RT extends Keyable, RU extends Keyable> {

    T createBbRecord();

    U createRegRecord();

    byte[] combineBallotContainer(T record);

    Ballot createBallot(FileName<Ref.BbRef> name, T record, VoterProvider vp, TsVerifier tsv)
            throws Exception, InvalidContainerException, ResultException;

    RT getResponse(T record);

    RU getRequest(U record);

    /**
     * Tries to create response key from potentially partial or invalid record in order to match it
     * later against the request key.
     * 
     * @param record
     * @return Optional response key, possibly empty optional.
     */
    Optional<Object> getResponseKey(T record);

    Result checkRegistration(RT response, RU request);

}


interface Keyable {
    Object getKey();
}
