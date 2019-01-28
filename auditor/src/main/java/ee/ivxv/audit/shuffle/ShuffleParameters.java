package ee.ivxv.audit.shuffle;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.List;

public class ShuffleParameters {
    private String version, auxsid, type;
    int width;

    /**
     * Initialize the shuffle parameters from values.
     * 
     * @param version
     * @param auxsid
     * @param width
     * @param type
     */
    ShuffleParameters(String version, String auxsid, int width, String type) {
        this.version = version;
        this.auxsid = auxsid;
        this.width = width;
        this.type = type;
    }

    /**
     * Initialize ShuffleParameters using a proof directory path. The corresponding files are
     * assumed to be at their default locations (stored by Verificatum).
     * 
     * @param proofdir Proof directory
     * @throws IOException If reading parameter file fails
     */
    ShuffleParameters(Path proofdir) throws IOException {
        Path versionpath = Paths.get(proofdir.toString(), "version");
        Path auxsidpath = Paths.get(proofdir.toString(), "auxsid");
        Path widthpath = Paths.get(proofdir.toString(), "width");
        Path typepath = Paths.get(proofdir.toString(), "type");
        this.version = read_parameters(versionpath);
        this.auxsid = read_parameters(auxsidpath);
        this.width = Integer.parseInt(read_parameters(widthpath));
        this.type = read_parameters(typepath);

    }

    private static String read_parameters(Path loc) throws IOException {
        List<String> l = Files.readAllLines(loc);
        if (l.size() != 1) {
            throw new IllegalArgumentException("Parameters file must have single line");
        }
        return l.get(0);
    }

    public String get_version() {
        return version;
    }

    public String get_auxsid() {
        return auxsid;
    }

    public int get_width() {
        return width;
    }

    public String get_type() {
        return type;
    }
}
