package ee.ivxv.common.math;

import java.math.BigInteger;

public abstract class GroupElement {
    public abstract BigInteger getOrder();

    public abstract GroupElement op(GroupElement other) throws MathException;

    public abstract GroupElement scale(BigInteger factor);

    public abstract GroupElement inverse();

    public abstract Group getGroup();

    public abstract byte[] getBytes();
}
