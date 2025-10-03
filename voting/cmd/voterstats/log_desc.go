package main

const (
	_VOTERSTATS_START            = "Failed to transform 'count votes from' time from election configuration file to RFC3339 format"
	_VOTERSTATS_STOP             = "Failed to transform 'count votes until' time from election configuration file to RFC3339 format"
	_VOTERSTATS_TIME             = "'count votes from' time > 'count votes until' time"
	_VOTERSTATS_TZ               = "Failed to load timezone"
	_VOTERSTATS_COUNTIES         = "Cannot fetch counties from a database"
	_VOTERSTATS_COUNTIES_JSON    = "Counties are invalid JSON format"
	_VOTERSTATS_COUNTIES_FOREIGN = "Counties contain 'FOREIGN' county, which is not allowed, since foreign county should have its own ID not a verbose name"
	_VOTERSTATS_COUNTIES_TOTAL   = "Counties contain 'TOTAL' county, which is unknown county"
	_VOTERSTATS_FILE             = "Failed to open a file"
	_VOTERSTATS_FILE_CLOSE       = "Failed to close a file"
	_VOTERSTATS_EXP              = "Failed to export voters statistics"
	_VOTERSTATS_JSON             = "Voters statistics are invalid JSON"
	_VOTERSTATS_EXP_INFO         = "Voters statistics export in progress"
	_VOTERSTATS_EXP_FAIL         = "Cannot fetch voters statistics from a database"
	_VOTERSTATS_OUTSIDE_TIME     = "Voter has voted after voters statistics export began"
	_VOTERSTATS_COUNTY_ADMIN     = "Admin code doesn't have corresponding county"
	_VOTERSTATS_NON_FATAL        = "Non-fatal error"
)
