package ee.ivxv.common.service.bbox.impl;

import ee.ivxv.common.service.bbox.Result;

public class ResultException extends RuntimeException {

    private static final long serialVersionUID = -1689828260220357741L;

    final Result result;
    final Object[] args;

    public ResultException(Result result, Object... args) {
        super(result.name());
        this.result = result;
        this.args = args;
    }

}
