package ee.ivxv.common.model;

public class Voter {

    static final String ACTIVITY_ADDITION = "lisamine";

    private final String code;
    private final String name;
    private final String parish;
    private final LName district;
    private final boolean isAddition;

    Voter(String code, String name, String activity, String parish, String district) {
        this(code, name, activity, parish, new LName(district));
    }

    public Voter(String code, String name, String activity, String parish, LName district) {
        this.code = code;
        this.name = name;
        this.parish = parish;
        this.district = district;
        isAddition = ACTIVITY_ADDITION.equals(activity);
    }

    public String getCode() {
        return code;
    }

    public String getName() {
        return name;
    }

    public LName getDistrict() {
        return district;
    }

    public String getParish() {
        return parish;
    }

    public boolean isAddition() {
        return isAddition;
    }

}
