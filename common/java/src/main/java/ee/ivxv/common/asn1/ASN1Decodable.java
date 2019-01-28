package ee.ivxv.common.asn1;

/**
 * Interface for classes which can decoded as ASN1 structures.
 *
 */
public interface ASN1Decodable {
    /**
     * Set the instance fields from DER-encoded byte array.
     * 
     * @param in ASN1-encoded serialization of instance.
     * @throws ASN1DecodingException When decoding fails.
     */
    void readFromBytes(byte[] in) throws ASN1DecodingException;
}
