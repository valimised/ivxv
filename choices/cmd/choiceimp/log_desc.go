package main

const (
	_CHOICES_VERSION_FROM_DB           = "Cannot fetch choices list version from a database"
	_CHOICES_LIST_BDOC_OPEN            = "Failed to open choices list BDOC container"
	_CHOICES_LIST_BDOC_NO_SIG          = "No signatures found in choices list BDOC container"
	_CHOICES_LIST_BDOC_SIG_INFO        = "Choices list BDOC container signature info"
	_CHOICES_LIST_BDOC_SIG_TO_JSON     = "Failed to JSON marshal choices list BDOC container signature"
	_CHOICES_LIST_BDOC_HAS_MANY_FILES  = "Choices list BDOC container should only have 1 file"
	_CHOICES_LIST_READ                 = "Reading choices list"
	_CHOICES_LIST_TO_DB_UPLOAD         = "Failed to upload choices list to database"
	_CHOICES_LIST_INVALID_JSON         = "Choices list has an invalid JSON format"
	_CHOICES_LIST_ID_MISMATCH          = "Choices list ID doesn't match the one used in election configuration file"
	_CHOICES_LIST_PREPARE_UPLOAD_TO_DB = "Choices list is ready to uploaded into database"
	_CHOICES_LIST_UPLOAD_TO_DB_FAIL    = "Unable to upload choices list to database"
)
