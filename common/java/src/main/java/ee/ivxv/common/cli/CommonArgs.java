package ee.ivxv.common.cli;

import java.nio.file.Path;

/**
 * CommonArgs is the list of common arguments that are processed on every run for every application.
 * Arguments 'conf' and 'force' are application-level arguments, but 'help' works both on
 * application or tool level - depending on whether tool is specified.
 */
public class CommonArgs extends Args {

    private final int p = Runtime.getRuntime().availableProcessors();

    public final Arg<Boolean> help = Arg.aFlag(Msg.arg_help).setOptional();
    public final Arg<Path> conf = Arg.aPath(Msg.arg_conf, true, null);
    public final Arg<Path> params = Arg.aPath(Msg.arg_params, true, false).setOptional();
    public final Arg<Boolean> force = Arg.aFlag(Msg.arg_force).setOptional();
    public final Arg<Boolean> quiet = Arg.aFlag(Msg.arg_quiet).setOptional();
    public final Arg<String> lang = Arg.aString(Msg.arg_lang).setOptional();
    public final Arg<Integer> ct = Arg.anInt(Msg.arg_container_threads).setDefault(0).setOptional();
    public final Arg<Integer> threads = Arg.anInt(Msg.arg_threads).setDefault(p + 1).setOptional();

    public CommonArgs() {
        args.add(help);
        args.add(conf);
        args.add(params);
        args.add(force);
        args.add(quiet);
        args.add(lang);
        args.add(ct);
        args.add(threads);
    }
}
