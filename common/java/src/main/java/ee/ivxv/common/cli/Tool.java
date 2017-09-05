package ee.ivxv.common.cli;

import ee.ivxv.common.util.NameHolder;
import java.util.function.Function;
import java.util.function.Supplier;

/**
 * Implements single functionality/command of a command line application.
 * 
 * @param <T> The type of the application context of this tool.
 * @param <U> The type of the arguments of this tool.
 */
public class Tool<T extends AppContext<?>, U extends Args> {

    public final NameHolder name;
    public final Supplier<U> argsSupplier;
    public final Function<T, Runner<U>> runnerSupplier;

    public Tool(NameHolder name, Supplier<U> argsSupplier, Function<T, Runner<U>> runnerSupplier) {
        this.name = name;
        this.argsSupplier = argsSupplier;
        this.runnerSupplier = runnerSupplier;
    }

    /** @return New instance of arguments for this particular type of tool. */
    protected U createArgs() {
        return argsSupplier.get();
    }

    /**
     * @param ctx The correct application context.
     * @return Returns the functional part of the application.
     */
    protected Runner<U> prepare(T ctx) {
        return runnerSupplier.apply(ctx);
    }

    /**
     * Separate interface to ensure that tools are run only after proper preparation.
     * 
     * @param <U> The type of the arguments of this tool.
     */
    public interface Runner<U extends Args> {
        /**
         * Runs the prepared tool with the specified arguments.
         * 
         * @param args Evaluated and valid arguments of the tool.
         * @return Whether the execution was successful.
         * @throws Exception any exception thrown here is handled by central error handling logic.
         */
        boolean run(U args) throws Exception;
    }

}
