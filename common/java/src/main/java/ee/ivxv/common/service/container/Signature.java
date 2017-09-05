package ee.ivxv.common.service.container;

import java.time.Instant;

public class Signature {

    private final Subject signer;
    private final Instant signingTime;
    private final byte[] value;

    public Signature(Subject signer, Instant signingTime, byte[] value) {
        this.signer = signer;
        this.signingTime = signingTime;
        this.value = value;
    }

    public Subject getSigner() {
        return signer;
    }

    public Instant getSigningTime() {
        return signingTime;
    }

    public byte[] getValue() {
        return value;
    }

}
