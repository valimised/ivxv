package ee.ivxv.common.cli;

import ee.ivxv.common.M;
import ee.ivxv.common.conf.Conf;
import ee.ivxv.common.conf.ConfLoader;
import ee.ivxv.common.conf.ConfVerifier;
import ee.ivxv.common.conf.LocaleConfLoader;
import ee.ivxv.common.service.container.InvalidContainerException;
import ee.ivxv.common.service.i18n.MessageException;
import ee.ivxv.common.util.ContainerHelper;
import ee.ivxv.common.util.I18nConsole;
import ee.ivxv.common.util.log.PerformanceLog;
import java.io.InputStream;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Locale;
import java.util.stream.Collectors;
import java.util.stream.Stream;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.slf4j.MDC;

/**
 * General logic for running IVXV command line application - taking care of common aspects, like
 * checking arguments, showing help, handling errors, generating session id for log, etc.
 * 
 * <p>
 * The expected execution pattern is one of:
 * 
 * <pre>
 * $ &lt;app&gt; [-h|--help]
 * $ &lt;app&gt; &lt;tool&gt; [-h|--help]
 * $ &lt;app&gt; &lt;tool&gt; [tool-arguments]
 * </pre>
 * 
 * @param <T> The type of the application context of the instance.
 */
public class AppRunner<T extends AppContext<?>> {

    static final Logger log = LoggerFactory.getLogger(AppRunner.class);

    private static final String SESSION_ID = "SID";

    private final InitialContext ictx;
    final App<T> app;
    private final CommonArgs cargs;
    final I18nConsole console;
    final AppHelper appHelper;

    // Error handler - different handling for application and tool level errors
    private ErrorHandler errorHandler;
    // The parsed command line, once arguments are provided
    private CommandLine cl;
    private final List<Runnable> finalizers = new ArrayList<>();

    /**
     * Calls the 2-parameter constructor with {@code ContextFactory.createInitialContext()}.
     * 
     * @param app The application to run.
     */
    public AppRunner(App<T> app) {
        this(ContextFactory.get().createInitialContext(), app);
    }

    /**
     * @param ctx The <i>initial</i> context to be used until the configuration is loaded and new
     *        application context created from it.
     * @param app The application to run.
     */
    public AppRunner(InitialContext ictx, App<T> app) {
        this.ictx = ictx;
        this.app = app;
        cargs = new CommonArgs();
        console = new I18nConsole(ictx.console, ictx.i18n);
        appHelper = new AppHelper(app, cargs, console);

        errorHandler = new AppErrorHandler();
        finalizers.add(() -> ictx.console.shutdown());
    }

    public boolean run(String[] args) {
        boolean result = false;
        long t = System.currentTimeMillis();

        try {
            MDC.put(SESSION_ID, ictx.sessionId);
            PerformanceLog.log.info("STARTING application '{}'", app.name.getName());
            LocaleConfLoader.load(ictx.locale);

            if (ContextFactory.get().isTestMode()) {
                console.println(Msg.w_test_mode);
            }

            cl = CommandLine.parse(args);

            log.info("Executing: {}", Stream.of(args).collect(Collectors.joining(" ")));

            result = runInternal();
        } catch (InvalidContainerException e) {
            errorHandler.handleInvalidContainerException(e);
        } catch (ParseException e) {
            errorHandler.handleParseException(e);
        } catch (MessageException e) {
            errorHandler.handleMessageException(e);
        } catch (Throwable e) {
            // Safety-net exception handling
            errorHandler.handleThrowable(e);
        } finally {
            Msg msg = result ? Msg.app_result_success : Msg.app_result_failure;
            console.println(msg, app.name);
            try {
                finalizers.forEach(f -> f.run());
            } catch (Throwable e) {
                errorHandler.handleThrowable(e);
            }
            PerformanceLog.log.info("FINISHED application '{}', TIME: {} ms, SUCCESS: {}",
                    app.name.getName(), (System.currentTimeMillis() - t), result);
            MDC.remove(SESSION_ID);
        }
        return result;
    }

    private boolean runInternal() throws Exception {
        cl.set(cargs);

        log.debug("Application '{}' arguments:", app.name.getName());
        logArgValues(cargs);

        // Set language, if requested
        if (cargs.lang.isSet()) {
            ictx.locale.setLocale(new Locale(cargs.lang.value()));
        }

        // Show application help, if requested (tool not selected)
        if (cl.getCommands().isEmpty() && cargs.help.value()) {
            appHelper.showHelp();
            return true;
        }

        // Proceed with selected tool
        Tool<T, ?> tool = selectTool();

        // Show tool help, if requested
        if (cargs.help.value()) {
            appHelper.showToolHelp(tool);
            return true;
        }

        PerformanceLog.log.info("Starting tool '{}', app-threads: {}, container-threads: {}",
                tool.name.getName(), cargs.threads.value(), cargs.ct.value());

        return runTool(tool);
    }

    private Tool<T, ?> selectTool() {
        List<String> commands = cl.getCommands();

        if (commands.isEmpty()) {
            throw new ParseException(Msg.e_tool_missing);
        }

        if (commands.size() > 1) {
            throw new ParseException(Msg.e_multiple_tools, commands);
        }

        String first = commands.get(0);
        if (!app.tools.containsKey(first)) {
            throw new ParseException(Msg.e_unknown_tool, first);
        }

        return app.tools.get(first);
    }

    private <U extends Tool<T, A>, A extends Args> boolean runTool(U tool) throws Exception {
        A args = tool.createArgs();

        errorHandler = new ToolErrorHandler<>(tool);

        cl.set(args);

        // Command line is processed - check unknown options
        checkUnknownOptions();

        if (!cargs.isValid()) {
            cargs.validate().getErrors().forEach(e -> console.println(e.error));
            throw new ParseException(Msg.e_common_args_invalid);
        }

        // Create application-specific application context and run the tool
        T ctx = createContext();
        console.println();

        readArgsFromYaml(ctx, tool.name.getName(), args);

        log.debug("Tool '{}' arguments:", tool.name.getName());
        logArgValues(args);

        if (!args.isValid()) {
            args.validate().getErrors().forEach(e -> console.println(e.error));
            throw new ParseException(Msg.e_tool_args_invalid);
        }

        return tool.prepare(ctx).run(args);
    }

    private void checkUnknownOptions() {
        boolean hasUnused = false;

        for (Option o : cl.getUnusedOptions()) {
            log.debug("Unknown option '{}'", o.getName());
            console.println(Msg.e_unknown_arg, o.getName());
            hasUnused = true;
        }

        if (hasUnused) {
            throw new ParseException(Msg.e_unknown_args_present);
        }
    }

    private T createContext() {
        // Load conf (conf as a mandatory field is set)
        Conf conf = ConfLoader.load(cargs.conf.value(), console);

        ConfVerifier.verify(conf);

        // Create application-specific application context with the loaded conf
        T ctx = app.createContext(ictx, conf, cargs);

        ConfVerifier.verifySignature(ctx, cargs.conf.value());

        // Must shut down the container service, otherwise the application may hang
        finalizers.add(() -> ctx.container.shutdown());

        return ctx;
    }

    private void readArgsFromYaml(AppContext<?> ctx, String toolName, Args args) throws Exception {
        if (!cargs.params.isSet()) {
            return;
        }
        Path path = cargs.params.value();
        console.println(M.m_loading_params, path);
        ctx.container.requireContainer(path);
        ContainerHelper ch = new ContainerHelper(console, ctx.container.read(path.toString()));
        try (InputStream in = ch.getSingleFileAndReport(M.m_params_arg_for_cont).getStream()) {
            YamlData yaml = YamlData.parse(path, in);
            yaml.set(toolName, args);
        }
    }

    private void logArgValues(Args args) {
        AppHelper.walk(args, 1, (arg, level) -> {
            String indent = String.join("", Collections.nCopies(level, "  "));
            if ((arg instanceof Arg.Tree) || (arg instanceof Arg.TreeList<?>)) {
                log.debug("{}{}:", indent, arg.name.getName());
            } else {
                log.debug("{}{} = {}", indent, arg.name.getName(), arg.value);
            }
        });
    }

    /**
     * Interface for local error handler - either application level or tool level.
     */
    interface ErrorHandler {
        void handleInvalidContainerException(InvalidContainerException e);

        void handleParseException(ParseException e);

        void handleMessageException(MessageException e);

        void handleThrowable(Throwable e);
    }

    class AppErrorHandler implements ErrorHandler {

        private final String appName;

        public AppErrorHandler() {
            appName = app.name.getName();
        }

        @Override
        public void handleInvalidContainerException(InvalidContainerException e) {
            log.warn("A container validation error occurred running app '{}'", appName, e);
            console.println(M.e_invalid_container, e.path);
        }

        @Override
        public void handleParseException(ParseException e) {
            log.debug("Parsing error occurred running app '{}'", appName, e);
            console.println(e.getKey(), e.getArgs());
            console.println();
            appHelper.showHelp();
        }

        @Override
        public void handleMessageException(MessageException e) {
            log.warn("An error occurred running app '{}'", appName, e);
            console.println(e.getKey(), e.getArgs());
        }

        @Override
        public void handleThrowable(Throwable e) {
            log.warn("An error occurred running app '{}'", appName, e);
            console.println(Msg.e_app_error, appName, e);
        }
    }

    class ToolErrorHandler<U extends Tool<T, ?>> implements ErrorHandler {

        private final U tool;
        private final String appName;
        private final String toolName;

        public ToolErrorHandler(U tool) {
            this.tool = tool;
            appName = app.name.getName();
            toolName = tool.name.getName();
        }

        @Override
        public void handleInvalidContainerException(InvalidContainerException e) {
            log.warn("A container validation error occurred running app '{}'", appName, e);
            console.println(M.e_invalid_container, e.path);
        }

        @Override
        public void handleParseException(ParseException e) {
            log.debug("Parsing error occurred running app '{}' tool '{}'", appName, toolName, e);
            console.println(e.getKey(), e.getArgs());
            console.println();
            appHelper.showToolHelp(tool);
        }

        @Override
        public void handleMessageException(MessageException e) {
            log.warn("An error occurred running app '{}' tool '{}'", appName, toolName, e);
            console.println(e.getKey(), e.getArgs());
        }

        @Override
        public void handleThrowable(Throwable e) {
            log.warn("An error occurred running app '{}' tool '{}'", appName, toolName, e);
            console.println(Msg.e_tool_error, toolName, e);
        }
    }

}
