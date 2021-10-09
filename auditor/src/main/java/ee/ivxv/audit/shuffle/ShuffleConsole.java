package ee.ivxv.audit.shuffle;

import ee.ivxv.audit.Msg;
import ee.ivxv.common.service.console.Progress;
import ee.ivxv.common.service.i18n.Translatable;
import ee.ivxv.common.util.I18nConsole;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.atomic.AtomicInteger;

public class ShuffleConsole {
    private final I18nConsole console;

    public ShuffleConsole(I18nConsole console) {
        this.console = console;
    }

    public void enter(ShuffleStep step) {
        int totalSteps = step.parent != null ? step.parent.subSteps.size() : 0;
        int thisStep = step.parent != null ? step.parent.nextSubStep() : 0;
        String stepstr = "";
        if (totalSteps > 0) {
            stepstr = String.format("(%d/%d): ", thisStep, totalSteps);
        }
        if (step.i18msg != null) {
            console.println(Msg.m_shuffle_step, step.getDepth(), stepstr, step.i18msg);
        } else if (step.msg != null) {
            console.println(Msg.m_shuffle_step, step.getDepth(), stepstr, step.msg);
        } else {
            console.println(Msg.m_shuffle_step, step.getDepth(), stepstr, "");
        }
    }

    public Progress enter(ShuffleStep step, long length) {
        enter(step);
        return console.startProgress(length, true);
    }

    public static enum ShuffleStep {
        READ(Msg.m_shuffle_read), //
        READ_PROT_INFO(READ, Msg.m_shuffle_read_prot_info), //
        READ_PUBKEY(READ, Msg.m_shuffle_read_pubkey), //
        READ_PC(READ, Msg.m_shuffle_read_pc), READ_POSC(READ, Msg.m_shuffle_read_posc), //
        READ_POSR(READ, Msg.m_shuffle_read_posr), //
        READ_CIPHS(READ, Msg.m_shuffle_read_ciphs), //
        READ_SHUFFLED(READ, Msg.m_shuffle_read_shuffled),

        VERIFY(Msg.m_shuffle_verify), //
        VERIFY_PARAMS(VERIFY, Msg.m_shuffle_verify_params), //
        VERIFY_NI(VERIFY, Msg.m_shuffle_verify_ni), //
        VERIFY_PERM(VERIFY, Msg.m_shuffle_verify_permutation), //
        VERIFY_RERAND(VERIFY, Msg.m_shuffle_verify_rerandomisation);

        final String msg;
        final Translatable i18msg;
        final ShuffleStep parent;
        final List<ShuffleStep> subSteps = new ArrayList<ShuffleStep>();
        final AtomicInteger completed = new AtomicInteger();

        ShuffleStep(Translatable msg) {
            this.msg = null;
            this.i18msg = msg;
            this.parent = null;
        }

        ShuffleStep(String msg) {
            this.msg = msg;
            this.i18msg = null;
            this.parent = null;
        }

        ShuffleStep(ShuffleStep parent, Translatable msg) {
            this.msg = null;
            this.i18msg = msg;
            this.parent = parent;
            parent.appendSubStep(this);
        }

        int getDepth() {
            int i = 0;
            ShuffleStep next = this.parent;
            while (next != null) {
                next = next.parent;
                i++;
            }
            return i;
        }

        void appendSubStep(ShuffleStep substep) {
            subSteps.add(substep);
        }

        int subStepCount() {
            return subSteps.size();
        }

        int nextSubStep() {
            return this.completed.incrementAndGet();
        }
    }
}
