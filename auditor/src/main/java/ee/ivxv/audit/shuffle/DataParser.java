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

public class DataParser {
    public static final String VER_MODP = "com.verificatum.arithm.ModPGroup";
    public static final String VER_EC = "com.verificatum.arithm.ECqPGroup";

    public static ByteTree[] unmarshalGroup(String pgroup) throws ShuffleException {
        String[] split = pgroup.split("::");
        if (split.length != 2) {
            throw new ShuffleException("Invalid pgroup description string");
        }
        String groupdesc = split[1];
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

    public static ECGroupElement parseECGroupGenerator(ByteTree groupRoot) throws ShuffleException {
        if (groupRoot.isLeaf()) {
            throw new ShuffleException("Invalid Elliptic Curve Group description");
        }
        Leaf ecgroupname = (Leaf) groupRoot;
        ECGroup parsed_group = new ECGroup(ecgroupname.getString());
        ECGroupElement generator = parsed_group.getBasePoint();
        return generator;
    }

    public static ECGroup parseECGroup(ByteTree groupRoot) throws ShuffleException {
        return (ECGroup) parseECGroupGenerator(groupRoot).getGroup();
    }

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

    public static ModPGroup parseModPGroup(ByteTree groupRoot) throws ShuffleException {
        return (ModPGroup) parseModPGroupGenerator(groupRoot).getGroup();
    }

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

    public static GroupElement getAsElement(Group group, ByteTree bt) {
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

    public static GroupElement getAsElement(Group group, ByteTree bt, int index) {
        if (bt.isLeaf()) {
            throw new IllegalArgumentException("Expecting node");
        }
        Node node = (Node) bt;
        return getAsElement(group, node.getNodes()[index]);
    }

    public static ModPGroupElement getAsElement(ModPGroup group, ByteTree bt) {
        if (!bt.isLeaf()) {
            throw new IllegalArgumentException("Expecting leaf");
        }
        Leaf leaf = (Leaf) bt;
        BigInteger val = leaf.getBigInteger();
        return new ModPGroupElement(group, val);
    }

    public static ECGroupElement getAsElement(ECGroup group, ByteTree bt) {
        throw new UnsupportedOperationException("ECGroup element get not supported currently");
    }

    public static ProductGroupElement getAsElement(ProductGroup group, ByteTree bt) {
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

    public static GroupElement[] getAsElementArray(Group group, ByteTree bt) {
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

    public static GroupElement[] getAsElementArray(ProductGroup group, ByteTree bt) {
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

    public static GroupElement[] getAsElementArray(Group group, ByteTree bt, int index) {
        if (bt.isLeaf()) {
            throw new IllegalArgumentException("Expecting node");
        }
        Node node = (Node) bt;
        return getAsElementArray(group, node.getNodes()[index]);
    }

    public static BigInteger getAsInteger(ByteTree bt, int index) {
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

    public static BigInteger[] getAsIntegerArray(ByteTree bt, int index) {
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

    public static MessageDigest getHash(String hashname) {
        MessageDigest md;
        try {
            md = MessageDigest.getInstance(hashname);

        } catch (NoSuchAlgorithmException e) {
            throw new IllegalArgumentException(
                    String.format("Hash function %s not supported", hashname));
        }
        return md;
    }

    public static ProductGroupElement toArray(GroupElement[] elements) {
        ProductGroupElement[] casted = new ProductGroupElement[elements.length];
        for (int i = 0; i < casted.length; i++) {
            if (!(elements[i] instanceof ProductGroupElement)) {
                throw new IllegalArgumentException("Invalid group");
            }
            casted[i] = (ProductGroupElement) elements[i];
        }
        return toArray(casted);
    }

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
