package ee.ivxv.key.tool;

import ee.ivxv.common.cli.Arg;
import ee.ivxv.common.cli.Args;
import ee.ivxv.common.cli.Tool;
import ee.ivxv.common.crypto.rnd.Rnd;
import ee.ivxv.common.math.IntegerConstructor;
import ee.ivxv.common.util.I18nConsole;
import ee.ivxv.key.KeyContext;
import ee.ivxv.key.Msg;
import ee.ivxv.key.RandomSourceArg;
import ee.ivxv.key.tool.GroupGenTool.GroupGenArgs;
import java.io.IOException;
import java.math.BigInteger;
import java.util.List;
import org.bouncycastle.pqc.math.linearalgebra.IntegerFunctions;

public class GroupGenTool implements Tool.Runner<GroupGenArgs> {

    private static final String MOD_GROUP = "mod";
    private static final String EC = "ec";

    private final I18nConsole console;

    public GroupGenTool(KeyContext ctx) {
        this.console = new I18nConsole(ctx.i.console, ctx.i.i18n);
    }

    @Override
    public boolean run(GroupGenArgs args) throws Exception {
        if (args.paramType.value().equals(MOD_GROUP)) {
            console.println(Msg.m_gen_group_params, args.len.value());
            Rnd rnd = RandomSourceArg.combineFromArgument(args.random);
            modGroupGen(args.len.value(), rnd);
            rnd.close();
            return true;

        } else if (args.paramType.value().equals(EC)) {
            throw new UnsupportedOperationException(
                    "EC group generation not yet implemented. Use standard groups.");
        }
        return false;
    }

    private int modGroupGen(int len, Rnd rnd) throws IOException {
        int sglen = len - 1;
        BigInteger p;
        BigInteger q;
        BigInteger g;
        BigInteger bigTwo = new BigInteger("2");

        // For performance reasons, we first test primes lightly.
        // If both pass, test them thoroughly.
        while (true) {
            q = IntegerConstructor.construct(rnd, bigTwo.pow(sglen).subtract(BigInteger.ONE));
            if (!q.isProbablePrime(2)) {
                continue;
            }
            p = q.multiply(bigTwo).add(BigInteger.ONE);
            if (!(p.bitLength() == len && p.isProbablePrime(2))) {
                continue;
            }

            if (p.isProbablePrime(128) && q.isProbablePrime(128)) {
                break;
            }
        }

        while (true) {
            g = IntegerConstructor.construct(rnd, p);
            if (IntegerFunctions.jacobi(g, p) == 1) {
                break;
            }
        }

        console.console.println("Zp(p=%s, q=%s, g=%s)", p.toString(), q.toString(), g.toString());
        return 0;
    }

    public static class GroupGenArgs extends Args {
        Arg<Integer> len = Arg.anInt(Msg.g_length);
        Arg<String> paramType = Arg.aChoice(Msg.arg_paramtype, MOD_GROUP, EC);
        Arg<List<RandomSourceArg.RndListEntry>> random = RandomSourceArg.getArgument();

        public GroupGenArgs() {
            super();
            args.add(len);
            args.add(random);
            args.add(paramType);
        }
    }
}
