package ee.ivxv.common.math;

import ee.ivxv.common.crypto.Plaintext;
import java.math.BigInteger;

public abstract class Group {
    public abstract GroupElement getElement(byte[] data) throws IllegalArgumentException;

    public abstract BigInteger getOrder();

    public abstract BigInteger getFieldOrder();

    public abstract GroupElement getIdentity();

    public abstract GroupElement encode(Plaintext msg) throws MathException;

    public abstract Plaintext pad(Plaintext msg);

    public abstract Plaintext decode(GroupElement msg);

    public abstract boolean isGroupElement(GroupElement el);

    public abstract byte[] getBytes();
}
