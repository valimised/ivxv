package ee.ivxv.common.service.i18n;

public class MessageException extends RuntimeException implements Translatable {

    private static final long serialVersionUID = 2925407779875606910L;

    private final Message msg;

    public MessageException(Enum<?> key, Object... args) {
        this(new Message(key, args));
    }

    public MessageException(Throwable cause, Enum<?> key, Object... args) {
        this(new Message(key, args), cause);
    }

    public MessageException(Message msg) {
        super(msg.key.name());
        this.msg = msg;
    }

    public MessageException(Message msg, Throwable cause) {
        super(msg.key.name(), cause);
        this.msg = msg;
    }

    @Override
    public Enum<?> getKey() {
        return msg.key;
    }

    @Override
    public Object[] getArgs() {
        return msg.args;
    }

}
