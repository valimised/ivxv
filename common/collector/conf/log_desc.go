package conf

const (
	_CONF_START               = "Starting to parse configuration .bdoc container"
	_CONF_TRUST               = "Failed to parse trust configuration .bdoc container"
	_CONF_ELECTION            = "Failed to parse election configuration .bdoc container"
	_CONF_TECH                = "Failed to parse technical configuration .bdoc container"
	_CONF_TRUST_READ          = "Failed to read trust configuration .bdoc container"
	_CONF_TRUST_BDOC_EXT      = "trust configuration container doesn't have an extension (.bdoc)"
	_CONF_TRUST_OPEN          = "Failed to open trust configuration .bdoc container"
	_CONF_TRUST_CONF          = "Failed to read trust configuration file"
	_CONF_BDOC_OPENER         = "Configure container opener based on a trust configuration file"
	_CONF_REWIND              = "Rewind (back to beginning) trust configuration .bdoc container reader's position pointer"
	_CONF_BDOC_VERIFY         = "Failed to verify .bdoc container"
	_CONF_TRUST_NO_SIG        = "No signatures found in trust configuration .bdoc container"
	_CONF_TRUST_BDOC_SIG_INFO = "Summary info of trust configuration .bdoc container"
	_CONF_BDOC_READ           = "Unable to read .bdoc container"
	_CONF_BDOC_NO_SIG         = "No signatures found in .bdoc container"
	_CONF_BDOC_SIG_INFO       = "Summary info of .bdoc container"
	_CONF_BDOC_DATA           = "Failed to read .bdoc container data"
	_CONF_YAML_READ           = "Failed to unmarshal YAML data from a .bdoc container"
	_CONF_YAML_NO_DATA        = "No such data in a .bdoc container"
)
