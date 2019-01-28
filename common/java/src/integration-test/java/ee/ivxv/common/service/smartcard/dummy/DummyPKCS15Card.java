package ee.ivxv.common.service.smartcard.dummy;

import ee.ivxv.common.service.smartcard.SmartCardException;
import ee.ivxv.common.service.smartcard.pkcs15.PKCS15Card;
import ee.ivxv.common.service.smartcard.pkcs15.PKCS15IndexedBlob;
import ee.ivxv.common.util.I18nConsole;

public class DummyPKCS15Card extends PKCS15Card {
    private DummyCardService cs;

    public DummyPKCS15Card(String id, I18nConsole console, DummyCardService cs) {
        super(id, console);
        this.cs = cs;
    }

    @Override
    public boolean isInitialized() {
        return true;
    }

    @Override
    public void eraseFilesystem() {
        cs.fses.removeFilesystem(getId());
    }

    @Override
    public void storeIndexedBlob(byte[] aid, byte[] identifier, byte[] blob, int index)
            throws SmartCardException {
        storeBlob(aid, identifier, new PKCS15IndexedBlob(index, blob).encode());
    }

    @Override
    public void storeBlob(byte[] aid, byte[] identifier, byte[] blob) {
        cs.fses.putFile(getId(), identifier, blob);
        cs.writeToFile();
    }

    @Override
    public boolean removeBlob(byte[] aid, byte[] identifier) {
        boolean res = cs.fses.removeFile(getId(), identifier);
        cs.writeToFile();
        return res;
    }

    @Override
    public byte[] getBlob(byte[] aid, byte[] identifier) {
        return cs.fses.getFile(getId(), identifier);
    }
}
