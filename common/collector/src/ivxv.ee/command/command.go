/*
Package command provides common code for command-line applications.
*/
package command // import "ivxv.ee/command"
import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"ivxv.ee/command/exit"
	"ivxv.ee/conf"
	"ivxv.ee/log"
	"ivxv.ee/storage"
)

func mkflag(name, value, desc string) *string {
	return flag.String(name, fmt.Sprint("/etc/ivxv/", value),
		fmt.Sprintf("`path` to the %s container. Must have an extension\n"+
			"corresponding to the container type it is, e.g., %s.\n",
			desc, value)) // New line for printing the default value.
}

func mkusage(usage string, args ...string) func() {
	flag.CommandLine.VisitAll(func(f *flag.Flag) {
		f.Usage = IndentUsage(f.Usage)
	})

	var argbuf bytes.Buffer
	for _, arg := range args {
		argbuf.WriteString(" <")
		argbuf.WriteString(arg)
		argbuf.WriteString(">")
	}
	argstr := argbuf.String()

	if len(usage) > 0 {
		usage += "\n\n"
	}

	return func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options]%s\n\n%soptions:\n",
			os.Args[0], argstr, usage)
		flag.PrintDefaults()
		os.Exit(exit.Usage)
	}
}

// Enumeration of C.Until values which indicate how far the command-line
// application should execute.
const (
	CheckTrust   = iota // Check the trust root: done by the command package.
	CheckConf           // Check configuration values.
	CheckVersion        // Check current loaded version.
	CheckInput          // Check input values.
	Execute             // Execute normally.
)

// C is used by a command-line application to perform common tasks.
type C struct {
	Ctx     context.Context
	Args    []string
	Conf    *conf.C
	Network string        // Network segment for this service instance.
	Service *conf.Service // Configuration for this service instance.
	Storage *storage.Client

	// Until indicates how far the command-line application should execute.
	// The enumerated constants are ordered, so e.g., if Until is
	// CheckInput, then that means that it must perform CheckTrust,
	// CheckConf and CheckInput, but not Execute.
	Until int

	cancel func()
}

// New performs common command-line application startup tasks.
//
// 1. It parses command line options and arguments, checking that all mandatory
//    arguments are given. It uses usage and argument names from args for usage
//    and error messages. Note that arguments are only mandatory if c.Until is
//    at least CheckInput and c.Args is nil otherwise.
//
// 2. It creates a context with a logger using the provided tag. It is the
//    caller's responsibility to call log.Close with this context after the
//    command has finished.
//
// 3. The context is set up to be cancelled on an interrupt or termination
//    signal.
//
// 4. It loads the collector configuration. Options for specifying
//    configuration file locations are automatically added. Callers must
//    examine c.Until to determine if they should execute normally or only
//    check configuration and input values and exit early. Also note that if
//    c.Until is CheckConf, then either c.Conf.Election or c.Conf.Technical is
//    nil.
//
// 5. It creates a storage service client: almost all command-line applications
//    will be performing some operations on the election data (see
//    NewWithoutStorage to skip this step).
//
// If an error occurs, then New will print (and log if possible) the error and
// call os.Exit. Because of this, make sure not to call New after defers have
// been set up, that must fire before exiting.
func New(tag string, usage string, args ...string) (c *C) {
	return newC(tag, usage, true, args...)
}

// NewWithoutStorage does the same startup tasks as New, but skips creating a
// storage service client. This method is only needed for services which do not
// communicate with storage directly.
func NewWithoutStorage(tag string, usage string, args ...string) (c *C) {
	return newC(tag, usage, false, args...)
}

func newC(tag string, usage string, withStorage bool, args ...string) (c *C) {
	c = new(C)

	trustp := mkflag("trust", "trust.bdoc", "configuration trust")
	elecp := mkflag("election", "election.bdoc", "election configuration")
	techp := mkflag("technical", "technical.bdoc", "technical configuration")
	instancep := flag.String("instance", "",
		"instance identifier for the service running this command.")
	checkp := flag.String("check", "", "check `target` configuration, "+
		"version, and/or input file and exit. target\nmust be either "+
		`"trust", "technical", "election", "version", or "input".`)

	flag.Usage = mkusage(usage, args...)
	flag.Parse()

	switch *checkp {
	case "trust":
		c.Until = CheckTrust
		*techp = "" // Skip loading other configuration files.
		*elecp = ""
	case "technical":
		c.Until = CheckConf
		*elecp = "" // Skip loading election configuration.
	case "election":
		c.Until = CheckConf
		*techp = "" // Skip loading technical configuration.
	case "version":
		c.Until = CheckVersion
	case "input":
		c.Until = CheckInput
	case "":
		c.Until = Execute
	default:
		fmt.Fprintln(os.Stderr, `error: bad "check" value:`, *checkp)
		flag.Usage()
	}

	switch m, n := flag.NArg(), len(args); {
	case m < n:
		// Only require arguments if at least checking input. Do not
		// disallow them to maintain backwards-compatibility with when
		// arguments were always mandatory.
		if c.Until >= CheckInput {
			fmt.Fprintln(os.Stderr, "error: missing argument:", args[m])
			flag.Usage()
		}
	case m > n:
		fmt.Fprintln(os.Stderr, "error: extra argument:", flag.Arg(n))
		flag.Usage()
	default:
		c.Args = flag.Args()
	}

	var err error
	if c.Ctx, err = log.NewContext(context.Background(), tag); err != nil {
		os.Exit(c.Error(exit.Unavailable, nil, "failed to configure logger:", err))
	}
	log.Log(c.Ctx, Started{Args: os.Args})

	c.Ctx, c.cancel = context.WithCancel(c.Ctx)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-ch
		log.Log(c.Ctx, Cancelled{Signal: s})
		c.cancel()
	}()

	var code int
	if c.Conf, code, err = conf.New(c.Ctx, *trustp, *elecp, *techp); err != nil {
		os.Exit(c.Error(code, ConfigurationError{Err: err},
			"failed to load configuration:", err))
	}

	// If only checking trust, then already exit here.
	if c.Until == CheckTrust {
		os.Exit(c.Cleanup(exit.OK))
	}

	// If we have the technical configuration, then find our network
	// segment and instance configuration.
	if c.Conf.Technical != nil {
		c.Network, c.Service = c.Conf.Technical.Service(*instancep)
		if c.Service == nil {
			os.Exit(c.Error(exit.Config, UnknownInstanceIDError{ID: *instancep},
				"no such service ID:", *instancep))
		}
	}

	if withStorage && c.Conf.Technical != nil {
		var servers []string
		for _, s := range c.Conf.Technical.Services(c.Network).Storage {
			servers = append(servers, s.Address)
		}
		if c.Storage, err = storage.New(&c.Conf.Technical.Storage,
			&storage.Services{
				Sensitive: conf.Sensitive(c.Service.ID),
				Servers:   servers,
			}); err != nil {

			os.Exit(c.Error(exit.Config, StorageConfigurationError{Err: err},
				"failed to configure storage client:", err))
		}
	}
	return
}

// Cleanup cancels the context to clean up resources and closes the logger.
// code is the exit code that the caller is exiting with: if it is OK, then it
// may be replaced by Cleanup if an error occurs, otherwise it is returned
// unmodified. Any errors will also be printed.
//
// Using this design, Cleanup can be deferred as such:
//
//	func appmain() (code int) {
//		c := command.New(...)
//		defer func() {
//			code = c.Cleanup(code)
//		}()
//		return
//	}
//
func (c *C) Cleanup(code int) int {
	c.cancel()
	log.Log(c.Ctx, Done{})
	if err := log.Close(c.Ctx); err != nil {
		ccode := c.Error(exit.Unavailable, nil, "failed to close logger:", err)
		if code == exit.OK {
			code = ccode
		}
	}
	return code
}

// Error logs err if not nil, prints an error message containing a, and returns
// the provided code. Meant to easily exit after an error in the main function
// of a command-line application.
func (c *C) Error(code int, err log.ErrorEntry, a ...interface{}) int {
	if err != nil {
		log.Error(c.Ctx, err)
	}

	fmt.Fprintln(os.Stderr, append([]interface{}{"error:"}, a...)...)
	return code
}

// IndentUsage indents a flag's usage message after each newline such that it
// lines up with previous lines when printed with flag.PrintDefaults.
func IndentUsage(usage string) string {
	return indentUsage(usage)
}
