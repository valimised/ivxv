package ee.ivxv.common.service.i18n;

/**
 * Container class to transport translatable message with parameters from lower levels to
 * application level, where it can be translated.
 */
public class Message implements Translatable {

    public final Enum<?> key;
    public final Object[] args;

    public Message(Enum<?> key, Object... args) {
        this.key = key;
        this.args = args;
    }

    @Override
    public Enum<?> getKey() {
        return key;
    }

    @Override
    public Object[] getArgs() {
        return args;
    }

}
