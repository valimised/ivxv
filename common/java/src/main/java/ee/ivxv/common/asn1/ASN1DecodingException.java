package ee.ivxv.common.asn1;

/**
 * ASN1DecodingException denotes ASN1 decoding exception.
 *
 */
@SuppressWarnings("serial")
public class ASN1DecodingException extends Exception {
    public ASN1DecodingException(String str) {
        super(str);
    }
}
