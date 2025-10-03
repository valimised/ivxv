package main

const (
	_DISTRICTS_VERSION_FROM_DB                          = "Cannot fetch districts list version from a database"
	_DISTRICTS_LIST_BDOC_OPEN                           = "Failed to open district list BDOC container"
	_DISTRICTS_LIST_BDOC_NO_SIG                         = "No signatures found in districts list BDOC container"
	_DISTRICTS_LIST_BDOC_SIG_INFO                       = "Districts list BDOC container signature info"
	_DISTRICTS_LIST_BDOC_SIG_TO_JSON                    = "Failed to JSON marshal districts list BDOC container signature"
	_DISTRICTS_LIST_BDOC_HAS_MANY_FILES                 = "Districts list BDOC container should only have 1 file"
	_DISTRICTS_LIST_READ                                = "Reading districts list"
	_DISTRICTS_LIST_TO_DB_UPLOAD                        = "Failed to upload districts list to database"
	_DISTRICTS_LIST_INVALID_JSON                        = "Choices list has an invalid JSON format"
	_DISTRICTS_LIST_ID_MISMATCH                         = "Choices list ID doesn't match the one used in election configuration file"
	_DISTRICTS_LIST_INVALID_COUNTIES_JSON               = "District list has an invalid 'counties' JSON"
	_DISTRICTS_LIST_PREPARE_UPLOAD_TO_DB                = "Choices list is ready to uploaded into database"
	_DISTRICTS_LIST_UPLOAD_TO_DB_FAIL                   = "Unable to upload districts list to database"
	_DISTRICTS_LIST_INVALID_DISTRICT_ID                 = "District ID has an invalid format"
	_DISTRICTS_LIST_PARISH_CONTAINS_SAME_DISTRICT_CODES = "Only 1 district code is allowed per parish"
)
