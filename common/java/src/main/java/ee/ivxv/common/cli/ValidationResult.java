package ee.ivxv.common.cli;

import ee.ivxv.common.service.i18n.Translatable;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

public class ValidationResult {

    private final List<Error> errors = new ArrayList<>();

    public List<Error> getErrors() {
        return Collections.unmodifiableList(errors);
    }

    public void addError(Arg<?> arg, Translatable msg) {
        errors.add(new Error(arg, msg));
    }

    public boolean isValid() {
        return errors.isEmpty();
    }

    public static class Error {
        public final Arg<?> arg;
        public final Translatable error;

        public Error(Arg<?> arg, Translatable error) {
            this.arg = arg;
            this.error = error;
        }
    }

}
