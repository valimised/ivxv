package ee.ivxv.common.conf;

import ch.qos.cal10n.BaseName;
import ch.qos.cal10n.LocaleData;
import ee.ivxv.common.service.i18n.Translatable;

@BaseName("i18n.common-conf-msg")
@LocaleData(defaultCharset = "UTF-8", value = {})
public enum Msg implements Translatable {
    // ConfLoader errors
    e_conf_file_not_found, e_conf_file_open_error, e_unsupported_conf_type, e_unused_file, //
    e_ca_not_ca_cert, e_ocsp_not_ocsp_cert, e_tsp_not_tsp_cert,

    w_unknown_property,

    // LocaleConf errors
    e_unsupported_locale, //

    m_loading_conf, //
    m_conf_arg_for_cont,

    ;

    @Override
    public Enum<?> getKey() {
        return this;
    }

}
