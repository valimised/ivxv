package ee.ivxv.common.service.bbox.impl.verify;

import java.io.IOException;
import org.bouncycastle.asn1.ASN1Encodable;
import org.bouncycastle.asn1.ASN1OctetString;
import org.bouncycastle.asn1.ASN1Sequence;
import org.bouncycastle.asn1.DEROctetString;
import org.bouncycastle.asn1.DERSequence;
import org.bouncycastle.asn1.x509.AlgorithmIdentifier;

public class TsSignature {

    private final AlgorithmIdentifier alg;
    private final byte[] signature;

    public TsSignature(AlgorithmIdentifier alg, byte[] signature) {
        this.alg = alg;
        this.signature = signature;
    }

    public static TsSignature fromBytes(byte[] in) {
        ASN1Sequence seq = ASN1Sequence.getInstance(in);
        AlgorithmIdentifier alg = AlgorithmIdentifier.getInstance(seq.getObjectAt(0));
        byte[] signature = ASN1OctetString.getInstance(seq.getObjectAt(1)).getOctets();

        return new TsSignature(alg, signature);
    }

    public AlgorithmIdentifier getAlg() {
        return alg;
    }

    public byte[] getSignature() {
        return signature;
    }

    /**
     * <pre>
     *    Signature ::= SEQUENCE  {
     *       signingAlgorithm   AlgorithmIdentifier,
     *       signature          ANY DEFINED BY signingAlgorithm
     * }
     * </pre>
     */
    public byte[] toBytes() throws IOException {
        return new DERSequence(new ASN1Encodable[] {alg, new DEROctetString(signature)})
                .getEncoded();
    }

}
