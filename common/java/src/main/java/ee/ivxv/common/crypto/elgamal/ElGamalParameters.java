package ee.ivxv.common.crypto.elgamal;

import ee.ivxv.common.asn1.ASN1DecodingException;
import ee.ivxv.common.asn1.Field;
import ee.ivxv.common.asn1.Sequence;
import ee.ivxv.common.math.ECGroup;
import ee.ivxv.common.math.ECGroupElement;
import ee.ivxv.common.math.Group;
import ee.ivxv.common.math.GroupElement;
import ee.ivxv.common.math.ModPGroup;
import ee.ivxv.common.math.ModPGroupElement;
import java.math.BigInteger;

public class ElGamalParameters {
    // http://tools.ietf.org/html/draft-rfced-info-pgutmann-00#section-2
    private final static String OID_MODP = "1.3.6.1.4.1.3029.2.1";
    // unassigned by IANA
    private final static String OID_EC = "1.3.6.1.4.1.99999.1";
    private final static String P384 = "P-384";

    private String oid;
    private String ec;
    private Group group;
    private GroupElement generator;
    private String electionIdentifier;

    public ElGamalParameters(Group group, GroupElement generator) throws IllegalArgumentException {
        if (!group.isGroupElement(generator)) {
            throw new IllegalArgumentException("Generator is not group element");
        }
        if (group instanceof ModPGroup && generator instanceof ModPGroupElement) {
            this.group = group;
            this.generator = generator;
            this.oid = OID_MODP;
        } else if (group instanceof ECGroup && generator instanceof ECGroupElement) {
            this.group = group;
            this.generator = generator;
            this.oid = OID_EC;
            this.ec = ((ECGroup) group).getCurveName();
        } else {
            throw new IllegalArgumentException("Group and generator values are unknown");
        }
    }

    public ElGamalParameters(String algOID, byte[] in) throws IllegalArgumentException {
        Sequence seq = new Sequence();
        try {
            seq.readFromBytes(in);
        } catch (ASN1DecodingException e) {
            throw new IllegalArgumentException("Parsing ElGamal parameters failed: " + e);
        }
        initGroupAndGenerator(algOID, seq);
    }

    // duplicate constructors with provided election identifier
    public ElGamalParameters(Group group, GroupElement generator, String electionIdentifier)
            throws IllegalArgumentException {
        this(group, generator);
        this.electionIdentifier = electionIdentifier;
    }

    private void initGroupAndGenerator(String algOID, Sequence seq)
            throws IllegalArgumentException {
        try {
            switch (algOID) {
                case OID_MODP:
                    initModPGroupAndGenerator(seq.getBytes());
                    break;
                case OID_EC:
                    initECGroupAndGenerator(seq.getStrings());
                    break;
                default:
                    throw new IllegalArgumentException("Unknown algorithm OID: " + algOID);
            }
        } catch (ASN1DecodingException e) {
            throw new IllegalArgumentException("Parsing ElGamal parameters failed: " + e);
        }
        oid = algOID;
    }

    private void initModPGroupAndGenerator(byte[][] fields) throws IllegalArgumentException {
        if (fields.length != 3) {
            throw new IllegalArgumentException("Invalid ElGamal parameters");
        }
        ModPGroup G = new ModPGroup(fields[0]);
        generator = new ModPGroupElement(G, fields[1]);
        group = G;
        Field f = new Field();
        String elId;
        try {
            f.readFromBytes(fields[2]);
            elId = f.getString();
        } catch (ASN1DecodingException e) {
            throw new IllegalArgumentException("Can not decode election identifier: " + e);
        }
        if (!elId.equals("")) {
            electionIdentifier = elId;
        }
    }

    private void initECGroupAndGenerator(String[] fields) throws IllegalArgumentException {
        if (fields.length != 2) {
            throw new IllegalArgumentException("Invalid elliptic curve ElGamal parameters");
        }
        switch (fields[0]) {
            case P384:
                ec = P384;
                ECGroup G = new ECGroup(P384);
                generator = G.getBasePoint();
                group = G;
                break;
            default:
                throw new IllegalArgumentException("Unknown elliptic curve ElGamal curve");
        }
        if (!fields[1].equals("")) {
            electionIdentifier = fields[1];
        }
    }

    // return asn1 packed parameters
    public byte[] getBytes() {
        switch (oid) {
            case OID_MODP:
                return getBytesModP();
            case OID_EC:
                return getBytesEC();
            default:
                // does not happen
                return null;
        }
    }

    private byte[] getBytesModP() {
        return new Sequence(group.getBytes(), generator.getBytes(),
                new Field(getElectionIdentifier()).encode()).encode();
    }

    private byte[] getBytesEC() {
        switch (ec) {
            case P384:
                return new Sequence("P-384", getElectionIdentifier()).encode();
            default:
                // does not happen
                return null;
        }
    }

    // return human-readable parameters
    @Override
    public String toString() {
        return String.format("ElGamalParameters(%s, %s)", getGroup(), getGenerator());
    }

    public GroupElement getGenerator() {
        return this.generator;
    }

    public BigInteger getOrder() {
        return this.group.getOrder();
    }

    public BigInteger getGeneratorOrder() {
        return this.generator.getOrder();
    }

    public String getElectionIdentifier() {
        return electionIdentifier == null ? "" : electionIdentifier;
    }

    public Group getGroup() {
        return this.group;
    }

    public String getOID() {
        // the oid depends on the group
        return this.oid;
    }

    @Override
    public boolean equals(Object other) {
        if (this == other)
            return true;
        if (other == null)
            return false;
        if (this.getClass() != other.getClass())
            return false;
        ElGamalParameters opar = (ElGamalParameters) other;
        return this.getElectionIdentifier().equals(opar.getElectionIdentifier())
                && this.getGroup().equals(opar.getGroup())
                && this.getGenerator().equals(opar.getGenerator());
    }

    @Override
    public int hashCode() {
        return this.getElectionIdentifier().hashCode() ^ this.getOrder().hashCode()
                ^ this.getGenerator().hashCode() ^ this.getGeneratorOrder().hashCode();
    }
}
