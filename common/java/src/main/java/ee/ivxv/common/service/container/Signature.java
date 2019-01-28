package ee.ivxv.common.service.container;

import java.time.Instant;

public class Signature {

    private final Subject signer;
    private final Instant signingTime;
    private final Profile profile;
    private final byte[] value;

    public Signature(Subject signer, Instant signingTime, Profile profile, byte[] value) {
        this.signer = signer;
        this.signingTime = signingTime;
        this.profile = profile;
        this.value = value;
    }

    public Subject getSigner() {
        return signer;
    }

    public Instant getSigningTime() {
        return signingTime;
    }

    public Profile getProfile() {
        return profile;
    }

    public byte[] getValue() {
        return value;
    }

    public enum Profile {
        UNKNOWN, BDOC_TM, BDOC_TS;
    }

}
