package ee.ivxv.common.service.smartcard.pkcs15;

import ee.ivxv.common.service.smartcard.Apdu;
import ee.ivxv.common.service.smartcard.CardInfo;
import ee.ivxv.common.service.smartcard.IndexedBlob;
import ee.ivxv.common.service.smartcard.SmartCardException;
import ee.ivxv.common.service.smartcard.TerminalUtil;
import ee.ivxv.common.util.I18nConsole;
import ee.ivxv.common.util.Util;
import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.math.BigInteger;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import javax.smartcardio.Card;
import javax.smartcardio.CardChannel;
import javax.smartcardio.CardException;
import javax.smartcardio.CardTerminal;
import javax.smartcardio.ResponseAPDU;
import javax.xml.bind.DatatypeConverter;

import org.bouncycastle.asn1.ASN1Encodable;
import org.bouncycastle.asn1.ASN1EncodableVector;
import org.bouncycastle.asn1.ASN1InputStream;
import org.bouncycastle.asn1.ASN1OctetString;
import org.bouncycastle.asn1.ASN1Primitive;
import org.bouncycastle.asn1.ASN1TaggedObject;
import org.bouncycastle.asn1.DERApplicationSpecific;
import org.bouncycastle.asn1.DERBitString;
import org.bouncycastle.asn1.DEROctetString;
import org.bouncycastle.asn1.DERSequence;
import org.bouncycastle.asn1.DERTaggedObject;
import org.bouncycastle.asn1.DERUTF8String;
import org.bouncycastle.asn1.DLSequence;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * PKCS15Card implements a card with PKCS15 file system.
 */
public class PKCS15Card implements ee.ivxv.common.service.smartcard.Card {
    private static final Logger log = LoggerFactory.getLogger(PKCS15Card.class);
    private static final byte[] MF_PATH = new byte[] {0x3f, 0x00};
    private static final byte[] PKCS15_PATH = new byte[] {0x50, 0x15};
    private static final byte[] ODF_PATH = Util.concatAll(PKCS15_PATH, new byte[] {0x50, 0x31});
    private static final byte[] DEF_DF_PATH = new byte[] {0x45, 0x01};
    private static final int RETRY_DELAY_SEC = 5;
    private static final int RETRY_COUNT = 3;
    private int retryCount;

    protected final I18nConsole console;
    private String id;
    private Card card;
    private CardChannel cardChannel;
    private int termNo = -1;
    private Apdu apdu;

    /**
     * Initialize using a card identifier and a console.
     *
     * @param id
     * @param console
     */
    public PKCS15Card(String id, I18nConsole console) {
        this(id, -1, console);
    }

    /**
     * Initialize using a card identifier, terminal number and a console.
     *
     * @param id
     * @param termNo
     * @param console
     */
    private PKCS15Card(String id, int termNo, I18nConsole console) {
        this.id = id;
        // store the terminal number for later use. do not open the channel
        // just yet
        this.termNo = termNo;
        this.console = console;
    }

    /**
     * Get all available PKCS15Cards.
     *
     * @param console
     * @return
     * @throws CardException
     */
    public static PKCS15Card[] getCards(I18nConsole console) throws CardException {
        List<CardTerminal> terminals = TerminalUtil.getTerminals();
        List<PKCS15Card> cardList = new ArrayList<>();
        for (int i = 0; i < terminals.size(); i++) {
            CardTerminal ct = terminals.get(i);
            if (ct.isCardPresent()) {
                cardList.add(new PKCS15Card(String.valueOf(i), i, console));
            }
        }
        PKCS15Card[] ret = new PKCS15Card[cardList.size()];
        return cardList.toArray(ret);
    }

    private static boolean comparePathRoot(byte[] first, byte[] second) {
        int l = first.length > second.length ? first.length : second.length;
        for (int i = 0; i < l - 1; i++) {
            if (first[i] != second[i]) {
                return false;
            }
        }
        return true;
    }

    private void retryThrow(CardException t, int retryThreshold) throws PKCS15Exception {
        if (t.getCause() != null && t.getCause().getMessage().equals("SCARD_W_RESET_CARD")) {
            initialize();
        }
        retryThrow(new PKCS15Exception("Communication with card failed", t), retryThreshold);
    }

    private void retryThrow(PKCS15Exception t, int retryThreshold) throws PKCS15Exception {
        if (retryCount >= retryThreshold) {
            log.debug("Retry threshold reached: {}/{}", retryCount, retryThreshold);
            retryCount = 0;
            throw t;
        }
        log.debug("Caught exception. Waiting and retrying. Try {}/{}", retryCount, retryThreshold);
        retryCount++;
        try {
            Thread.sleep(1000 * RETRY_DELAY_SEC);
        } catch (InterruptedException e) {
            // single-threaded application
            log.error("Retry interrupted", e);
        }
    }

    private static ASN1Primitive[] readObjects(byte[] bytes) throws IOException {
        ByteArrayInputStream bs = new ByteArrayInputStream(bytes);
        ArrayList<ASN1Primitive> res = new ArrayList<>();
        try (ASN1InputStream input = new ASN1InputStream(bs, false)) {
            int b;
            while (true) {
                bs.mark(0); // ahead does not affect methods
                b = bs.read();
                if (b == 0) {
                    // we have padding
                    break;
                } else if (b == -1) {
                    // stream end
                    break;
                }
                bs.reset();
                ASN1Primitive obj = input.readObject();
                if (obj == null) {
                    // end of meaningful data in the stream
                    break;
                }
                res.add(obj);
            }
        }
        return res.toArray(new ASN1Primitive[] {});
    }

    @Override
    public boolean isInitialized() {
        if (cardChannel == null) {
            return false;
        }
        try {
            cardChannel.getChannelNumber();
        } catch (IllegalStateException e) {
            return false;
        }
        return true;
    }

    @Override
    public void initialize() throws PKCS15Exception {
        List<CardTerminal> list;
        try {
            list = TerminalUtil.getTerminals();
        } catch (Exception e) {
            throw new PKCS15Exception("Could not get card teminal list", e);
        }
        try {
            card = list.get(termNo).connect("*");
        } catch (Exception e) {
            throw new PKCS15Exception("Could not get connect to card", e);
        }
        cardChannel = card.getBasicChannel();
        apdu = new Apdu(cardChannel, console);
        try {
            CardInfo info = getCardInfo();
            id = info == null ? id : info.getId();
        } catch (SmartCardException e) {
            throw new PKCS15Exception("Could not read from the card", e);
        }
        apdu.setId(getId());

    }

    @Override
    public void close() throws PKCS15Exception {
        if (!isInitialized()) {
            return;
        }
        try {
            // Card works seemingly the same regardless how/whether we
            // close the connection.
            card.disconnect(true);
        } catch (CardException e) {
            throw new PKCS15Exception("Could not close connection to card", e);
        }
        cardChannel = null;
        apdu = null;
    }

    @Override
    public String getId() {
        return id;
    }

    /*
     * interface-specific methods start here
     */

    /**
     * Erase the file system.
     * <p>
     * Currently erasing the file system is not implemented and this method is a no-op.
     *
     * @throws PKCS15Exception
     */
    public void eraseFilesystem() throws PKCS15Exception {
        // not supported
    }

    /**
     * Create a PKCS15 file system.
     * <p>
     * Currently erasing the file system is not implemented and this method is a no-op.
     *
     * @throws PKCS15Exception
     */
    public void createFilesystem() throws PKCS15Exception {
        // not supported
    }

    @Override
    public void storeIndexedBlob(byte[] aid, byte[] identifier, byte[] blob, int index)
            throws SmartCardException {
        storeBlobRetry(aid, identifier, new PKCS15IndexedBlob(index, blob).encode(), RETRY_COUNT);
    }

    @Override
    public void storeCardInfo(CardInfo data) throws SmartCardException {
        storeBlobRetry(null, CardInfo.IDENTIFIER, data.encode(), RETRY_COUNT);
    }

    private DLSequence[] getDODF() throws CardException, IOException, PKCS15Exception {
        short len = selectDODF();
        byte[] dodfBytes = apdu.readBinary(len);
        // Get the metainfo of objects written to the card
        ASN1Primitive[] vals = readObjects(dodfBytes);
        DLSequence[] res = new DLSequence[vals.length];
        for (int i = 0; i < vals.length; i++) {
            DLSequence v = (DLSequence) vals[i];
            if (v.size() != 3) {
                throw new PKCS15Exception("DODF entry corrupted");
            }
            if (((DLSequence) v.getObjectAt(1)).size() != 1) {
                throw new PKCS15Exception("DODF entry corrupted");
            }
            res[i] = (DLSequence) vals[i];
        }
        return res;
    }

    private byte[] getAvailablePath(DLSequence[] blobmetas) throws PKCS15Exception {
        byte[] path = Util.concatAll(PKCS15_PATH, DEF_DF_PATH);
        if (blobmetas.length == 0) {
            return path;
        }
        byte[] objpath;
        for (DLSequence obj : blobmetas) {
            objpath = getPath((ASN1TaggedObject) obj.getObjectAt(2));
            objpath = getRelativePath(objpath);
            // first case is when there are no free paths left
            if (objpath[objpath.length - 1] == (byte) 0xFF) {
                throw new PKCS15Exception("No free path on card");
            }
            // we are only interested in paths having specific root
            if (!comparePathRoot(path, objpath)) {
                continue;
            }
            // otherwise we take such path which is at least one higher than any
            // existing path
            if (objpath[objpath.length - 1] >= path[path.length - 1]) {
                path[path.length - 1] = (byte) (objpath[objpath.length - 1] + 1);
            }
        }
        return path;
    }

    private byte[] getRelativePath(byte[] objpath) {
        if (objpath[0] == MF_PATH[0] && objpath[1] == MF_PATH[1]) {
            return Arrays.copyOfRange(objpath, 2, objpath.length);
        }
        return objpath;
    }

    private void createNewFile(byte[] path, byte[] blob, boolean readAuth) throws CardException {
        // Select blob's parent DF
        apdu.selectFile(Arrays.copyOfRange(path, 0, path.length - 2));
        // Create new file to store blob in
        short len = (short) blob.length;
        byte[] relatPath = Arrays.copyOfRange(path, path.length - 2, path.length);
        byte[] createFileData = constructCreateFileApduData(relatPath, len, readAuth);
        apdu.createFile(createFileData);
    }

    private void writeFile(byte[] path, byte[] blob) throws CardException {
        // Select created file
        apdu.selectFile(path);
        // Update file with blob data
        apdu.updateBinary(blob, 0);
    }

    private void updateDODF(byte[] path, byte[] aid, byte[] identifier, DLSequence[] existing)
            throws CardException, PKCS15Exception, IOException {
        // Create new dodf entry
        byte[] dodfEntry = createDodfEntry(Util.concatAll(MF_PATH, path), aid, identifier);
        if (existing == null) {
            existing = getDODF();
        } else {
            selectDODF();
        }
        int offset = 0;
        for (DLSequence v : existing) {
            try {
                offset += v.getEncoded("DER").length;
            } catch (IOException e) {
                // does not happen. the input is controlled
            }
        }
        apdu.updateBinary(dodfEntry, offset);
    }

    /**
     * Store a blob on the card.
     *
     * @param aid Authentication identifier to associate with the file.
     * @param identifier File identifier.
     * @param blob File data.
     * @throws PKCS15Exception
     */
    public void storeBlob(byte[] aid, byte[] identifier, byte[] blob) throws PKCS15Exception {
        storeBlobRetry(aid, identifier, blob, 0);
    }

    // store a data object 'blob' at a location 'identifier', protected by
    // authentication ID 'aid'
    private void storeBlobRetry(byte[] aid, byte[] identifier, byte[] blob, int retryThreshold)
            throws PKCS15Exception {
        if (!isInitialized()) {
            throw new PKCS15Exception("Card not initialized");
        }
        boolean readAuth = aid != null;
        DLSequence[] bm;
        for (retryCount = 0; true;) {
            try {
                bm = getDODF();
                break;
            } catch (IOException e) {
                retryThrow(new PKCS15Exception("Metainfo corrupted"), retryThreshold);
            } catch (CardException e) {
                retryThrow(e, retryThreshold);
            }
        }
        if (findDF(bm, identifier) != null) {
            throw new PKCS15Exception("Blob with identifier already exists");
        }
        byte[] path;
        for (retryCount = 0; true;) {
            try {
                path = getAvailablePath(bm);
                break;
            } catch (PKCS15Exception e) {
                retryThrow(e, retryThreshold);
            }
        }
        log.debug("Blob path: {}", DatatypeConverter.printHexBinary(path));
        for (retryCount = 0; true;) {
            try {
                createNewFile(path, blob, readAuth);
                break;
            } catch (CardException e) {
                retryThrow(e, retryThreshold);
            }
        }
        for (retryCount = 0; true;) {
            try {
                writeFile(path, blob);
                break;
            } catch (CardException e) {
                retryThrow(e, retryThreshold);

            }
        }
        for (retryCount = 0; true;) {
            try {
                updateDODF(path, aid, identifier, bm);
                break;
            } catch (PKCS15Exception e) {
                retryThrow(e, retryThreshold);
            } catch (CardException e) {
                retryThrow(e, retryThreshold);
            } catch (IOException e) {
                // does not happen, we provide bm and then the path where
                // IOException happens is not taken
            }
        }
    }

    /**
     * Remove the data object at a location protected with an authentication identifier.
     * <p>
     * The implementation does not support removing files. It returns false on every call.
     *
     * @param aid Authentication identifier
     * @param identifier File identifier
     * @return Success of removing the blob.
     * @throws PKCS15Exception
     */
    public boolean removeBlob(byte[] aid, byte[] identifier) throws PKCS15Exception {
        return false;
    }

    @Override
    public IndexedBlob getIndexedBlob(byte[] aid, byte[] identifier) throws SmartCardException {
        return PKCS15IndexedBlob.create(getBlob(aid, identifier));
    }

    @Override
    public CardInfo getCardInfo() throws PKCS15Exception, SmartCardException {
        try {
            return CardInfo.create(getBlob(null, CardInfo.IDENTIFIER));
        } catch (PKCS15Exception e) {
            if (e.getMessage().equals("File does not exist")) {
                return null;
            }
            throw e;
        }
    }

    /**
     * Get the blob with authentication identifier.
     *
     * @param aid Authentication identifier.
     * @param identifier File identifier.
     * @return File content.
     * @throws PKCS15Exception
     */
    public byte[] getBlob(byte[] aid, byte[] identifier) throws PKCS15Exception {
        return getBlobRetry(aid, identifier, RETRY_COUNT);
    }

    private byte[] getBlobRetry(byte[] aid, byte[] identifier, int retryThreshold)
            throws PKCS15Exception {
        if (!isInitialized()) {
            throw new PKCS15Exception("Card not initialized");
        }
        DLSequence[] bm;
        ResponseAPDU r;
        DLSequence seq;
        int len;
        byte[] blob, path;
        for (retryCount = 0; true;) {
            try {
                bm = getDODF();
                break;
            } catch (IOException e) {
                retryThrow(new PKCS15Exception("Metainfo corrupted"), retryThreshold);
            } catch (CardException e) {
                retryThrow(e, retryThreshold);
            }
        }
        seq = findDF(bm, identifier);
        if (seq == null) {
            throw new PKCS15Exception("File does not exist");
        }
        // verify aid
        if (aid != null) {
            verifyAid((DLSequence) seq.getObjectAt(0), aid);
        }
        // get path of blob data
        path = getPath((DERTaggedObject) seq.getObjectAt(2));
        // read the blob
        for (retryCount = 0; true;) {
            try {
                r = apdu.selectFile(path);
                len = getFileLen(r.getData());
                blob = apdu.readBinary(len);
                break;
            } catch (CardException e) {
                retryThrow(e, retryThreshold);
            } catch (IOException e) {
                retryThrow(new PKCS15Exception("Card response corrupted"), retryThreshold);
            }
        }
        return blob;
    }

    @Override
    public int getTerminal() {
        return termNo;
    }

    @Override
    public void setTerminal(int terminalNo) {
        this.termNo = terminalNo;
    }

    private DLSequence findDF(DLSequence[] DFs, byte[] identifier) {
        DLSequence idSeq;
        byte[] eid;
        for (DLSequence d : DFs) {
            idSeq = (DLSequence) d.getObjectAt(1);
            eid = ((DERUTF8String) idSeq.getObjectAt(0)).getString().getBytes();
            if (Arrays.equals(eid, identifier)) {
                return d;
            }
        }
        return null;
    }

    private byte[] getPath(ASN1TaggedObject obj) throws PKCS15Exception {
        DERSequence seq = (DERSequence) obj.getObject();
        if (seq.size() != 1) {
            throw new PKCS15Exception("Invalid ASN1 structure");
        }
        ASN1Encodable el = seq.getObjectAt(0);
        if (el instanceof DERSequence) {
            // newer Aventra myEID driver creates different structure
            el = ((DERSequence) el).getObjectAt(0);
        }
        DEROctetString octetString = (DEROctetString) el;
        return octetString.getOctets();
    }

    private ASN1TaggedObject getObjByTagNo(byte[] bytes, int tagNo) throws IOException {
        ASN1Primitive[] vals = readObjects(bytes);
        ASN1TaggedObject vv;
        for (ASN1Primitive v : vals) {
            vv = (ASN1TaggedObject) v;
            if (vv.getTagNo() == tagNo) {
                return vv;
            }
        }
        return null;
    }

    private byte[] constructCreateFileApduData(byte[] identifier, short len, boolean readAuth) {
        byte authByte = readAuth ? (byte) 0x11 : (byte) 0x01;
        return new byte[] {0x62, // FCP tag
                0x17, // length
                (byte) 0x80, // File Size tag - transparent EF
                0x02, // length
                (byte) (len >> 8), (byte) len, // File Size value
                (byte) 0x82, // File Description tag
                0x01, // length
                0x01, // File Descriptor - transparent EF
                (byte) 0x83, // File Identifier tag
                0x02, // length
                identifier[0], identifier[1], // File Identifier value
                (byte) 0x86, // Security Attributes tag
                0x03, // length
                authByte, // Pin reference for Read and Update
                0x1F, // Pin reference for Delete
                (byte) 0xFF, // RFU
                (byte) 0x85, // Proprietary Information tag
                0x02, // length
                0x00, // RFU in case of EF
                0x00, // RFU
                (byte) 0x8A, // Life Cycle Status tag
                0x01, // length
                0x00, // RFU
        };
    }

    private byte[] createDodfEntry(byte[] path, byte[] aid, byte[] identifier) {
        ASN1EncodableVector root = new ASN1EncodableVector();

        ASN1EncodableVector v1 = new ASN1EncodableVector();
        if (aid != null) {
            v1.add(new DERBitString(new byte[] {(byte) 0x80}, 6));
            v1.add(new DEROctetString(aid));
        } else {
            v1.add(new DERBitString(new byte[] {(byte) 0x80}, 6));
        }

        ASN1EncodableVector v2 = new ASN1EncodableVector();
        v2.add(new DERUTF8String(new String(identifier)));

        ASN1EncodableVector v3 = new ASN1EncodableVector();
        v3.add(new DEROctetString(path));

        root.add(new DERSequence(v1));
        root.add(new DERSequence(v2));
        root.add(new DERTaggedObject(1, new DERSequence(v3)));

        byte[] res;
        try {
            res = new DERSequence(root).getEncoded("DER");
        } catch (IOException e) {
            // if exception is thrown here then this is a programming error.
            // these error must be fixed during development phase and thus they
            // are not thrown in production.
            res = null;
        }
        return res;
    }

    private short getFileLen(byte[] bytes) throws IOException {
        ASN1InputStream stream = new ASN1InputStream(new ByteArrayInputStream(bytes));
        DERApplicationSpecific parent = (DERApplicationSpecific) stream.readObject();
        stream = new ASN1InputStream(new ByteArrayInputStream(parent.getContents()));
        ASN1TaggedObject obj;
        do {
            obj = (ASN1TaggedObject) stream.readObject();
        } while (obj.getTagNo() != 0);
        DEROctetString octetString = (DEROctetString) obj.getObject();
        return new BigInteger(1, octetString.getOctets()).shortValueExact();
    }

    private short selectDODF() throws CardException, PKCS15Exception {
        // read Object Directory File (ODF)
        byte[] bytes;
        ResponseAPDU r;
        short len;
        r = apdu.selectFile(ODF_PATH);
        try {
            len = getFileLen(r.getData());
        } catch (IOException e) {
            throw new PKCS15Exception("Card ODF file corrupted");
        }
        bytes = apdu.readBinary(len);
        // find object with Data Objects tag (7)
        ASN1TaggedObject obj;
        try {
            obj = getObjByTagNo(bytes, 7);
        } catch (IOException e) {
            throw new PKCS15Exception("Card ODF entry corrupted");
        }
        if (obj == null) {
            throw new PKCS15Exception("Card does not have DODF object");
        }
        // get Data Object Directory File (DODF) path
        byte[] path = getPath(obj);
        // read DODF
        r = apdu.selectFile(path);
        try {
            len = getFileLen(r.getData());
        } catch (IOException e) {
            throw new PKCS15Exception("Card DODF file corrupted");
        }
        return len;
    }

    private void verifyAid(DLSequence parent, byte[] aid) throws PKCS15Exception {
        ASN1OctetString octetString = (ASN1OctetString) parent.getObjectAt(1);
        byte[] foundAid = octetString.getOctets();
        if (!Arrays.equals(foundAid, aid)) {
            throw new PKCS15Exception(String.format("Aid mismatch! required: %s, found: %s",
                    DatatypeConverter.printHexBinary(aid),
                    DatatypeConverter.printHexBinary(foundAid)));
        }
    }

}
