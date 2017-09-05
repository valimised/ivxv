package ee.ivxv.common.service.smartcard.dummy;

import ee.ivxv.common.service.smartcard.SmartCardException;
import ee.ivxv.common.service.smartcard.pkcs15.PKCS15Card;
import ee.ivxv.common.service.smartcard.pkcs15.PKCS15IndexedBlob;
import ee.ivxv.common.util.I18nConsole;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.ObjectInputStream;
import java.io.ObjectOutputStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.HashMap;
import javax.xml.bind.DatatypeConverter;

public class DummyPKCS15Card extends PKCS15Card {

    private HashMap<String, byte[]> fs;
    private Path cardFile;

    public DummyPKCS15Card(String id, I18nConsole console, Path cardDir) {
        super(id, console);
        if (cardDir != null) {
            cardFile = cardDir.resolve(getId());
        }
        createFilesystem();
    }

    @Override
    public boolean isInitialized() {
        return true;
    }

    @Override
    public void eraseFilesystem() {
        fs.clear();
    }

    @SuppressWarnings("unchecked")
    @Override
    public void createFilesystem() {
        if (cardFile != null && Files.exists(cardFile)) {
            try {
                InputStream in = Files.newInputStream(cardFile);
                ObjectInputStream oin = new ObjectInputStream(in);
                fs = (HashMap<String, byte[]>) oin.readObject();
            } catch (IOException | ClassNotFoundException e) {
                throw new RuntimeException("Error reading dummycard from file", e);
            }
        } else {
            fs = new HashMap<>();
        }
    }

    @Override
    public void storeIndexedBlob(byte[] aid, byte[] identifier, byte[] blob, int index)
            throws SmartCardException {
        storeBlob(aid, identifier, new PKCS15IndexedBlob(index, blob).encode());
    }

    @Override
    public void storeBlob(byte[] aid, byte[] identifier, byte[] blob) {
        fs.put(DatatypeConverter.printHexBinary(identifier), blob);
        writeToFile();
    }

    @Override
    public boolean removeBlob(byte[] aid, byte[] identifier) {
        boolean res = fs.remove(identifier) != null;
        writeToFile();
        return res;
    }

    @Override
    public byte[] getBlob(byte[] aid, byte[] identifier) {
        return fs.get(DatatypeConverter.printHexBinary(identifier));
    }

    @Override
    public boolean isAvailable() {
        return true;
    }

    private void writeToFile() {
        if (cardFile == null) {
            return;
        }
        try {
            ByteArrayOutputStream bout = new ByteArrayOutputStream();
            ObjectOutputStream oout = new ObjectOutputStream(bout);
            oout.writeObject(fs);
            Files.write(cardFile, bout.toByteArray());
        } catch (IOException e) {
            throw new RuntimeException("Error when storing dummycard in file", e);
        }
    }
}
