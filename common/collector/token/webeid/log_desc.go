package webeid

const (
	_WEBEID_REGEX      = "Web-eID token regex failed"
	_WEBEID_VER        = "Web-eID token major release is not supported by IVXV"
	_WEBEID_HASH_ALG   = "Hashing Web-eID origin with specified in Web-eID token algorithm failed"
	_WEBEID_HASH_NONCE = "Hashing Web-eID nonce with specified in Web-eID token algorithm failed"
	_WEBEID_SIG        = "Failed to verify Web-eID token signature"
	_WEBEID_JSON       = "JSON unmarshal Web-eID token failed"
	_WEBEID_CERT_B64   = "base-64 decoding Web-eID token certificate failed"
	_WEBEID_SIG_B64    = "base-64 decoding Web-eID token signature failed"
	WEBEID_ALG         = "Unknown hashing algorithm"
	WEBEID_ALG_REGEX   = "Hashing algorithm regex failed"
)
