package tsp

const (
	_TSA_CFG           = "Reading YAMl configuration of OCSP client failed"
	_TSA_NEW           = "Failed to create new OCSP client"
	_TSA_KEY           = "Failed to read TSA client private key from file system"
	_TSA_KEY_PARSE     = "Failed to parse PEM TSA client private key"
	_TSA_KEY_PARSE_RSA = "Failed to parse RSA TSA client private key"
	_TSA_NO_SIG        = "No signatures in a .bdoc container"
	_TSA_DATAER        = ".bdoc container cannot be casted to TimestampDataer interface"
	_TSA_SIG_VERIFY    = "Failed to verify .bdoc signature timestamp"
	_TSA_SIG_TS        = "Failed to sign .bdoc signature timestamp"
	_TSA_REQ           = "Failed to create client timestamp request to TSA provider"
	_TSA_RSA_SIG       = "Failed to sign data using RSA algorithm"
	_TSA_VERIFY        = "Failed to verify OCSP response of a .bdoc container"
	_TSA_UNKNOWN       = "Unknown OCSP response status"
	_TSA_REVOKED       = "OCSP response status is 'revoked'"
)
