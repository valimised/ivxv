package ee.ivxv.common.service.console;

public class InfiniteProgressImpl extends ProgressImpl {
    InfiniteProgressImpl(long total, Runnable finishedNotifier) {
        super(total, finishedNotifier);
    }

    @Override
    public void increase(int amount) {
        if (super.getValue() >= super.getTotal()) {
            super.reset();
        }
        super.increase(amount);
    }
}
