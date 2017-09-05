package ee.ivxv.processor;

import ch.qos.cal10n.BaseName;
import ch.qos.cal10n.LocaleData;
import ee.ivxv.common.util.NameHolder;

@BaseName("i18n.processor-msg")
@LocaleData(defaultCharset = "UTF-8", value = {})
public enum Msg implements NameHolder {
    /*-
     * The part of the enum name until the first '_' (including) is excluded from the getName().
     * This is a means to provide multiple translations for the same tool or argument name.
     */

    // App
    app_processor,

    // Tools
    tool_check, tool_squash, tool_revoke, tool_anonymize, tool_export,

    // Tool arguments
    arg_ballotbox("bb"), arg_ballotbox_checksum("bbcs"), //
    arg_registrationlist, arg_registrationlist_checksum, //
    arg_districts, arg_revocationlists, //
    arg_tskey, arg_vlkey, //
    arg_voterlists, arg_path, arg_signature, //
    arg_voter_id, //
    arg_election_start, //
    arg_out("o"),

    // Error messages
    e_tehcnical_error,

    e_vl_vlkey_missing, //
    e_vl_first_not_initial, e_vl_initial_not_first, //
    e_vl_read_error, e_vl_signature_error, e_vl_election_id, //
    e_vl_invalid_header, e_vl_invalid_voter_row, e_vl_invalid_row_number, //
    e_vl_invalid_district, e_vl_invalid_station, //
    e_vl_voter_already_removed, e_vl_removed_voter_missing, //
    e_vl_voter_already_added, e_vl_added_voter_exists, //
    e_vl_fictive_single_district_and_station_required, //
    e_vl_error_report,

    e_bb_read_error, e_bb_ballot_processing, e_reg_record_processing, //
    e_bb_invalid_file_name, e_bb_missing_file, e_bb_repeated_file, e_bb_unknown_file_type, //
    e_ocsp_resp_invalid, e_ocsp_resp_status_not_suffessful, e_ocsp_resp_cert_status_not_good, //
    e_ocsp_resp_not_basic, e_ocsp_resp_sig_not_valid, e_ocsp_resp_issuer_unknown, //
    e_ballot_signature_invalid, e_ballot_missing_voter_signature, e_active_voter_not_found, //
    e_time_before_start, //
    e_reg_resp_invalid, e_reg_req_invalid, //
    e_reg_resp_not_unique, e_reg_req_not_unique, //
    e_reg_resp_no_nonce, e_reg_resp_nonce_not_sig, e_reg_resp_nonce_alg_mismatch, //
    e_reg_resp_nonce_sig_invalid, //
    e_unknown_file_in_vote_container, //
    e_reg_resp_req_unmatch, e_reg_req_without_ballot, e_ballot_without_reg_req, //
    e_same_time_as_latest, //
    e_bb_error_report,

    e_rl_read_error, e_rl_election_id, //
    e_rl_processing_error, //
    e_rl_voter_not_found_in_bb, e_rl_ballot_already_revoked, e_rl_ballot_already_restored,

    e_writing_ivoter_list, //
    e_writing_revocation_report, //
    e_writing_log_n, //
    e_writing_error_report,

    // Messages
    m_output_file, //
    m_read, //

    m_vl_reading, //
    m_vl, //
    m_vl_type, //
    m_vl_total_added, //
    m_vl_total_removed, //
    m_vl_fictive_warning, //
    m_vl_fictive_voter_name,

    m_rl_loading, //
    m_rl_loaded, //
    m_rl_arg_for_cont, //
    m_rl_checking_integrity, //
    m_rl_data_is_integrous, //
    m_rl_revoke_total, //
    m_rl_restore_total, //
    m_rl_revoke_start, //
    m_rl_revoke_ballots_before, //
    m_rl_revoke_count, //
    m_rl_revoke_ballots_after, //
    m_rl_revoke_done, //
    m_rl_restore_start, //
    m_rl_restore_ballots_before, //
    m_rl_restore_count, //
    m_rl_restore_ballots_after, //
    m_rl_restore_done, //

    m_removing_recurrent_votes, //
    m_applying_revocation_lists, //
    m_anonymizing_ballot_box,

    m_writing_ivoter_list, //
    m_writing_revocation_report, //
    m_writing_log_n, //

    ;

    private final String shortName;

    private Msg() {
        this(null);
    }

    private Msg(String shortName) {
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
