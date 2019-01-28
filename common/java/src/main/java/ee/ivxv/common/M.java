package ee.ivxv.common;

import ch.qos.cal10n.BaseName;
import ch.qos.cal10n.LocaleData;
import ee.ivxv.common.service.i18n.Translatable;

/**
 * Generic re-usable messages for common module and applications.
 */
@BaseName("i18n.common-msg")
@LocaleData(defaultCharset = "UTF-8", value = {})
public enum M implements Translatable {
    // Error messages
    e_bb_invalid_type, //
    e_cert_not_found, //
    e_cert_read_error, //
    e_checksum_mismatch, //
    e_cont_signature_expected, //
    e_cont_single_file_expected, //
    e_decryption_error, //
    e_file_not_container, //
    e_file_not_found, //
    e_invalid_container, //

    e_dist_read_error, //
    e_dist_station_id_not_unique, //
    e_dist_station_id_invalid, //
    e_dist_station_region_unknown, //
    e_dist_bb_station_missing, //

    e_cand_read_error, //
    e_cand_invalid_dist, //
    e_cand_duplicate_id, //

    m_election_id, //
    m_datetime_pattern, //
    m_progress_bar, //

    m_loading_params, //
    m_params_arg_for_cont, //

    m_reading_container, //
    m_files, //
    m_file_row, //
    m_signatures, //
    m_signature_row, //

    m_cont_checking_signature, //
    m_cont_signer, //
    m_cont_signature_time, //
    m_cont_signature_is_valid, //

    m_checksum_loading, //
    m_checksum_loaded, //
    m_checksum_arg_for_cont, //
    m_checksum_calculate, //
    m_checksum_ok, //

    m_bb_arg_for_checksum, //
    m_bb_loading, //
    m_bb_loaded, //
    m_bb_checking_type, //
    m_bb_type, //
    m_bb_numof_collector_ballots, //
    m_bb_numof_ballots, //
    m_bb_checking_integrity, //
    m_bb_data_is_integrous, //
    m_bb_checking_ballot_sig, //
    m_bb_total_ballots, //
    m_bb_numof_ballots_sig_valid, //
    m_bb_numof_ballots_sig_invalid, //
    m_bb_all_ballots_sig_valid, //
    m_bb_compare_with_reg, //
    m_bb_ballot_missing_reg, //
    m_bb_reg_missing_ballot, //
    m_bb_in_compliance_with_reg, //
    m_reg_in_compliance_with_bb, //
    m_bb_total_checked_ballots, //
    m_bb_exporting, //
    m_bb_exporting_voter, //
    m_bb_exported, //

    m_reg_arg_for_checksum, //
    m_reg_loading, //
    m_reg_loaded, //
    m_reg_numof_records, //
    m_reg_checking_integrity, //
    m_reg_data_is_integrous, //

    m_bb_saving, //
    m_bb_saved, //
    m_bb_checksum_saving, //
    m_bb_checksum_saved, //

    m_cand_loading, //
    m_cand_arg_for_cont, //
    m_cand_loaded, //
    m_cand_count, //

    m_dist_loading, //
    m_dist_arg_for_cont, //
    m_dist_loaded, //
    m_dist_count,

    m_proof_loading, //
    m_proof_loaded, //
    m_proof_count, //

    m_out_start, //
    m_out_done, //

    r_ivl_description, //
    r_ivl_station_name, //
    ;

    @Override
    public Enum<?> getKey() {
        return this;
    }

}
