package ee.ivxv.common.asn1;

/**
 * Interface to classes which can be encoded in ASN1 DER format.
 *
 */
public interface ASN1Encodable {
    /**
     * Serialize the instance using ASN1.
     * 
     * @return ASN1-encoded serialization of the instance.
     */
    byte[] encode();
}
