package ee.ivxv.common.math;

import ee.ivxv.common.crypto.Plaintext;
import java.math.BigInteger;

/**
 * Interface for defining a group.
 */
public abstract class Group {
    /**
     * Get a group element from serialized representation.
     * 
     * @param data
     * @return
     * @throws IllegalArgumentException
     */
    public abstract GroupElement getElement(byte[] data) throws IllegalArgumentException;

    /**
     * Get the order of the group.
     * 
     * @return
     */
    public abstract BigInteger getOrder();

    /**
     * Get the order of the underlying field of the group.
     * 
     * @return
     */
    public abstract BigInteger getFieldOrder();

    /**
     * Get the identity element of the group.
     * 
     * @return
     */
    public abstract GroupElement getIdentity();

    /**
     * Encode a message as a group element.
     * 
     * @param msg
     * @return
     * @throws MathException
     */
    public abstract GroupElement encode(Plaintext msg) throws MathException;

    /**
     * Pad the message to the full length.
     * 
     * @param msg
     * @return
     */
    public abstract Plaintext pad(Plaintext msg);

    /**
     * Check that an arbitrary group element can be decoded as this group element.
     * 
     * @see Decodable
     * 
     * @param el
     * @return
     */
    public abstract Decodable isDecodable(GroupElement el);

    /**
     * Decode a group element as a plaintext message.
     * 
     * @param msg
     * @return
     */
    public abstract Plaintext decode(GroupElement msg);

    /**
     * Check that the element is group element.
     * 
     * @param el
     * @return
     */
    public abstract boolean isGroupElement(GroupElement el);

    /**
     * Serialize the group representation.
     * 
     * @return
     */
    public abstract byte[] getBytes();

    /**
     * Values for decodability check result.
     */
    public enum Decodable {
        VALID, INVALID_GROUP, INVALID_RANGE, INVALID_QR, INVALID_POINT;
    }
}
