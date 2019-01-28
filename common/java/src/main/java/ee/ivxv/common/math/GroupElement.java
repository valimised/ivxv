package ee.ivxv.common.math;

import java.math.BigInteger;

/**
 * Interface for defining a group element and required supported operations.
 */
public abstract class GroupElement {
    /**
     * Get the order of the element.
     * 
     * @return
     */
    public abstract BigInteger getOrder();

    /**
     * Perform a group operation with other group element.
     * 
     * @param other
     * @return
     * @throws MathException
     */
    public abstract GroupElement op(GroupElement other) throws MathException;

    /**
     * Scale the element. Equivalent to invoking {@link #op(GroupElement)} factor times.
     * 
     * @param factor
     * @return
     */
    public abstract GroupElement scale(BigInteger factor);

    /**
     * Return the inverse element.
     * 
     * @return
     */
    public abstract GroupElement inverse();

    /**
     * Get the group of the element.
     * 
     * @return
     */
    public abstract Group getGroup();

    /**
     * Serialize the element.
     * 
     * @return
     */
    public abstract byte[] getBytes();
}
