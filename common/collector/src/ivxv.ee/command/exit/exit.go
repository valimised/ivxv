/*
Package exit contains exit codes for command-line application.

This is a separate package to avoid import cycles between ivxv.ee/command and
ivxv.ee/conf.
*/
package exit // import "ivxv.ee/command/exit"

// Exit code contant values and descriptions taken from
// /usr/include/sysexits.h.
const (
	OK = 0 // successful termination

	// EX_USAGE -- The command was used incorrectly, e.g., with the wrong
	//         number of arguments, a bad flag, a bad syntax in a
	//         parameter, or whatever.
	Usage = 64

	// EX_DATAERR -- The input data was incorrect in some way. This should
	//         only be used for user's data & not system files.
	DataErr = 65

	// EX_NOINPUT -- An input file (not a system file) did not exist or was
	//         not readable. This could also include errors like "No
	//         message" to a mailer (if it cared to catch it).
	NoInput = 66

	// EX_NOUSER -- The user specified did not exist. This might be used
	//         for mail addresses or remote logins.
	NoUser = 67

	// EX_NOHOST -- The host specified did not exist. This is used in mail
	//         addresses or network requests.
	NoHost = 68

	// EX_UNAVAILABLE -- A service is unavailable. This can occur if a
	//         support program or file does not exist. This can also be
	//         used as a catchall message when something you wanted to do
	//         doesn't work, but you don't know why.
	Unavailable = 69

	// EX_SOFTWARE -- An internal software error has been detected. This
	//         should be limited to non-operating system related errors as
	//         possible.
	Software = 70

	// EX_OSERR -- An operating system error has been detected. This is
	//         intended to be used for such things as "cannot fork",
	//         "cannot create pipe", or the like. It includes things like
	//         getuid returning a user that does not exist in the passwd
	//         file.
	OSErr = 71

	// EX_OSFILE -- Some system file (e.g., /etc/passwd, /etc/utmp, etc.)
	//         does not exist, cannot be opened, or has some sort of error
	//         (e.g., syntax error).
	OSFile = 72

	// EX_CANTCREAT -- A (user specified) output file cannot be created.
	CantCreate = 73

	// EX_IOERR -- An error occurred while doing I/O on some file.
	IOErr = 74

	// EX_TEMPFAIL -- Temporary failure, indicating something that is not
	//         really an error. In sendmail, this means that a mailer
	//         (e.g.) could not create a connection, and the request should
	//         be reattempted later.
	TempFail = 75

	// EX_PROTOCOL -- The remote system returned something that was "not
	//         possible" during a protocol exchange.
	Protocol = 76

	// EX_NOPERM -- You did not have sufficient permission to perform the
	//         operation. This is not intended for file system problems,
	//         which should use NOINPUT or CANTCREAT, but rather for higher
	//         level permissions.
	NoPerm = 77

	Config = 78 // configuration error
)
