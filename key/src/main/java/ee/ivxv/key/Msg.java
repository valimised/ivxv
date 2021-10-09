package ee.ivxv.key;

import ch.qos.cal10n.BaseName;
import ch.qos.cal10n.LocaleData;
import ee.ivxv.common.util.NameHolder;

@BaseName("i18n.key-msg")
@LocaleData(defaultCharset = "UTF-8", value = {})
public enum Msg implements NameHolder {
    /*-
     * The part of the enum name until the first '_' (including) is excluded from the getName().
     * This is a means to provide multiple translations for the same tool or argument name.
     */

    // App
    app_key,

    // Tools
    tool_decrypt, tool_groupgen, tool_init, tool_util, tool_testkey,

    // Common tool arguments
    arg_identifier("i"), arg_parties("n"), arg_threshold("m"), arg_out("o"), //
    arg_mod, arg_ec, arg_paramtype, arg_random_source("r"), arg_random_source_type, //
    arg_random_source_path, arg_fastmode,

    // 'decrypt' tool arguments
    d_anonballotbox, d_anonballotbox_checksum, //
    d_questioncount, d_candidates, d_districts, d_recover, d_protocol, //
    d_provable, d_check_decodable, //

    // 'groupgen' tool arguments
    g_length("l"), g_init_template,

    // 'init' tool arguments
    i_p, i_g, i_name, //
    i_desmedt("d"), i_genprotocol, i_signaturekeylen("s"), //
    i_signcn, i_signsn, i_enccn, i_encsn, i_skiptest, //
    i_required_randomness,

    // 'util' tool arguments
    u_listreaders, u_testkey,

    // error
    e_testencryption_fail, e_quorum_test_fail, e_no_cardterminals_found, //
    e_abb_invalid_question_count, e_illegal_vote_district, e_illegal_vote_parish,

    // messages
    m_id, m_name, m_with_card, m_yes, m_no, m_quorum_test_ok, m_gen_group_params, //
    m_certificates_generated, m_generate_decryption_key, m_test_decryption_key, //
    m_generate_signature_key, m_test_signature_key, m_votecount, //
    m_abb_dist_verifying, m_abb_dist_ok, m_protocol_init, m_protocol_init_ok, //
    m_dec_start, m_dec_done, m_out_tally, m_out_proof, m_out_invalid, m_out_logs, //
    m_keys_saved, m_collecting_required_randomness, m_with_proof, m_without_proof, m_card_id, //
    m_fastmode_disabled, m_fastmode_enabled, m_storing_shares, m_generating_certificate;


    private final String shortName;

    Msg() {
        this(null);
    }

    Msg(String shortName) {
        this.shortName = shortName;
    }

    @Override
    public String getShortName() {
        return shortName;
    }

    @Override
    public String getName() {
        return extractName(name());
    }

    @Override
    public Enum<?> getKey() {
        return this;
    }
}
