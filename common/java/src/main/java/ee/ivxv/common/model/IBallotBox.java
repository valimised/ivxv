package ee.ivxv.common.model;

import ch.qos.cal10n.BaseName;
import ch.qos.cal10n.LocaleData;
import ee.ivxv.common.M;
import ee.ivxv.common.service.i18n.MessageException;
import ee.ivxv.common.service.i18n.Translatable;

public interface IBallotBox {

    String getElection();

    Type getType();

    int getNumberOfBallots();

    default void requireType(Type t) {
        if (getType() != t) {
            throw new MessageException(M.e_bb_invalid_type, t, getType());
        }
    }

    @BaseName("i18n.common-model-bb-type")
    @LocaleData(defaultCharset = "UTF-8", value = {})
    enum Type implements Translatable {
        UNORGANIZED, //
        BACKUP, //
        INTEGRITY_CONTROLLED, //
        RECURRENT_VOTES_REMOVED, //
        DOUBLE_VOTERS_REMOVED, //
        ANONYMIZED;

        @Override
        public Enum<?> getKey() {
            return this;
        }
    }

}
