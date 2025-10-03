package ee.ivxv.common.service.bbox;

/**
 * Represents the result of various operations related to ballot box processing.
 *
 * The ballot box processing includes importing and analyzing both ballots and
 * registration requests.
 *
 * Required files for a ballot are: version, ocsp, bdoc, tspreg
 * Required files for a registration request are: request
 *
 */
public enum Result {
    /**
     * Operation completed successfully.
     */
    OK,

    /**
     * The file name provided does not correspond to the expected pattern.
     */
    INVALID_FILE_NAME,

    /**
     * A required file for an entity is missing.
     */
    MISSING_FILE,

    /**
     * A required file for an entity has been found multiple times.
     */
    REPEATED_FILE,

    /**
     * The file type is unknown or unsupported.
     */
    UNKNOWN_FILE_TYPE,

    /**
     * The file size for a signed ballot does not meet the required criteria.
     */
    INVALID_FILE_SIZE,

    /**
     * The ballot signature is invalid.
     */
    INVALID_BALLOT_SIGNATURE,

    /**
     * The ballot does not contain the voter's signature
     */
    MISSING_VOTER_SIGNATURE,

    /**
     * The voter was not eligible according to the voterlists at the time of voting
     */
    VOTER_NOT_FOUND,

    /**
     * The voter list corresponding to the version cannot be found.
     */
    VOTERLIST_NOT_FOUND,

    /**
     * The vote has been cast before the beginning of the voting period.
     */
    TIME_BEFORE_START,

    /**
     * The registration response is invalid.
     */
    REG_RESP_INVALID,

    /**
     * The registration request is invalid.
     */
    REG_REQ_INVALID,

    /**
     * The registration response is not unique.
     *
     * It may happen, when registration request is sent for the same ballot
     * more than once. Only the earliest response is used, the repeated
     * responses are reported.
     */
    REG_RESP_NOT_UNIQUE,

    /**
     * The registration request is not unique.
     *
     * It may happen, when registration request is sent for the same ballot
     * more than once. Only the earliest response is used, the repeated
     * responses are reported.
     *
     */
    REG_REQ_NOT_UNIQUE,

    /**
     * No nonce is provided in the registration request.
     */
    REG_NO_NONCE,

    /**
     * The provided nonce is not signed according to IVXV protocol.
     */
    REG_NONCE_NOT_SIG,

    /**
     * The algorithm used for the nonce signature does not match the expected one.
     */
    REG_NONCE_ALG_MISMATCH,

    /**
     * The nonce signature is invalid.
     */
    REG_NONCE_SIG_INVALID,

    /**
     * An unknown file is present in the vote container.
     */
    UNKNOWN_FILE_IN_VOTE_CONTAINER,

    /**
     * A technical error occurred during processing.
     */
    TECHNICAL_ERROR,

    /**
     * The data in registration response does not match the data in the
     * registration request.
     *
     * This situation must be investigated, it must not occur under normal
     * circumstances.
     */
    REG_RESP_REQ_UNMATCH,

    /**
     * A registration request has been provided, but the ballot is missing.
     *
     * This situation must be investigated, since it seems that collection
     * service has handed over incomplete ballot box. It must not occur under
     * normal circumstances.
     */
    REG_REQ_WITHOUT_BALLOT,

    /**
     * A ballot is submitted without a corresponding registration request.
     */
    BALLOT_WITHOUT_REG_REQ,

    /**
     *  There are two votes from the same voter that could be considered the latest.
     *
     *  Their timestamps identify the same time. The valid vote is selected randomly.
     */
    SAME_TIME_AS_LATEST,

    /**
     * The signature profile of the vote is invalid.
     */
    INVALID_SIGNATURE_PROFILE,
}
