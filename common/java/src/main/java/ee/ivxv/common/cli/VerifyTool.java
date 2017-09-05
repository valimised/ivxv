package ee.ivxv.common.cli;

import ee.ivxv.common.M;
import ee.ivxv.common.cli.VerifyTool.VerifyArgs;
import ee.ivxv.common.service.container.Container;
import ee.ivxv.common.service.container.ContainerReader;
import ee.ivxv.common.util.I18nConsole;
import java.nio.file.Path;

public class VerifyTool implements Tool.Runner<VerifyArgs> {

    private final I18nConsole console;
    private final ContainerReader container;

    public VerifyTool(AppContext<?> ctx) {
        console = new I18nConsole(ctx.i.console, ctx.i.i18n);
        container = ctx.container;
    }

    @Override
    public boolean run(VerifyArgs args) throws Exception {
        verify(args.file.value());

        return true;
    }

    private void verify(Path path) throws Exception {
        console.println();
        console.println(M.m_reading_container, path);
        container.requireContainer(path);
        Container c = container.read(path.toString());

        console.println(M.m_signatures);
        c.getSignatures().forEach(s -> console.println(M.m_signature_row,
                s.getSigner().getSerialNumber(), s.getSigner().getName(), s.getSigningTime()));

        console.println(M.m_files);
        c.getFiles().forEach(f -> console.println(M.m_file_row, f.getName()));
    }

    public static class VerifyArgs extends Args {

        Arg<Path> file = Arg.aPath(Msg.arg_file, true, false);

        public VerifyArgs() {
            args.add(file);
        }

    }

}
