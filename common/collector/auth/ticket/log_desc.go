package ticket

const (
	_TICKET_NEW_COOKIE = "Failed to create new cookie from provided AES key"
	_TICKET_KEY        = "Failed to read AES key file from file system"
	_TICKET_NEW        = "Failed to create authentication ticket (cookie)"
	_TICKET_VOTEID     = "Failed to create vote ID (16-byte random value)"
	_TICKET_JSON       = "Failed to JSON marshal authentication ticket content"
	_TICKET_VERIFY     = "Failed to verify authentication ticket passed by a client"
	_TICKET_UJSON      = "Failed to JSON unmarshal authentication ticket content"
)
