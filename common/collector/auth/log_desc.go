package auth

const (
	_AUTH_MALFORMED  = "Wraps errors which are cause by malformed authentication tokens"
	_AUTH_CERT       = "Wraps errors which are cause by invalid authentication certificates"
	_AUTH_AUTHORIZED = "Wraps errors where authentication succeeded, but the client is not authorized to use any services"
	_AUTH_UNKNOWN    = "Unknown authentication method"
	_AUTH_CFG        = "Failed to configure authentication utility"
)
