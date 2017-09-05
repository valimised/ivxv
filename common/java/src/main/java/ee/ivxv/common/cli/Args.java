package ee.ivxv.common.cli;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

/**
 * List of command line arguments and some convenience methods.
 */
public class Args {

    public final List<Arg<?>> args = new ArrayList<>();

    public Args(Arg<?>... args) {
        this(Arrays.asList(args));
    }

    public Args(List<Arg<?>> args) {
        super();
        this.args.addAll(args);
    }

    /**
     * @return Whether any argument is required.
     */
    public boolean isRequired() {
        return args.stream().anyMatch(Arg::isRequired);
    }

    /**
     * @return Whether any argument is set.
     */
    public boolean isSet() {
        return args.stream().anyMatch(Arg::isSet);
    }

    /**
     * @return Whether all arguments are valid.
     */
    public boolean isValid() {
        return validate().isValid();
    }

    /**
     * Validates the arguments and returns the validation result.
     * 
     * @return Returns the validation result.
     */
    public final ValidationResult validate() {
        return validate(true, new ValidationResult());
    }

    /**
     * Validates the argument list.
     * 
     * @param shouldExist Indicates whether the node containing the argument list should exist,
     *        hence should check the arguments.
     * @param r The result where to accumulate the results.
     * @return
     */
    final ValidationResult validate(boolean shouldExist, ValidationResult r) {
        boolean check = isSet() || shouldExist;
        args.stream().filter(a -> a.isSet() || check).forEach(a -> a.validate(r));
        return r;
    }

    public Arg<?> find(String name) {
        return args.stream().filter(a -> a.name.getName().equals(name)).findFirst().orElse(null);
    }

}
