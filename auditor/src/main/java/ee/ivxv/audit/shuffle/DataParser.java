package ee.ivxv.audit.shuffle;

import ee.ivxv.audit.shuffle.ByteTree.Leaf;
import ee.ivxv.audit.shuffle.ByteTree.Node;
import ee.ivxv.common.math.ECGroup;
import ee.ivxv.common.math.ECGroupElement;
import ee.ivxv.common.math.Group;
import ee.ivxv.common.math.GroupElement;
import ee.ivxv.common.math.ModPGroup;
import ee.ivxv.common.math.ModPGroupElement;
import ee.ivxv.common.math.ProductGroup;
import ee.ivxv.common.math.ProductGroupElement;
import java.math.BigInteger;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import javax.xml.bind.DatatypeConverter;

/**
 * DataParser is a utility class for performing operations with ByteTree objects.
 */
public class DataParser {
    /**
     * Identifier for groups of integers modulo a value.
     */
    public static final String VER_MODP = "com.verificatum.arithm.ModPGroup";
    /**
     * Identifier for elliptic curve groups.
     */
    public static final String VER_EC = "com.verificatum.arithm.ECqPGroup";

    /**
     * Extract Verificatum group information node from Verificatum group description.
     * <p>
     * Returns list [{@literal description}, {@literal groupinfo}], where {@literal decription} is a
     * string description of the group and {@literal groupinfo} is group-specific ByteTree
     * description of the group.
     * 
     * @param pgroup String group description
     * @return Array of ByteTree elements representing group
     * @throws ShuffleException When unmarshalling fails.
     */
    public static ByteTree[] unmarshalGroup(String pgroup) throws ShuffleException {
        String[] split = pgroup.split("::");
        if (split.length != 2) {
            throw new ShuffleException("Invalid pgroup description string");
        }
        String groupdesc = split[1];
        // in some cases, the pgroup description has newline in the end. omit it
        if (groupdesc.length() % 2 == 1 && groupdesc.charAt(groupdesc.length() - 1) == '\n') {
            groupdesc = groupdesc.substring(0, groupdesc.length() - 1);
        }
        byte[] groupdescb = DatatypeConverter.parseHexBinary(groupdesc);
        ByteTree G = ByteTree.parse(groupdescb);
        if (G.isLeaf()) {
            throw new ShuffleException("Invalid pgroup description bytetree");
        }
        ByteTree.Node pginfo = (ByteTree.Node) G;
        if (pginfo.getLength() != 2) {
            throw new ShuffleException("Invalid marshalled pgroup description");
        }
        return pginfo.getNodes();
    }

    /**
     * Extract elliptic curve generator from Verificatum elliptic curve group description leaf.
     * <p>
     * Returns the base point corresponding to the elliptic curve. Description leaf should be a
     * string leaf containing curve name.
     * 
     * @param groupRoot ByteTree representation of elliptic curve group
     * @return Elliptic curve base point
     * @throws ShuffleException When parsing fails.
     * @throws IllegalArgumentException When unknown elliptic curve group is defined.
     */
    public static ECGroupElement parseECGroupGenerator(ByteTree groupRoot)
            throws ShuffleException, IllegalArgumentException {
        if (groupRoot.isLeaf()) {
            throw new ShuffleException("Invalid Elliptic Curve Group description");
        }
        Leaf ecgroupname = (Leaf) groupRoot;
        ECGroup parsed_group = new ECGroup(ecgroupname.getString());
        ECGroupElement generator = parsed_group.getBasePoint();
        return generator;
    }

    /**
     * Extract elliptic curve group from Verificatum elliptic curve group description node.
     * <p>
     * 
     * @see #parseECGroupGenerator(ByteTree) for input node format.
     * 
     * @param groupRoot ByteTree representation of elliptic curve group.
     * @return {@link ee.ivxv.common.math.ECGroup} instance representing the group.
     * @throws ShuffleException When failing to parse the node
     */
    public static ECGroup parseECGroup(ByteTree groupRoot)
            throws ShuffleException, IllegalArgumentException {
        return (ECGroup) parseECGroupGenerator(groupRoot).getGroup();
    }

    /**
     * Extract generator for group of integers modulo a safe prime from Verificatum group
     * description node.
     * <p>
     * Expect as input a node [{@literal p}, {@literal q}, {@literal g}], where p is the modulus of
     * the group, q the order of the group and g the generator of the group.
     * 
     * @param groupRoot ByteTree representation of group of integers modulo a prime.
     * @return Group generator
     * @throws ShuffleException When failing to parse the node.
     */
    public static ModPGroupElement parseModPGroupGenerator(ByteTree groupRoot)
            throws ShuffleException {
        if (groupRoot.isLeaf()) {
            throw new ShuffleException("Invalid ModPGroup description");
        }
        ByteTree.Node marshalled = (ByteTree.Node) groupRoot;
        if (marshalled.getLength() != 4) {
            throw new ShuffleException("Invalid ModPGroup description length");
        }
        ByteTree pnode = marshalled.getNodes()[0];
        if (!pnode.isLeaf()) {
            throw new ShuffleException("Invalid modulus leaf");
        }
        BigInteger p = ((Leaf) pnode).getBigInteger();
        ByteTree gnode = marshalled.getNodes()[2];
        if (!gnode.isLeaf()) {
            throw new ShuffleException("Invalid generator leaf");
        }
        BigInteger g = ((Leaf) gnode).getBigInteger();
        ModPGroup parsed_group = new ModPGroup(p);
        ModPGroupElement generator = new ModPGroupElement(parsed_group, g);
        return generator;
    }

    /**
     * Extract group of integers modulo a prime from Verificatum group description node.
     * <p>
     * 
     * @see #parseModPGroupGenerator(ByteTree).
     * 
     * @param groupRoot ByteTree representation of group of integers modulo a prime.
     * @return {@link ee.ivxv.common.math.ModPGroup} representing a group
     * @throws ShuffleException When failing to parse the node
     */
    public static ModPGroup parseModPGroup(ByteTree groupRoot) throws ShuffleException {
        return (ModPGroup) parseModPGroupGenerator(groupRoot).getGroup();
    }

    /**
     * Parse group generator from Verificatum group description node.
     * <p>
     * 
     * @see #parseModPGroupGenerator(ByteTree)
     * @see #parseECGroupGenerator(ByteTree)
     * 
     * @param pgroup ByteTree representation of group.
     * @return {@link ee.ivxv.common.math.GroupElement} representing the group generator.
     * @throws ShuffleException When failing to parse the node
     */
    public static GroupElement parseGroupGenerator(String pgroup) throws ShuffleException {
        ByteTree[] nodes = unmarshalGroup(pgroup);
        if (!nodes[0].isLeaf()) {
            throw new ShuffleException("Invalid marshalled pgroup description string");
        }
        Leaf descstring = (Leaf) nodes[0];
        GroupElement generator = null;
        if (descstring.getString().equals(VER_MODP)) {
            generator = parseModPGroupGenerator(nodes[1]);
        } else if (descstring.getString().equals(VER_EC)) {
            generator = parseECGroupGenerator(nodes[1]);
        } else {
            throw new ShuffleException("Invalid group");
        }
        return generator;
    }

    /**
     * Get the ByteTree node as a group element.
     * <p>
     * The Verificatum element representation node bt must be a correct representation of the group
     * element. See descriptions of specific group methods for more detailed structures.
     * <p>
     * 
     * @see #getAsElement(ECGroup, ByteTree)
     * @see #getAsElement(ModPGroup, ByteTree)
     * @see #getAsElement(ProductGroup, ByteTree)
     * 
     * @param group Group where the element belongs
     * @param bt Element representation in Verificatum ByteTree format
     * @return Group element instance
     * @throws IllegalArgumentException when parsing fails
     */
    public static GroupElement getAsElement(Group group, ByteTree bt)
            throws IllegalArgumentException {
        if (group instanceof ModPGroup) {
            return getAsElement((ModPGroup) group, bt);
        } else if (group instanceof ECGroup) {
            return getAsElement((ECGroup) group, bt);
        } else if (group instanceof ProductGroup) {
            return getAsElement((ProductGroup) group, bt);
        } else {
            throw new IllegalArgumentException("Invalid group");
        }
    }

    /**
     * Get the indexed node from the ByteTree nodes as a group element
     * <p>
     * Indexed version of {@link #getAsElement(Group, ByteTree)}.
     * <p>
     * 
     * @see #getAsElement(Group, ByteTree).
     * 
     * @param group Group where the element belongs
     * @param bt Elements node representation in Verificatum ByteTree format
     * @param index Element index
     * @return Group element instance
     * @throws IllegalArgumentException when parsing fails
     */
    public static GroupElement getAsElement(Group group, ByteTree bt, int index)
            throws IllegalArgumentException {
        if (bt.isLeaf()) {
            throw new IllegalArgumentException("Expecting node");
        }
        Node node = (Node) bt;
        return getAsElement(group, node.getNodes()[index]);
    }

    /**
     * Get the ByteTree leaf as ModPGroupElement
     * <p>
     * We assume that the leaf is an integer. It is represented as an element in group of modulo a
     * prime.
     * 
     * @param group Group where the element belongs.
     * @param bt Integer leaf
     * @return {@link ee.ivxv.common.math.ModPGroupElement} instance
     * @throws IllegalArgumentException When parsing fails
     */
    public static ModPGroupElement getAsElement(ModPGroup group, ByteTree bt)
            throws IllegalArgumentException {
        if (!bt.isLeaf()) {
            throw new IllegalArgumentException("Expecting leaf");
        }
        Leaf leaf = (Leaf) bt;
        BigInteger val = leaf.getBigInteger();
        return new ModPGroupElement(group, val);
    }

    /**
     * Get the ByteTree node as an ECGroupElement
     * <p>
     * Currently not supported.
     * 
     * @param group
     * @param bt
     * @return
     * @throws UnsupportedOperationException Always.
     */
    public static ECGroupElement getAsElement(ECGroup group, ByteTree bt)
            throws UnsupportedOperationException {
        throw new UnsupportedOperationException("ECGroup element get not supported currently");
    }

    /**
     * Get the ByteTree node as an ProductGroupElement.
     * <p>
     * We assume that the node is an array of group element nodes. Every node is parsed as a
     * suitable group element from the product group.
     * 
     * @param group Product group instance where the element belongs
     * @param bt Node representing the element
     * @return {@link ee.ivxv.common.math.ProductGroupElement} instance.
     * @throws IllegalArgumentException When parsing fails.
     */
    public static ProductGroupElement getAsElement(ProductGroup group, ByteTree bt)
            throws IllegalArgumentException {
        if (bt.isLeaf()) {
            throw new IllegalArgumentException("Expecting node");
        }
        Group[] groups = ((ProductGroup) group).getGroups();
        GroupElement[] el = new GroupElement[groups.length];
        for (int i = 0; i < groups.length; i++) {
            el[i] = getAsElement(groups[i], bt, i);
        }
        return new ProductGroupElement(group, el);
    }

    /**
     * Get the ByteTree node as an array of group elements.
     * <p>
     * We assume that the node is an array of group element representations. See specific group
     * element getter methods for exact structures.
     * <p>
     * 
     * @see #getAsElement(Group, ByteTree).
     * 
     * @param group Group which the elements are part of.
     * @param bt Node which is an array of elements.
     * @return An array of {@link ee.ivxv.common.math.GroupElement} instances.
     * @throws IllegalArgumentException When parsing fails.
     */
    public static GroupElement[] getAsElementArray(Group group, ByteTree bt)
            throws IllegalArgumentException {
        if (bt.isLeaf()) {
            throw new IllegalArgumentException("Expecting root node");
        }
        if (group instanceof ProductGroup) {
            // handle special case
            return getAsElementArray((ProductGroup) group, bt);
        }
        Node rootnode = (Node) bt;
        ByteTree[] nodes = rootnode.getNodes();
        GroupElement[] ret = new GroupElement[nodes.length];
        for (int i = 0; i < nodes.length; i++) {
            ret[i] = getAsElement(group, nodes[i]);
        }
        return ret;
    }

    /**
     * Get the ByteTree node as an array of product group elements.
     * <p>
     * We assume that the node is an array of group element representations. See specific group
     * element getter methods for exact structures.
     * <p>
     * 
     * @see #getAsElement(ProductGroup, ByteTree)
     * 
     * @param group Group which the elements are part of.
     * @param bt Node which is an array of elements.
     * @return An array of {@link ee.ivxv.common.math.GroupElement} instances.
     * @throws IllegalArgumentException When parsing fails.
     */
    public static GroupElement[] getAsElementArray(ProductGroup group, ByteTree bt)
            throws IllegalArgumentException {
        Node rootnode = (Node) bt;
        Group[] groups = group.getGroups();
        GroupElement[][] sub = new GroupElement[groups.length][];
        for (int i = 0; i < groups.length; i++) {
            sub[i] = getAsElementArray(groups[i], rootnode.getNodes()[i]);
        }
        GroupElement[] ret = new GroupElement[sub[0].length];
        for (int i = 0; i < ret.length; i++) {
            GroupElement[] cons = new GroupElement[groups.length];
            for (int j = 0; j < groups.length; j++) {
                cons[j] = sub[j][i];
            }
            ret[i] = new ProductGroupElement(group, cons);
        }
        return ret;
    }

    /**
     * Get the indexed ByteTree node as an array of group elements.
     * <p>
     * Indexed version of {@link #getAsElementArray(Group, ByteTree)}.
     * <p>
     * 
     * @see #getAsElementArray(Group, ByteTree)
     * 
     * @param group Group which the elements are part of.
     * @param bt Node which is an array of elements.
     * @param index Index of the node to use.
     * @return An array of {@link ee.ivxv.common.math.GroupElement} instances.
     * @throws IllegalArgumentException When parsing fails.
     */
    public static GroupElement[] getAsElementArray(Group group, ByteTree bt, int index)
            throws IllegalArgumentException {
        if (bt.isLeaf()) {
            throw new IllegalArgumentException("Expecting node");
        }
        Node node = (Node) bt;
        return getAsElementArray(group, node.getNodes()[index]);
    }

    /**
     * Get the indexed ByteTree node as an integer.
     * <p>
     * The index element of the node must be an integer leaf.
     * 
     * @param bt Node
     * @param index Index of the element
     * @return Integer value
     * @throws IllegalArgumentException When parsing fails.
     */
    public static BigInteger getAsInteger(ByteTree bt, int index) throws IllegalArgumentException {
        if (bt.isLeaf()) {
            throw new IllegalArgumentException("Root should be node");
        }
        ByteTree[] nodes = ((Node) bt).getNodes();
        if (!nodes[index].isLeaf()) {
            throw new IllegalArgumentException("Should be leaf");
        }
        Leaf leaf = (Leaf) nodes[index];
        BigInteger res = leaf.getBigInteger();
        return res;
    }

    /**
     * Get the indexed ByteTree node as an array of integers.
     * <p>
     * The index element of the node must be a node consisting of integer leafs.
     * 
     * @param bt Node
     * @param index Index of the element.
     * @return Array of integer values.
     * @throws IllegalArgumentException When parsing fails.
     */
    public static BigInteger[] getAsIntegerArray(ByteTree bt, int index)
            throws IllegalArgumentException {
        if (bt.isLeaf()) {
            throw new IllegalArgumentException("Root should be node");
        }
        ByteTree[] nodes = ((Node) bt).getNodes();
        if (nodes[index].isLeaf()) {
            throw new IllegalArgumentException("Should be node");
        }
        ByteTree[] intnodes = ((Node) nodes[index]).getNodes();
        BigInteger[] res = new BigInteger[intnodes.length];
        for (int i = 0; i < intnodes.length; i++) {
            if (!intnodes[i].isLeaf()) {
                throw new IllegalArgumentException("Should be leaf");
            }
            res[i] = ((Leaf) intnodes[i]).getBigInteger();
        }
        return res;
    }

    /**
     * Get the hash corresponding to string representation.
     * <p>
     * The hash name is delegated to {@link java.security.MessageDigest}.
     * 
     * @param hashname Hash name.
     * @return Hash instance.
     * @throws IllegalArgumentException When can not find hash with hashname.
     */
    public static MessageDigest getHash(String hashname) throws IllegalArgumentException {
        MessageDigest md;
        try {
            md = MessageDigest.getInstance(hashname);

        } catch (NoSuchAlgorithmException e) {
            throw new IllegalArgumentException(
                    String.format("Hash function %s not supported", hashname));
        }
        return md;
    }

    /**
     * Convert GroupElement array to a ProductGroupElement instance.
     * <p>
     * Currently only ProductGroupElement instances are supported.
     * 
     * @param elements
     * @return
     * @throws IllegalArgumentException When converting fails
     */
    public static ProductGroupElement toArray(GroupElement[] elements)
            throws IllegalArgumentException {
        ProductGroupElement[] casted = new ProductGroupElement[elements.length];
        for (int i = 0; i < casted.length; i++) {
            if (!(elements[i] instanceof ProductGroupElement)) {
                throw new IllegalArgumentException("Invalid group");
            }
            casted[i] = (ProductGroupElement) elements[i];
        }
        return toArray(casted);
    }

    /**
     * Convert ProductGroupElement array to a ProductGroupElement instance.
     * <p>
     * 
     * @see #toArray(GroupElement[]).
     * 
     * @param elements
     * @return
     */
    public static ProductGroupElement toArray(ProductGroupElement[] elements) {
        ProductGroupElement r = toArray_first(elements[0], elements.length);
        for (int i = 0; i < elements.length; i++) {
            toArray_second(r, elements[i], i);
        }
        return r;
    }

    private static ProductGroupElement toArray_first(ProductGroupElement in, int N) {
        GroupElement[] gs = in.getElements();
        GroupElement[] ret = new GroupElement[gs.length];
        Group[] retgs = new Group[gs.length];
        for (int i = 0; i < gs.length; i++) {
            if (gs[i] instanceof ProductGroupElement) {
                ret[i] = toArray_first((ProductGroupElement) gs[i], N);
            } else {
                ret[i] = new ProductGroupElement(new ProductGroup(gs[i].getGroup(), N));
            }
            retgs[i] = ret[i].getGroup();
        }
        return new ProductGroupElement(new ProductGroup(retgs), ret);
    }

    private static void toArray_second(ProductGroupElement out, ProductGroupElement in, int N) {
        GroupElement[] inel = in.getElements();
        GroupElement[] outel = out.getElements();
        for (int i = 0; i < inel.length; i++) {
            if (inel[i] instanceof ProductGroupElement) {
                toArray_second((ProductGroupElement) outel[i], (ProductGroupElement) inel[i], N);
            } else {
                ((ProductGroupElement) outel[i]).getElements()[N] = inel[i];
            }
        }
    }
}
