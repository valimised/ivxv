package ee.ivxv.common.util;

import ee.ivxv.common.M;
import ee.ivxv.common.service.container.Container;
import ee.ivxv.common.service.container.DataFile;
import ee.ivxv.common.service.i18n.MessageException;
import java.util.stream.Collectors;

/**
 * ContainerHelper is a dedicated class containing helper methods for handling document containers.
 */
public class ContainerHelper {

    private final I18nConsole console;
    private final Container c;

    public ContainerHelper(I18nConsole console, Container c) {
        this.console = console;
        this.c = c;
    }

    public String getSignerNames() {
        return c.getSignatures().stream().map(s -> s.getSigner().getName())
                .collect(Collectors.joining(", "));
    }

    /**
     * Checks that the container has signatures and contains a single file. Also reports standard
     * messages about checking signature, signer name, signing time and the result on the console.
     * 
     * @param nameParam The first parameter for {@code M.m_cont_*} messages describing the container
     * @return
     * @throws MessageException
     */
    public DataFile getSingleFileAndReport(Object nameParam) throws MessageException {
        if (c.getFiles().size() != 1) {
            throw new MessageException(M.e_cont_single_file_expected, c.getFiles().size());
        }

        reportSignatures(nameParam);

        return c.getFiles().get(0);
    }

    public void reportSignatures(Object nameParam) throws MessageException {
        if (c.getSignatures().isEmpty()) {
            throw new MessageException(M.e_cont_signature_expected);
        }

        console.println(M.m_cont_checking_signature, nameParam);
        c.getSignatures().forEach(s -> {
            console.println(M.m_cont_signer, nameParam, s.getSigner().getName());
            console.println(M.m_cont_signature_time, nameParam, s.getSigningTime());
        });
        console.println(M.m_cont_signature_is_valid, nameParam);
    }

}
