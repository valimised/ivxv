package ee.ivxv.common.util;

import java.io.Closeable;
import java.io.IOException;
import java.io.OutputStream;
import java.util.ArrayList;
import java.util.List;
import org.apache.pdfbox.pdmodel.PDDocument;
import org.apache.pdfbox.pdmodel.PDPage;
import org.apache.pdfbox.pdmodel.PDPageContentStream;
import org.apache.pdfbox.pdmodel.common.PDRectangle;
import org.apache.pdfbox.pdmodel.font.PDFont;
import org.apache.pdfbox.pdmodel.font.PDType0Font;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class PdfDoc implements Closeable {

    private static final Logger log = LoggerFactory.getLogger(PdfDoc.class);

    private static final String FONT_PATH_REGULAR = "fonts/Roboto-Regular.ttf";
    private static final String FONT_PATH_BOLD = "fonts/Roboto-Bold.ttf";

    private static final float FONT_SIZE_NORMAL = 11;
    private static final float FONT_SIZE_TITLE = 16;
    private static final float LINE_SPACING_COEF = 1.2f;
    private static final float MARGIN_L = 50;
    private static final float MARGIN_R = 50;
    private static final float MARGIN_T = 40;
    private static final float MARGIN_B = 40;

    private final OutputStream out;
    private final PDDocument document;
    private final PDFont fontRegular;
    private final PDFont fontBold;

    private PDPage page;
    private PDPageContentStream cs;
    private PDFont font;
    private float fontSize;
    private float leading;
    private float offsetY;
    private float tab;
    private int pageNumber = 0;

    public PdfDoc(OutputStream out) throws IOException {
        this.out = out;
        document = new PDDocument();
        fontRegular = PDType0Font.load(document, Util.getResource(FONT_PATH_REGULAR));
        fontBold = PDType0Font.load(document, Util.getResource(FONT_PATH_BOLD));
        font = fontRegular;
        fontSize = FONT_SIZE_NORMAL;
        startPage();
    }

    private static List<String> safeString(Object o, PDFont font) throws IOException {
        List<String> lines = new ArrayList<>();

        String s = o == null ? "" : String.valueOf(o);
        for (String line : s.split("\n")) {
            lines.add(safeString(line, font));
        }

        return lines;
    }

    private static String safeString(String s, PDFont font) throws IOException {
        try {
            // Check if the provided font supports the provided string.
            font.encode(s);
            return s;
        } catch (RuntimeException e) {
            // May be "U+FFFD is not available in this font's encoding". Fall through.
            log.error("Exception occurred while encoding string {}", s, e);
        }

        // Font does not support all characters in the string. Try 1-by-1 and replace unsupported.

        StringBuilder sb = new StringBuilder();
        for (int offset = 0; offset < s.length();) {
            int codePoint = s.codePointAt(offset);
            char[] chars = Character.toChars(codePoint);
            String u = new String(chars);

            try {
                font.encode(u);
            } catch (RuntimeException e) {
                // May be "U+FFFD is not available in this font's encoding". Replace character.
                log.error("Exception occurred while encoding string {}", s, e);
                u = "?";
            }
            sb.append(u);

            offset += Character.charCount(codePoint);
        }

        return sb.toString();
    }

    public void addText(Object o) throws IOException {
        setFont(fontRegular, FONT_SIZE_NORMAL);
        addTextFlow(o);
    }

    public void addText(Object o, float w, Alignment a) throws IOException {
        setFont(fontRegular, FONT_SIZE_NORMAL);
        addTextBox(o, normalizeTextWidth(w), a);
    }

    public void addTitle(Object o) throws IOException {
        setFont(fontBold, FONT_SIZE_TITLE);
        addTextFlow(o);
    }

    public void addTitle(Object o, float w, Alignment a) throws IOException {
        setFont(fontBold, FONT_SIZE_TITLE);
        addTextBox(o, normalizeTextWidth(w), a);
    }

    private float normalizeTextWidth(float w) {
        float maxWidth = page.getMediaBox().getWidth() - MARGIN_L - tab - MARGIN_R;
        return w <= 0 ? maxWidth : Math.min(w, maxWidth);
    }

    private void addTextBox(Object o, float w, Alignment a) throws IOException {
        List<String> lines = new ArrayList<>();

        for (String s : safeString(o, font)) {
            int lineStart = 0;

            for (int space = -1, lastSpace = -1; space < s.length(); lastSpace = space) {
                space = s.indexOf(' ', lastSpace + 1);
                space = space < 0 ? s.length() : space;
                // If line length exceeds the allowed width, split
                String current = s.substring(lineStart, space);
                float currentWidth = getWidth(current);
                if (currentWidth > w && lastSpace >= 0) {
                    lines.add(s.substring(lineStart, lastSpace));
                    lineStart = lastSpace + 1;
                }
            }

            lines.add(s.substring(lineStart));
        }

        float origTab = tab;
        boolean isFirst = true;
        for (String l : lines) {
            if (!isFirst) {
                // TODO Use internal newLine, remember the number of new lines and consider it in
                // public newLine method to avoid overwriting. Keep in mind page-breaks!
                newLine();
                tab(origTab);
            }
            isFirst = false;

            float currentTab = 0;
            float currentWidth = getWidth(l);
            if (a == Alignment.RIGHT) {
                currentTab = w - currentWidth;
            } else if (a == Alignment.CENTER) {
                currentTab = (w - currentWidth) / 2;
            }

            tab(currentTab);
            addTextFlow(l);
            tab(-currentTab);
        }
        // TODO restore offsetY
        // offsetY += (lines.size() - 1) * leading;
        // cs.newLineAtOffset(0, (lines.size() - 1) * leading);
    }

    private float getWidth(String s) throws IOException {
        return fontSize * font.getStringWidth(s) / 1000;
    }

    private void addTextFlow(Object o) throws IOException {
        boolean isFirst = true;
        for (String s : safeString(o, font)) {
            if (!isFirst) {
                newLine();
            }
            isFirst = false;
            cs.showText(s);
        }
    }

    public void newLine() throws IOException {
        if (offsetY - leading < MARGIN_B) {
            endPage();
            startPage();
        } else {
            cs.newLine();
            cs.newLineAtOffset(-1 * tab, 0);
            offsetY -= leading;
            tab = 0;
        }
    }

    public void newPage() throws IOException {
        endPage();
        startPage();
    }

    public void tab(float x) throws IOException {
        tab += x;
        cs.newLineAtOffset(x, 0);
    }

    public void resetPageNumber() {
        pageNumber = 1;
    }

    @Override
    public void close() throws IOException {
        endPage();
        document.save(out);
        document.close();
    }

    private void startPage() throws IOException {
        page = new PDPage(PDRectangle.A4);
        cs = new PDPageContentStream(document, page);

        cs.beginText();
        setFont(font, fontSize);

        offsetY = page.getMediaBox().getHeight() - MARGIN_T;
        tab = 0;
        pageNumber++;

        cs.newLineAtOffset(MARGIN_L, offsetY);
    }

    private void endPage() throws IOException {
        addPageNumber();
        cs.endText();
        cs.close();
        document.addPage(page);
    }

    private void setFont(PDFont font, float fontSize) throws IOException {
        this.font = font;
        this.fontSize = fontSize;
        this.leading = LINE_SPACING_COEF * fontSize;
        cs.setFont(font, fontSize);
        cs.setLeading(leading);
    }

    private void addPageNumber() throws IOException {
        // Place the page number not in the middle of the bottom margin but a bit higher
        cs.newLineAtOffset(-1 * tab, MARGIN_B * 3f / 5f - offsetY);
        tab = 0;
        addText(pageNumber, -1, Alignment.CENTER);
    }

    public enum Alignment {
        LEFT, CENTER, RIGHT;
    }

}
