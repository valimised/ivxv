package ee.ivxv.common.service.container;

public class Subject {

    private final String serialNumber;
    private final String name;

    public Subject(String serialNumber, String name) {
        this.serialNumber = serialNumber;
        this.name = name;
    }

    public String getSerialNumber() {
        return serialNumber;
    }

    public String getName() {
        return name;
    }

}
