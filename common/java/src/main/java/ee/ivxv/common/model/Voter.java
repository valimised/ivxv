package ee.ivxv.common.model;

public class Voter {

    static final String ACTIVITY_ADDITION = "lisamine";

    private final String code;
    private final String name;
    private final LName district;
    private final LName station;
    private final Long rowNumber;
    private final boolean isAddition;

    Voter(String code, String name, String activity, String districtId, String stationId,
            Long rowNumber) {
        this(code, name, activity, new LName(districtId), new LName(stationId), rowNumber);
    }

    public Voter(String code, String name, String activity, LName district, LName station,
            Long rowNumber) {
        this.code = code;
        this.name = name;
        this.district = district;
        this.station = station;
        this.rowNumber = rowNumber;
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

    public LName getStation() {
        return station;
    }

    public Long getRowNumber() {
        return rowNumber;
    }

    public boolean isAddition() {
        return isAddition;
    }

}
