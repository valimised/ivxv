package tls

const (
	_TLS_CFG                   = "Cannot YAML parse TLS authentication configuration"
	_TLS_NEW                   = "Failed to configure TLS authentication utility"
	_TLS_CA                    = "No CA certificate that would verify client certificates provided"
	_TLS_CA_PARSE              = "Failed to parse CA certificate"
	_TLS_INTERMEDIATE_CA_PARSE = "Failed to parse intermediate CA certificate"
	_TLS_OCSP                  = "Failed to configure OCSP client for voter certificate validity verification"
	_TLS_TICKET                = "TLS authenticated client should not support ticket authentication"
	_TLS_CERT                  = "Voter hasn't provided certificates on TLS authentication"
	_TLS_CERT_VERIFY           = "Voter certificate verification failed"
	_TLS_CERT_VERIFY_OCSP      = "Voter certificate OCSP verification failed"
	_TLS_CERT_VERIFY_OCSP_BAD  = "Voter certificate OCSP status is not good"
)
