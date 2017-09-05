package ee.ivxv.audit.shuffle;

import ee.ivxv.common.math.Group;
import ee.ivxv.common.math.GroupElement;
import java.io.IOException;
import java.nio.file.Path;
import javax.xml.parsers.DocumentBuilder;
import javax.xml.parsers.DocumentBuilderFactory;
import javax.xml.parsers.ParserConfigurationException;
import org.w3c.dom.Document;
import org.w3c.dom.Element;
import org.w3c.dom.Node;
import org.w3c.dom.NodeList;
import org.xml.sax.SAXException;

public class ProtocolInformation {
    private String version, sid, name, pgroup, prg, rohash, auxsid, type;
    private int statdist, width, ebitlenro, vbitlenro, keywidth;
    private Group pgroup_parsed;
    private GroupElement generator_parsed;

    private static final String default_auxsid = "default";
    private static final String default_type = "shuffling";
    private static final int default_keywidth = 5;

    public ProtocolInformation(String version, String sid, String name, String pgroup, int keywidth,
            int vbitlenro, int ebitlenro, String prg, String rohash, int width, int statdist) {
        this.version = version;
        this.sid = sid;
        this.name = name;
        this.pgroup = pgroup;
        this.keywidth = keywidth;
        this.vbitlenro = vbitlenro;
        this.ebitlenro = ebitlenro;
        this.prg = prg;
        this.rohash = rohash;
        this.width = width;
        this.statdist = statdist;
        this.auxsid = default_auxsid;
        this.type = default_type;
    }

    /**
     * Initialize the ProtocolInformation from a file.
     * 
     * @param protinfo Protocol information file
     * @throws ShuffleException When parsing fails
     */
    public ProtocolInformation(Path protinfo) throws ShuffleException {
        DocumentBuilderFactory dbf = DocumentBuilderFactory.newInstance();
        DocumentBuilder db;
        Document dom;
        try {
            db = dbf.newDocumentBuilder();
        } catch (ParserConfigurationException e) {
            // there should be a default parser
            throw new RuntimeException("XML parser not configured", e);
        }
        try {
            dom = db.parse(protinfo.toFile());
        } catch (SAXException e) {
            throw new IllegalArgumentException("Invalid Protocol Information file", e);
        } catch (IOException e) {
            throw new IllegalArgumentException("Error while reading Protocol Information file", e);
        }
        Element root = dom.getDocumentElement();
        root.normalize();
        if (!root.getNodeName().equals("protocol")) {
            throw new IllegalArgumentException("Root node must be 'protocol'");
        }
        if (root.getNodeType() != Node.ELEMENT_NODE) {
            throw new IllegalArgumentException("Protocol node must be an element node");
        }
        Element rootel = (Element) root;
        this.version = get_element(rootel, "version");
        this.sid = get_element(rootel, "sid");
        this.name = get_element(rootel, "name");
        this.pgroup = get_element(rootel, "pgroup");
        this.generator_parsed = DataParser.parseGroupGenerator(this.pgroup);
        this.pgroup_parsed = this.generator_parsed.getGroup();
        this.keywidth = Integer.parseInt(get_element(rootel, "keywidth"));
        if (this.keywidth != default_keywidth) {
            throw new IllegalArgumentException("Invalid keywidth");
        }
        this.vbitlenro = Integer.parseInt(get_element(rootel, "vbitlenro"));
        this.ebitlenro = Integer.parseInt(get_element(rootel, "ebitlenro"));
        this.prg = get_element(rootel, "prg");
        this.rohash = get_element(rootel, "rohash");
        this.width = Integer.parseInt(get_element(rootel, "width"));
        this.statdist = Integer.parseInt(get_element(rootel, "statdist"));
        this.auxsid = default_auxsid;
        this.type = default_type;
        String corr = get_element(rootel, "corr");
        if (!corr.equals("noninteractive")) {
            throw new IllegalArgumentException("Only non-interactive protocol is supported");
        }
    }

    private static String get_element(Element node, String name) {
        NodeList els = node.getElementsByTagName(name);
        String val = null;
        if (els.getLength() < 1) {
            throw new IllegalArgumentException(String.format("Element '%s' missing", name));
        }
        if (els.getLength() > 1) {
            for (int i = 0; i < els.getLength(); i++) {
                if (els.item(i).getParentNode().equals(node)) {
                    val = els.item(i).getTextContent();
                }
            }
            if (val == null) {
                throw new IllegalArgumentException(String.format("Element '%s' not found", name));
            }
        } else {
            val = els.item(0).getTextContent();
        }
        return val;
    }

    public String get_version() {
        return version;
    }

    public String get_sid() {
        return sid;
    }

    public String get_auxsid() {
        return auxsid;
    }

    public String get_name() {
        return name;
    }

    public String get_pgroup() {
        return pgroup;
    }

    public Group get_parsed_pgroup() {
        return pgroup_parsed;
    }

    public GroupElement get_parsed_generator() {
        return generator_parsed;
    }

    public int get_keywidth() {
        return keywidth;
    }

    public int get_vbitlenro() {
        return vbitlenro;
    }

    public int get_ebitlenro() {
        return ebitlenro;
    }

    public String get_prg() {
        return prg;
    }

    public String get_rohash() {
        return rohash;
    }

    public int get_width() {
        return width;
    }

    public int get_statdist() {
        return statdist;
    }

    public String get_type() {
        return type;
    }
}
