package ee.ivxv.common.service.smartcard.dummy;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;
import ee.ivxv.common.service.console.Console;
import ee.ivxv.common.service.i18n.I18n;
import ee.ivxv.common.service.smartcard.Card;
import ee.ivxv.common.service.smartcard.CardService;
import ee.ivxv.common.service.smartcard.Cards;
import ee.ivxv.common.util.I18nConsole;
import ee.ivxv.common.util.Json;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.HashMap;
import java.util.Map;
import javax.xml.bind.DatatypeConverter;

/**
 * DummyCardService is a card service which uses JSON storage for content.
 */
public class DummyCardService implements CardService {
    static class DummyFilesystems {
        private Map<String, Map<String, byte[]>> filesystems;

        @JsonCreator
        private DummyFilesystems( //
                @JsonProperty("filesystems") Map<String, Map<String, byte[]>> filesystems) {
            this.filesystems = filesystems;
        }

        private DummyFilesystems() {
            this.filesystems = new HashMap<String, Map<String, byte[]>>();
        }

        @JsonIgnore
        private void createFilesystem(String id) {
            if (filesystems.get(id) == null) {
                filesystems.put(id, new HashMap<String, byte[]>());
            }
        }

        @JsonIgnore
        Map<String, byte[]> getFilesystem(String id) {
            return filesystems.get(id);
        }

        @JsonIgnore
        byte[] getFile(String id, byte[] path) {
            Map<String, byte[]> fs = filesystems.get(id);
            return fs == null ? null : fs.get(hex(path));
        }

        @JsonIgnore
        void putFile(String id, byte[] path, byte[] content) {
            filesystems.get(id).put(hex(path), content);
        }

        @JsonIgnore
        boolean removeFilesystem(String id) {
            boolean res = filesystems.remove(id) != null;
            return res;
        }

        @JsonIgnore
        boolean removeFile(String id, byte[] path) {
            Map<String, byte[]> fs = getFilesystem(id);
            if (fs == null) {
                return false;
            }
            boolean res = fs.remove(hex(path)) != null;
            return res;
        }

        public Map<String, Map<String, byte[]>> getFilesystems() {
            return filesystems;
        }

        @JsonIgnore
        private static String hex(byte[] in) {
            return DatatypeConverter.printHexBinary(in);
        }
    }

    DummyFilesystems fses;

    /**
     * The environment variable to read the dummy card filesystems path.
     */
    public static final String ENV_CARD_FS_PATH_VAR = "DUMMY_CARDS_PATH";
    /**
     * If the environment variable defined by the key {@#ENV_CARD_FS_PATH_VAR} does not hold the
     * path for dummy cards use the following default location.
     */
    public static final String DEFAULT_CARD_FS_PATH = "dummy_card_filesystems";

    private final I18nConsole console;
    private final Path cardFsPath;

    public DummyCardService(Console console, I18n i18n) {
        this(console, i18n, false);
    }

    public DummyCardService(Console console, I18n i18n, boolean persistent) {
        this.console = new I18nConsole(console, i18n);
        String envpath = System.getenv(ENV_CARD_FS_PATH_VAR);
        String path = envpath != null ? envpath : DEFAULT_CARD_FS_PATH;
        cardFsPath = persistent ? Paths.get(path) : null;
        if (cardFsPath != null && Files.exists(cardFsPath)) {
            try {
                fses = Json.read(cardFsPath, DummyFilesystems.class);
            } catch (Exception e) {
                throw new RuntimeException("Unable to read dummy filesystems", e);
            }
        }
        if (this.fses == null) {
            this.fses = new DummyFilesystems();
        }
    }

    @Override
    public Card createCard(String id) {
        fses.createFilesystem(id);
        return new DummyPKCS15Card(id, console, this);
    }

    @Override
    public Cards createCards() {
        Cards cards = new Cards(this, console) {
            @Override
            public Card getCard(int index) {
                return cards.get(index);
            }
        };
        return cards;
    }

    @Override
    public boolean isPluggableService() {
        return false;
    }

    /**
     * Commit the current storage to file.
     */
    protected void writeToFile() {
        if (cardFsPath == null) {
            return;
        }
        try {
            Json.write(fses, cardFsPath);
        } catch (Exception e) {
            throw new RuntimeException("Unable to write dummy filesystems", e);
        }
    }
}
