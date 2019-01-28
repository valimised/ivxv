package ee.ivxv.common.service.smartcard;

import ch.qos.cal10n.BaseName;
import ch.qos.cal10n.LocaleData;

@BaseName("i18n.common-smartcard-msg")
@LocaleData(defaultCharset = "UTF-8", value = {})
public enum Msg {
    insert_card_indexed, remove_card_indexed, enter_terminal_id, //
    insert_unprocessed_card, inserted_card_id,

    // APDU
    enter_pin_keyboard, enter_pin_pinpad, verify_fail_retry
}
