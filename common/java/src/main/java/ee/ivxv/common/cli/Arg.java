package ee.ivxv.common.cli;

import ee.ivxv.common.crypto.CryptoUtil;
import ee.ivxv.common.crypto.CryptoUtil.PublicKeyHolder;
import ee.ivxv.common.service.i18n.Message;
import ee.ivxv.common.util.NameHolder;
import ee.ivxv.common.util.Util;
import java.math.BigInteger;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.time.Instant;
import java.time.LocalDate;
import java.time.LocalTime;
import java.time.OffsetDateTime;
import java.time.ZoneOffset;
import java.time.format.DateTimeParseException;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.function.Function;
import java.util.function.Supplier;
import java.util.stream.Collectors;
import org.bouncycastle.asn1.ASN1ObjectIdentifier;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Instances of <code>Arg</code> are the building blocks of the application (and tool) arguments'
 * data model. The argument values can be acquired from either command line ({@link Option}) or a
 * configuration file.
 * 
 * @param <T> The type of the argument value.
 */
public abstract class Arg<T> {

    static final Logger log = LoggerFactory.getLogger(Arg.class);

    /*
     * Re-used parser instances. Since parsers are stateless, there is no need for creating new one
     * for each argument.
     */
    private static final Parser<String> STRING_PARSER = new StringParser();
    private static final Parser<Boolean> BOOLEAN_PARSER = new BooleanParser();
    private static final Parser<byte[]> BYTE_ARRAY_PARSER = new ByteArrayParser();
    private static final Parser<Integer> INT_PARSER = new IntParser();
    private static final Parser<BigInteger> BIGINT_PARSER = new BigIntParser();
    private static final Parser<Instant> INSTANT_PARSER = new InstantParser();
    private static final Parser<LocalDate> LOCAL_DATE_PARSER = new LocalDateParser();
    private static final Parser<Path> PATH_PARSER = new PathParser();
    static final Parser<Path> EXISTING_FILE_PARSER = new PathParser(true, false);

    static final Resolver IDENTITY_RESOLVER = new Resolver();

    public final NameHolder name;
    T value;
    private boolean isRequired = true;
    private T def;

    public Arg(NameHolder name) {
        this.name = name;
    }

    // ******************************** STATIC CONVENIENCE METHODS ****************************** //

    public static Arg<String> aString(NameHolder n) {
        return new SingleValueArg<>(n, STRING_PARSER);
    }

    public static Arg<Boolean> aFlag(NameHolder n) {
        return new SingleValueArg<Boolean>(n, BOOLEAN_PARSER) {

            @Override
            public void parse(List<String> v, Resolver r) throws ParseException {
                // Allow empty argument list, which means 'true'
                if (v.isEmpty()) {
                    value = Boolean.TRUE;
                    return;
                }
                super.parse(v, r);
            }

        }.setDefault(Boolean.FALSE).setOptional();

    }

    public static Arg<byte[]> aByteArray(NameHolder n) {
        return new SingleValueArg<>(n, BYTE_ARRAY_PARSER);
    }

    public static Arg<Integer> anInt(NameHolder n) {
        return new SingleValueArg<>(n, INT_PARSER);
    }

    public static Arg<BigInteger> aBigInt(NameHolder n) {
        return new SingleValueArg<>(n, BIGINT_PARSER);
    }

    public static Arg<Instant> anInstant(NameHolder n) {
        return new SingleValueArg<>(n, INSTANT_PARSER);
    }

    public static Arg<LocalDate> aLocalDate(NameHolder n) {
        return new SingleValueArg<>(n, LOCAL_DATE_PARSER);
    }

    public static Arg<Path> aPath(NameHolder n) {
        return new SingleValueArg<>(n, PATH_PARSER);
    }

    public static Arg<Path> aPath(NameHolder n, Boolean exists, Boolean isDir) {
        return new SingleValueArg<>(n, new PathParser(exists, isDir));
    }

    public static Arg<PublicKeyHolder> aPublicKey(NameHolder n, ASN1ObjectIdentifier alg) {
        return new SingleValueArg<>(n, new PublicKeyParser(alg));
    }

    /**
     * Creates an argument instance for a value from the specified choices, which is found by
     * matching the result of <tt>toString()</tt> of a choice with the string being parsed.
     * 
     * @param n
     * @param values
     * @return
     */
    @SafeVarargs
    public static <U> Arg<U> aChoice(NameHolder n, U... values) {
        return new SingleValueArg<>(n, new ChoiceParser<>(values));
    }

    @SafeVarargs
    public static <U> Arg<U> aChoice(NameHolder n, Function<String, U> mapper, U... values) {
        return new SingleValueArg<>(n, new ChoiceParser<>(mapper, values));
    }

    public static Arg<List<String>> listOfStrings(NameHolder n) {
        return new MultiValueArg<>(n, STRING_PARSER);
    }

    public static Arg<List<Boolean>> listOfFlags(NameHolder n) {
        return new MultiValueArg<>(n, BOOLEAN_PARSER);
    }

    public static Arg<List<byte[]>> listOfByteArrays(NameHolder n) {
        return new MultiValueArg<>(n, BYTE_ARRAY_PARSER);
    }

    public static Arg<List<Integer>> listOfInts(NameHolder n) {
        return new MultiValueArg<>(n, INT_PARSER);
    }

    public static Arg<List<BigInteger>> listOfBigInts(NameHolder n) {
        return new MultiValueArg<>(n, BIGINT_PARSER);
    }

    public static Arg<List<Instant>> listOfInstants(NameHolder n) {
        return new MultiValueArg<>(n, INSTANT_PARSER);
    }

    public static Arg<List<Path>> listOfPaths(NameHolder n) {
        return new MultiValueArg<>(n, new PathParser());
    }

    public static Arg<List<Path>> listOfPaths(NameHolder n, Boolean exists, Boolean isDir) {
        return new MultiValueArg<>(n, new PathParser(exists, isDir));
    }

    public static Arg<List<PublicKeyHolder>> listOfPublicKeys(NameHolder n,
            ASN1ObjectIdentifier alg) {
        return new MultiValueArg<>(n, new PublicKeyParser(alg));
    }

    /**
     * Creates an argument instance for a list of values from the specified choices, which are found
     * by matching the result of <tt>toString()</tt> of a choice with the string being parsed.
     * 
     * @param n
     * @param values
     * @return
     */
    @SafeVarargs
    public static <U> Arg<List<U>> listOfChoices(NameHolder n, U... values) {
        return new MultiValueArg<>(n, new ChoiceParser<>(values));
    }

    @SafeVarargs
    public static <U> Arg<List<U>> listOfChoices(NameHolder n, Function<String, U> mapper,
            U... values) {
        return new MultiValueArg<>(n, new ChoiceParser<>(mapper, values));
    }

    // ******************************** INSTANCE METHODS ******************************** //

    /**
     * Sets the argument optional. By default arguments are required.
     * 
     * @return This
     */
    public Arg<T> setOptional() {
        isRequired = false;
        return this;
    }

    /**
     * Set the default value of the argument. Note that argument with a not-null default has always
     * a value.
     * 
     * @return This
     */
    public Arg<T> setDefault(T v) {
        def = v;
        return this;
    }

    /**
     * @return Whether the argument is required in the parent context - parent is set or required.
     */
    public boolean isRequired() {
        return isRequired;
    }

    /**
     * @param s Not-null list of strings
     * @throws ParseException
     */
    public final void parse(List<String> s) throws ParseException {
        parse(s, IDENTITY_RESOLVER);
    }

    // TODO replace exception throwing with returned errors object
    public abstract void parse(List<String> s, Resolver r) throws ParseException;

    /**
     * @return The value of the argument: either the parsed or the default value.
     */
    public T value() {
        return value != null ? value : def;
    }

    /**
     * @return Whether the argument has any value.
     */
    public boolean isSet() {
        return value() != null;
    }

    /**
     * @return Whether the argument value is valid. Only checks that required argument is set.
     */
    public final boolean isValid() {
        return validate().isValid();
    }

    /**
     * Validates the argument and possibly it's children and returns the validation result.
     * 
     * @return Returns the validation result.
     */
    public final ValidationResult validate() {
        return validate(new ValidationResult());
    }

    ValidationResult validate(ValidationResult result) {
        if (isRequired() && !isSet()) {
            result.addError(this, new Message(Msg.e_arg_required, name.getName()));
        }
        return result;
    }

    // ******************************** IMPLEMENTATIONS ******************************** //

    /**
     * Argument with single value.
     * 
     * @param <T> The value type.
     */
    public static class SingleValueArg<T> extends Arg<T> {

        private final Parser<T> parser;

        public SingleValueArg(NameHolder name, Parser<T> parser) {
            super(name);
            this.parser = parser;
        }

        @Override
        public void parse(List<String> v, Resolver r) throws ParseException {
            if (v.size() != 1) {
                throw new ParseException(Msg.e_requires_single_value, name.getName(), v);
            }
            try {
                value = parser.parse(v.get(0), r);
            } catch (Exception e) {
                throw new ParseException(Msg.e_arg_parse_error, name.getName(), v, e);
            }
        }
    }

    /**
     * Argument with multiple values, i.e list of values.
     * 
     * @param <T> The value type.
     */
    public static class MultiValueArg<T> extends Arg<List<T>> {

        private final Parser<T> parser;

        public MultiValueArg(NameHolder name, Parser<T> parser) {
            super(name);
            this.parser = parser;
        }

        @Override
        public void parse(List<String> v, Resolver r) throws ParseException {
            try {
                value = v.stream().map(s -> parser.parse(s, r)).collect(Collectors.toList());
            } catch (Exception e) {
                throw new ParseException(Msg.e_arg_parse_error, name.getName(), v, e);
            }
        }

        @Override
        public boolean isSet() {
            return value != null && !value.isEmpty();
        }
    }

    /**
     * Represents hierarchical argument tree. Tree value is set if any of the sub-arguments is set.
     * If tree is required, it must be set even if all sub-arguments are optional.
     */
    public static class Tree extends Arg<Args> {

        private boolean isExclusive;

        public Tree(NameHolder name, Arg<?>... args) {
            super(name);
            this.value = new Args(args);
        }

        /**
         * If <tt>true</tt> at most one branch can be set. It does not make sense to have tree
         * exclusive and sub-arguments required.
         */
        public boolean isExclusive() {
            return isExclusive;
        }

        public Tree setExclusive() {
            isExclusive = true;
            return this;
        }

        @Override
        public void parse(List<String> v, Resolver r) throws ParseException {
            if (!v.isEmpty()) {
                throw new ParseException(Msg.e_value_not_allowed, name.getName(), v);
            }
        }

        @Override
        public boolean isSet() {
            return value.isSet();
        }

        @Override
        public final ValidationResult validate(ValidationResult result) {
            long setCount = value.args.stream().filter(a -> a.isSet()).count();
            if (isExclusive && setCount > 1) {
                result.addError(this,
                        new Message(Msg.e_requires_single_value, name.getName(), setCount));
            }
            super.validate(result);
            value.validate(isRequired(), result);
            return result;
        }
    }

    /**
     * TreeList is a arbitrary-size list with structural content, i.e. list of <tt>Args</tt>.
     * 
     * @param <T> The actual type of <tt>Args</tt> contained in this list.
     */
    public static class TreeList<T extends Args> extends Arg<List<T>> {

        private final Supplier<T> factory;

        public TreeList(NameHolder name, Supplier<T> factory) {
            super(name);
            this.factory = factory;
            value = new ArrayList<>();
        }

        @Override
        public void parse(List<String> v, Resolver r) throws ParseException {
            if (!v.isEmpty()) {
                throw new ParseException(Msg.e_value_not_allowed, name.getName(), v);
            }
        }

        @Override
        public boolean isSet() {
            return value.stream().anyMatch(Args::isSet);
        }

        @Override
        public final ValidationResult validate(ValidationResult result) {
            super.validate(result);
            // Use 'true' as 'shouldExist' in Args.validate(), because node obviously exists
            value.forEach(a -> a.validate(true, result));
            return result;
        }

        /**
         * @return Adds new instance of <tt>T</tt> to the value (list of <tt>T</tt>) and returns it.
         */
        public T addNew() {
            T e = factory.get();
            value.add(e);
            return e;
        }
    }

    // ******************************** PARSERS ******************************** //

    /**
     * Parser is an interface for parsing a string into another type.
     * 
     * @param <T> The target type to parse into.
     */
    @FunctionalInterface
    interface Parser<T> {
        default T parse(String s) throws ParseException {
            return parse(s, IDENTITY_RESOLVER);
        }

        T parse(String s, Resolver r) throws ParseException;
    }

    static class StringParser implements Parser<String> {
        @Override
        public String parse(String s, Resolver r) {
            return s;
        }
    }

    static class BooleanParser implements Parser<Boolean> {
        static final List<String> TRUE_STRINGS = Arrays.asList("true", "TRUE");
        static final List<String> FALSE_STRINGS = Arrays.asList("false", "FALSE");

        @Override
        public Boolean parse(String s, Resolver r) {
            if (TRUE_STRINGS.contains(s)) {
                return Boolean.TRUE;
            } else if (FALSE_STRINGS.contains(s)) {
                return Boolean.FALSE;
            } else {
                throw new ParseException(Msg.e_invalid_boolean, s);
            }
        }
    }

    static class ByteArrayParser implements Parser<byte[]> {
        @Override
        public byte[] parse(String s, Resolver r) {
            return Util.toBytes(s);
        }
    }

    static class IntParser implements Parser<Integer> {
        @Override
        public Integer parse(String s, Resolver r) {
            try {
                return Integer.parseInt(s);
            } catch (NumberFormatException e) {
                throw new ParseException(Msg.e_invalid_int, s);
            }
        }
    }

    static class BigIntParser implements Parser<BigInteger> {
        @Override
        public BigInteger parse(String s, Resolver r) {
            try {
                return new BigInteger(s, 10);
            } catch (NumberFormatException e) {
                throw new ParseException(Msg.e_invalid_number, s);
            }
        }
    }

    static class InstantParser implements Parser<Instant> {
        @Override
        public Instant parse(String s, Resolver r) {
            try {
                // Use OffsetDateTime parser to allow "+/-zz:zz" format time zone
                return OffsetDateTime.parse(s).toInstant();
            } catch (DateTimeParseException e) {
                throw new ParseException(Msg.e_invalid_instant, s);
            }
        }
    }

    static class LocalDateParser implements Parser<LocalDate> {
        @Override
        public LocalDate parse(String s, Resolver r) {
            try {
                return LocalDate.parse(s);
            } catch (DateTimeParseException e) {
                // Dates parsed via YAML are reformatted with time and zone offset, so also
                // check for that format, but ensure that time and offset are both zero.
                try {
                    OffsetDateTime full = OffsetDateTime.parse(s);
                    if (!full.toLocalTime().equals(LocalTime.MIDNIGHT)
                            || !full.getOffset().equals(ZoneOffset.UTC)) {
                        throw new ParseException(Msg.e_invalid_local_date, s);
                    }
                    return full.toLocalDate();
                } catch (DateTimeParseException e2) {
                    throw new ParseException(Msg.e_invalid_local_date, s);
                }
            }
        }
    }

    static class PathParser implements Parser<Path> {
        private final Boolean exists;
        private final Boolean isDir;

        PathParser() {
            this(null, null);
        }

        /**
         * @param exists Must the path exist.
         * @param isDir If <tt>exists</tt>, must the path be a directory.
         */
        PathParser(Boolean exists, Boolean isDir) {
            this.exists = exists;
            this.isDir = exists != null && exists ? isDir : null;
        }

        @Override
        public Path parse(String s, Resolver r) {
            Path p = r.resolve(Paths.get(s));

            if (exists != null && Files.exists(p) != exists) {
                Msg msg = exists ? Msg.e_invalid_path_not_exists : Msg.e_invalid_path_exists;
                throw new ParseException(msg, s);
            }
            if (isDir != null && Files.isDirectory(p) != isDir) {
                Msg msg = isDir ? Msg.e_invalid_path_not_dir : Msg.e_invalid_path_not_file;
                throw new ParseException(msg, s);
            }

            return p;
        }
    }

    static class PublicKeyParser implements Parser<PublicKeyHolder> {
        private final ASN1ObjectIdentifier alg;

        PublicKeyParser(ASN1ObjectIdentifier alg) {
            this.alg = alg;
        }

        @Override
        public PublicKeyHolder parse(String s, Resolver r) throws ParseException {
            Path path = EXISTING_FILE_PARSER.parse(s, r);

            try {
                return CryptoUtil.loadPublicKey(path, alg);
            } catch (Exception e) {
                throw new ParseException(e, Msg.e_invalid_public_key, path.toString(), alg, e);
            }
        }
    }

    static class ChoiceParser<T> implements Parser<T> {
        private final List<T> values;
        private final Function<String, T> mapper;

        ChoiceParser(T[] values) {
            this(s -> defaultMapper(s, values), values);
        }

        ChoiceParser(Function<String, T> mapper, T[] values) {
            this.values = Arrays.asList(values);
            this.mapper = mapper;
        }

        private static <T> T defaultMapper(String s, T[] values) {
            for (T v : values) {
                if (v.toString().equals(s)) {
                    return v;
                }
            }
            return null;
        }

        @Override
        public T parse(String s, Resolver r) {
            try {
                T value = mapper.apply(s);
                if (value != null) {
                    return value;
                }
            } catch (Exception e) {
                log.debug("parse({}), mapper error occurred: {}", s, e.getMessage(), e);
            }
            throw new ParseException(Msg.e_invalid_choice, s, values);
        }
    }

    // ******************************** OTHER ******************************** //

    /**
     * Resolver represents the context in which an argument is being parsed and helps to resolve or
     * translate the parsed value according to the context.
     */
    static class Resolver {

        private final Function<Path, Path> pathResolver;

        Resolver() {
            this(p -> p);
        }

        Resolver(Function<Path, Path> pathResolver) {
            this.pathResolver = pathResolver;
        }

        Path resolve(Path p) {
            return pathResolver.apply(p);
        }

        // Add support for more types as necessary

    }

}
