package ee.ivxv.key.tool;

import ee.ivxv.common.cli.Arg;
import ee.ivxv.common.cli.Args;
import ee.ivxv.common.cli.Tool;
import ee.ivxv.common.crypto.rnd.Rnd;
import ee.ivxv.common.math.ECGroup;
import ee.ivxv.common.math.Group;
import ee.ivxv.common.math.GroupElement;
import ee.ivxv.common.math.MathException;
import ee.ivxv.common.math.ModPGroup;
import ee.ivxv.common.math.ModPGroupElement;
import ee.ivxv.common.service.console.Progress;
import ee.ivxv.common.util.I18nConsole;
import ee.ivxv.common.util.Util;
import ee.ivxv.key.KeyContext;
import ee.ivxv.key.Msg;
import ee.ivxv.key.RandomSourceArg;
import ee.ivxv.key.tool.GroupGenTool.GroupGenArgs;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Iterator;
import java.util.List;
import java.util.concurrent.Callable;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import java.util.concurrent.FutureTask;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.TimeoutException;

/**
 * GroupGenTool is a tool for generating ElGamal group parameters.
 */
public class GroupGenTool implements Tool.Runner<GroupGenArgs> {

    private static final String MOD_GROUP = "mod";
    private static final String EC_GROUP = "ec";

    private final I18nConsole console;
    private final KeyContext ctx;

    public GroupGenTool(KeyContext ctx) {
        this.console = new I18nConsole(ctx.i.console, ctx.i.i18n);
        this.ctx = ctx;
    }

    @Override
    public boolean run(GroupGenArgs args) throws Exception {
        Group group = null;
        GroupElement generator = null;
        Rnd rnd = RandomSourceArg.combineFromArgument(args.random);
        Progress p = console.startInfiniteProgress(1000);
        try {
            group = groupGen(args.paramType.value(), args.len.value(), rnd, p);
            generator = generatorGen(group, rnd);
        } finally {
            rnd.close();
            p.finish();
        }
        writeParameters(group, generator, args.initTemplate.value());
        return true;
    }

    private Group groupGen(String groupType, int len, Rnd rnd, Progress p) throws Exception {
        switch (groupType) {
            case MOD_GROUP:
                return modpGroupGen(len, rnd, p);
            case EC_GROUP:
                return new ECGroup(len);
            default:
                // this path should not be executed
                throw new UnsupportedOperationException("Unsupported group type");
        }
    }

    private ModPGroup modpGroupGen(int len, Rnd rnd, Progress p) throws Exception {
        ModPGroup res = null;
        ExecutorService executor = Executors.newFixedThreadPool(ctx.args.threads.value());
        List<Future<ModPGroup>> futures = new ArrayList<>();
        for (int i = 0; i < ctx.args.threads.value(); i++) {
            FutureTask<ModPGroup> ft = new FutureTask<>(new Callable<ModPGroup>() {
                @Override
                public ModPGroup call() throws IllegalArgumentException, IOException {
                    ModPGroup g;
                    while (!Thread.currentThread().isInterrupted()) {
                        try {
                            p.increase(1);
                            g = new ModPGroup(len, rnd, 1);
                        } catch (MathException e) {
                            continue;
                        }
                        return g;
                    }
                    return null;
                }
            });
            executor.submit(ft);
            futures.add(ft);
        }
        Iterator<Future<ModPGroup>> it = futures.iterator();
        while (it.hasNext()) {
            Future<ModPGroup> future = it.next();
            if (!it.hasNext()) {
                it = futures.iterator();
            }
            try {
                res = future.get(10, TimeUnit.MILLISECONDS);
            } catch (InterruptedException e) {
                continue;
            } catch (ExecutionException e) {
                throw e;
            } catch (TimeoutException e) {
                continue;
            }
            break;
        }
        executor.shutdownNow();
        return res;
    }

    private GroupElement generatorGen(Group group, Rnd rnd) throws IOException {
        if (group instanceof ModPGroup) {
            return ((ModPGroup) group).getRandomElement(rnd);
        } else if (group instanceof ECGroup) {
            return ((ECGroup) group).getBasePoint();
        } else {
            throw new IllegalArgumentException("Unknown group");
        }
    }

    private void writeParameters(Group group, GroupElement generator, Path template)
            throws IOException {
        String templatestr;
        if (group instanceof ModPGroup) {
            templatestr = String.format("  paramtype:\n    mod:\n      p: %s\n      g: %s\n",
                    group.getOrder(), ((ModPGroupElement) generator).getValue());
        } else if (group instanceof ECGroup) {
            templatestr = String.format("  paramtype:\n    ec:\n      name: %s\n",
                    ((ECGroup) group).getCurveName());
        } else {
            throw new IllegalArgumentException("Unknown group");
        }
        Files.write(template, Util.toBytes(templatestr));
    }

    public static class GroupGenArgs extends Args {
        Arg<Integer> len = Arg.anInt(Msg.g_length);
        Arg<String> paramType = Arg.aChoice(Msg.arg_paramtype, MOD_GROUP, EC_GROUP);
        Arg<Path> initTemplate = Arg.aPath(Msg.g_init_template, false, false);
        Arg<List<RandomSourceArg.RndListEntry>> random = RandomSourceArg.getArgument();

        public GroupGenArgs() {
            super();
            args.add(len);
            args.add(random);
            args.add(paramType);
            args.add(initTemplate);
        }
    }
}
