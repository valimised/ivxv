package ee.ivxv.common.service.bbox.impl;

import ee.ivxv.common.service.bbox.BboxHelper.Stage;

abstract class AbstractStage implements Stage {

    private final int numberOfValidBallots;
    private final int numberOfInvalidBallots;

    AbstractStage(int numberOfValidBallots) {
        this(numberOfValidBallots, numberOfValidBallots);
    }

    AbstractStage(int numberOfValidBallots, int oldValidBallots) {
        this.numberOfValidBallots = numberOfValidBallots;
        this.numberOfInvalidBallots = oldValidBallots - numberOfValidBallots;
    }

    @Override
    public int getNumberOfValidBallots() {
        return numberOfValidBallots;
    }

    @Override
    public int getNumberOfInvalidBallots() {
        return numberOfInvalidBallots;
    }

}
