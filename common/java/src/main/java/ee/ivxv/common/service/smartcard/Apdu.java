package ee.ivxv.common.service.smartcard;

import ee.ivxv.common.util.I18nConsole;
import ee.ivxv.common.util.Util;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.nio.ByteBuffer;
import java.util.Arrays;
import javax.smartcardio.CardChannel;
import javax.smartcardio.CardException;
import javax.smartcardio.CommandAPDU;
import javax.smartcardio.ResponseAPDU;
import javax.xml.bind.DatatypeConverter;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class Apdu {
    private static final Logger log = LoggerFactory.getLogger(Apdu.class);
    private static final int APDU_RESPONSE_CODE_SUCCESS = 36864; // 0x9000
    private static final int APDU_RESPONSE_CODE_PIN_REQUIRED = 27010; // 0x6982

    private static final byte FEATURE_VERIFY_PIN_DIRECT = 0x06;
    private static final int IOCTL_GET_FEATURE_REQUEST = scardCtlCode(3400);

    private final CardChannel channel;
    private final I18nConsole console;
    private final String cardId;

    private enum Instruction {
        CREATE_FILE(0xE0, "CREATE FILE"), READ_BINARY(0xB0, "READ BINARY"), SELECT_FILE(0xA4,
                "SELECT FILE"), UPDATE_BINARY(0xD6, "UPDATE BINARY"), VERIFY(0x20, "VERIFY");

        private final int code;
        private final String str;

        Instruction(int code, String str) {
            this.code = code;
            this.str = str;
        }

        public int getCode() {
            return code;
        }

        @Override
        public String toString() {
            return str;
        }
    }

    public Apdu(CardChannel channel, I18nConsole console, String cardId) {
        this.channel = channel;
        this.console = console;
        this.cardId = cardId;
    }

    private static int scardCtlCode(int code) {
        int ioctl;
        String os_name = System.getProperty("os.name").toLowerCase();
        if (os_name.contains("windows")) {
            ioctl = (0x31 << 16 | (code) << 2);
        } else {
            ioctl = 0x42000000 + (code);
        }
        return ioctl;
    }

    /**
     * Create new file on the card
     *
     * @param data Apdu data field value
     */
    public ResponseAPDU createFile(byte[] data) throws CardException {
        return transmit(Instruction.CREATE_FILE, 0, 0, data, 0);
    }

    /**
     * Read the selected file
     *
     * @param len number of bytes to read
     * @return Read bytes
     * @throws CardException
     */
    public byte[] readBinary(int len) throws CardException {
        int offset = 0;
        int toRead = len < 0xFF ? len : 0xFF;
        ByteArrayOutputStream outputStream = new ByteArrayOutputStream(len);
        byte[] data;
        do {
            ResponseAPDU r = readBinaryChunk(offset, toRead);
            data = r.getData();
            try {
                outputStream.write(data);
            } catch (IOException e) {
                throw new RuntimeException(e);
            }
            offset += data.length;
            toRead = len - offset < 0xFF ? len - offset : 0xFF;
        } while (toRead != 0);
        byte[] res = outputStream.toByteArray();
        log.debug("Data read: {}", DatatypeConverter.printHexBinary(res));
        return outputStream.toByteArray();
    }

    private ResponseAPDU readBinaryChunk(int offset, int len) throws CardException {
        return transmit(Instruction.READ_BINARY, offset >> 8, offset & 0xFF, null, len);
    }

    /**
     * Select file by path relative from MF
     *
     * @param path path without the identifier of the MF
     * @return Response of the select command
     * @throws CardException
     */
    public ResponseAPDU selectFile(byte[] path) throws CardException {
        if (path[0] == 0x3f && path[1] == 0x00) {
            return transmit(Instruction.SELECT_FILE, 8, 0, Arrays.copyOfRange(path, 2, path.length),
                    0);
        } else {
            return transmit(Instruction.SELECT_FILE, 8, 0, path, 0);
        }
    }

    public void updateBinary(byte[] data, int offset) throws CardException {
        int len = data.length;
        int dataOffset = 0;
        int toWrite = len < 0xFF ? len : 0xFF;
        ByteBuffer buf = ByteBuffer.wrap(data);
        do {
            byte[] chunk = new byte[toWrite];
            buf.get(chunk);

            updateBinaryChunk(offset, chunk);

            dataOffset += toWrite;
            offset += toWrite;
            toWrite = len - dataOffset < 0xFF ? len - dataOffset : 0xFF;
        } while (toWrite != 0);
    }

    private ResponseAPDU updateBinaryChunk(int offset, byte[] chunk) throws CardException {
        return transmit(Instruction.UPDATE_BINARY, offset >> 8, offset & 0xFF, chunk, 0);
    }

    private ResponseAPDU transmit(Instruction ins, int p1, int p2, byte[] data, int le)
            throws CardException {
        return transmit(ins, p1, p2, data, le, 2);
    }

    private ResponseAPDU transmit(Instruction ins, int p1, int p2, byte[] data, int le,
            int retryCount) throws CardException {
        CommandAPDU apdu = new CommandAPDU(0, ins.getCode(), p1, p2, data, le);
        log.debug("{}: {}", ins, DatatypeConverter.printHexBinary(apdu.getBytes()));
        ResponseAPDU r = channel.transmit(apdu);
        try {
            if (r.getSW() == APDU_RESPONSE_CODE_PIN_REQUIRED) {
                verify();
                transmit(ins, p1, p2, data, le, retryCount);
            } else if (r.getSW() != APDU_RESPONSE_CODE_SUCCESS && retryCount > 0) {
                log.debug("unsuccessful transmit, retying({}): {}", retryCount,
                        Integer.toHexString(r.getSW()));
                transmit(ins, p1, p2, data, le, retryCount - 1);
            } else {
                checkResponseStatus(ins, r);
            }
        } catch (SmartCardException e) {
            throw new CardException(e);
        }
        return r;
    }

    private void verify() throws CardException, SmartCardException {
        javax.smartcardio.Card card = channel.getCard();
        byte[] resp = null;
        boolean usePinPad = false;
        int i = 0;
        try {
            resp = card.transmitControlCommand(IOCTL_GET_FEATURE_REQUEST, new byte[0]);
            for (; i < resp.length; i += 6) {
                if (resp[i] == FEATURE_VERIFY_PIN_DIRECT) {
                    usePinPad = true;
                    break;
                }
            }
        }
        catch (CardException e) {
        //    console.log(e);
        }

        byte[] apdu = new byte[] {0x00, // CLA
                0x20, // Verify INS
                0x00, // P1
                0x01, // Reference number of the PIN to verify against
                0x08, // PIN length
                (byte) 0xFF, (byte) 0xFF, (byte) 0xFF, (byte) 0xFF, // will be replaced with real pw
                (byte) 0xFF, (byte) 0xFF, (byte) 0xFF, (byte) 0xFF,};
        if (usePinPad) {
            int verify_ioctl = (0xff & resp[i + 2]) << 24 | (0xff & resp[i + 3]) << 16
                    | (0xff & resp[i + 4]) << 8 | 0xff & resp[i + 5];
            verifyPinPad(verify_ioctl, apdu);
        } else {
            verifyKeyboard(apdu);
        }
    }

    private void verifyPinPad(int ioctl, byte[] apdu) throws CardException, SmartCardException {
        byte[] cmd = Util.concatAll(constructPINVerifyStructure(), apdu);
        boolean success;
        do {
            console.println(Msg.enter_pin_pinpad, cardId);
            byte[] resp;
            int retryCount = 3;
            while (true) {
                try {
                    resp = channel.getCard().transmitControlCommand(ioctl, cmd);
                    break;
                } catch (CardException e) {
                    if (retryCount-- == 0) {
                        throw e;
                    }
                }
            }
            success = processVerifyResponse(DatatypeConverter.printHexBinary(resp));
        } while (!success);
    }

    private void verifyKeyboard(byte[] apdu) throws CardException, SmartCardException {
        boolean success;
        do {
            console.println(Msg.enter_pin_keyboard, cardId);
            byte[] pw = console.console.readPw().getBytes();
            System.arraycopy(pw, 0, apdu, 5, pw.length);
            ResponseAPDU r = channel.transmit(new CommandAPDU(apdu));
            success = processVerifyResponse(Integer.toHexString(r.getSW()));
        } while (!success);
    }

    private boolean processVerifyResponse(String respHex) throws SmartCardException {
        respHex = respHex.toUpperCase();
        if (respHex.equals("9000")) {
            return true;
        } else if (respHex.substring(0, 3).equals("63C")) {
            console.println(Msg.verify_fail_retry, respHex.substring(3));
            return false;
        } else if (respHex.equals("6983")) {
            throw new SmartCardException(
                    "Verification failed and no number of retries left - PIN blocked");
        } else if (respHex.equals("6985")) {
            throw new SmartCardException("PIN locked - CONDITIONS NOT SATISFIED");
        } else {
            throw new SmartCardException("Unexpected response from VERIFY PIN: " + respHex);
        }
    }

    private byte[] constructPINVerifyStructure() {
        return new byte[] {0x1E, // bTimeOut (30 sec)
                0x00, // bTimeOut2
                (byte) 0x82, // bmFormatString
                0x04, // bmPINBlockString
                0x00, // bmPINLengthFormat
                0x08, // max PIN length
                0x04, // min PIN length
                0x02, // bEntryValidationCondition
                0x01, // bNumberMessage
                0x09, 0x04, // wLangId (English)
                0x00, // bMsgIndex
                0x00, 0x00, 0x00, // bTeoPrologue
                0x0D, 0x00, 0x00, 0x00 // ulDataLength
        };
    }

    private void checkResponseStatus(Instruction command, ResponseAPDU r)
            throws SmartCardException {
        if (r.getSW() != APDU_RESPONSE_CODE_SUCCESS) {
            log.debug("BADRESPONSE: {}", DatatypeConverter.printHexBinary(r.getBytes()));
            String code = Integer.toHexString(r.getSW());
            throw new SmartCardException(
                    String.format("Non-successful response code to %s: %s", command, code));
        }
    }

}
