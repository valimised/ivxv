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
    tool_check, tool_squash, tool_revoke, tool_anonymize, tool_export, tool_stats, tool_statsdiff,

    // Tool arguments
    arg_ballotbox("bb"), arg_ballotbox_checksum("bbcs"), //
    arg_registrationlist, arg_registrationlist_checksum, //
    arg_districts, arg_revocationlists, //
    arg_tskey, arg_vlkey, //
    arg_voterlists, arg_path, arg_signature, arg_skip_cmd, //
    arg_districts_mapping, //
    arg_voter_id, //
    arg_election_start, //
    arg_voterforeignehak, //
    arg_enckey, //
    arg_election_day, arg_period_start, arg_period_end, //
    arg_compare, arg_to, arg_diff, //
    arg_out("o"),

    // Error messages
    e_tehcnical_error,

    e_vl_voterlists_missing, e_vl_vlkey_missing, //
    e_vl_first_not_initial, e_vl_initial_not_first, //
    e_vl_read_error, e_vl_signature_error, e_vl_election_id, //
    e_vl_invalid_header, e_vl_invalid_voter_row, e_vl_invalid_row_number, //
    e_vl_invalid_changeset, e_vl_invalid_version, e_vl_invalid_time, //
    e_vl_invalid_district, e_vl_invalid_parish, //
    e_vl_voter_already_removed, e_vl_removed_voter_missing, //
    e_vl_voter_already_added, e_vl_added_voter_exists, //
    e_vl_fictive_single_district_and_parish_required, //
    e_vl_error_report,

    e_dist_mapping_invalid_row,
    e_skip_cmd_loading,
    e_reg_checksum_missing, //
    e_bb_read_error, e_bb_ballot_processing, e_reg_record_processing, //
    e_bb_invalid_file_name, e_bb_missing_file, e_bb_repeated_file, e_bb_unknown_file_type, //
    e_ballot_signature_invalid, e_ballot_missing_voter_signature, //
    e_active_voter_not_found, e_active_voterlist_not_found, //
    e_time_before_start, //
    e_reg_resp_invalid, e_reg_req_invalid, //
    e_reg_resp_not_unique, e_reg_req_not_unique, //
    e_reg_resp_no_nonce, e_reg_resp_nonce_not_sig, e_reg_resp_nonce_alg_mismatch, //
    e_reg_resp_nonce_sig_invalid, //
    e_unknown_file_in_vote_container, //
    e_reg_resp_req_unmatch, e_reg_req_without_ballot, e_ballot_without_reg_req, //
    e_same_time_as_latest, e_invalid_signature_profile, //
    e_bb_error_report,

    e_rl_read_error, e_rl_election_id, //
    e_rl_processing_error, //
    e_rl_voter_not_found_in_bb, e_rl_ballot_already_revoked, e_rl_ballot_already_restored,

    e_bb_ciphertext_checking, //
    e_ciphertext_invalid, //
    e_ciphertext_invalid_bytes, //
    e_ciphertext_invalid_group, //
    e_ciphertext_invalid_point, //
    e_ciphertext_invalid_qr, //
    e_ciphertext_invalid_range, //

    e_writing_ivoter_list, //
    e_writing_revocation_report, //
    e_writing_log_n, //
    e_writing_error_report,

    e_stats_code_not_estonian,

    // Messages
    m_output_file, //
    m_read, //

    m_bb_unsigned_skipping_output, //
    m_reg_skipping_compare, //
    m_bb_grouping_votes_by_voter, //

    m_vl_reading, //
    m_vl, //
    m_vl_type, //
    m_vl_skipped, //
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

    m_dist_mapping_loading, //
    m_dist_mapping_arg_for_cont, //
    m_dist_mapping_loaded, //

    m_removing_recurrent_votes, //
    m_removing_invalid_ciphertexts, //
    m_applying_revocation_lists, //
    m_anonymizing_ballot_box,

    m_writing_ivoter_list, //
    m_writing_revocation_report, //
    m_writing_log_n, //

    m_stats_generating, m_stats_generated, m_stats_ballot_errors, m_stats_valid_ballots, //
    m_stats_json_saved, m_stats_csv_saved, m_stats_diff_saved, //

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
