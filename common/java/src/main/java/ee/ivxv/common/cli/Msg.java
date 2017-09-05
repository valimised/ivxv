package ee.ivxv.common.cli;

import ch.qos.cal10n.BaseName;
import ch.qos.cal10n.LocaleData;
import ee.ivxv.common.util.NameHolder;

@BaseName("i18n.common-cli-msg")
@LocaleData(defaultCharset = "UTF-8", value = {})
public enum Msg implements NameHolder {
    // Used by AppRunner
    e_app_error, e_common_args_invalid, e_multiple_tools, e_tool_args_invalid, e_tool_error, //
    e_tool_missing, e_unknown_arg, e_unknown_args_present, e_unknown_tool, //

    w_test_mode,

    app, tool, row_indent, tool_placeholder, value_placeholder, //
    usage, usage_arg_w_value, usage_arg_names, usage_arg_optional, //
    tools, tool_row, //
    args, arg_row, arg_name_required, //
    params, param_row, param_name_required, //

    app_result_success, //
    app_result_failure, //

    // Used by Arg, CommandLine and YamlData
    e_arg_parse_error, //
    e_invalid_boolean, e_invalid_choice, e_invalid_int, e_invalid_number, e_invalid_instant, //
    e_invalid_path_exists, e_invalid_path_not_exists, e_invalid_path_not_dir, //
    e_invalid_path_not_file, e_invalid_public_key, //
    e_multiple_assignments, e_requires_single_value, e_value_not_allowed, e_branch_expected, //
    e_yaml_invalid_file, e_yaml_invalid_key, e_yaml_list_expected, e_yaml_map_expected, //
    e_yaml_scalar_expected, //
    e_arg_required, //

    // Common tools
    tool_verify,

    // Common arguments
    arg_help("h"), arg_conf("c"), arg_params("p"), arg_force("f"), arg_quiet("q"), arg_lang, //
    arg_container_threads("ct"), arg_threads("t"),

    // Verify tool arguments
    arg_file;

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
