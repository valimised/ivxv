package ee.ivxv.common.asn1;

public interface ASN1Decodable {
    void readFromBytes(byte[] in) throws ASN1DecodingException;
}
