package ee.ivxv.common.model;

/**
 * LName (location name) represents a voting district or voting station name. It consist of 2 parts:
 * {@code regionCode} and {@code number} and has the id of the form {@code <regionCode>.<number>}.
 */
public class LName {

    private final String id;
    private final String regionCode;
    private final String number;

    public LName(String id) {
        this.id = id;
        int i = id.indexOf('.');
        this.regionCode = id.substring(0, i < 0 ? id.length() : i);
        this.number = i < 0 ? null : id.substring(i + 1);
    }

    public LName(String regionCode, String number) {
        StringBuilder sb = new StringBuilder(regionCode);
        if (number != null) {
            sb.append('.').append(number);
        }
        this.id = sb.toString();
        this.regionCode = regionCode;
        this.number = number;
    }

    public String getId() {
        return id;
    }

    public String getRegionCode() {
        return regionCode;
    }

    public String getNumber() {
        return number;
    }

}
