package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"ivxv.ee/command/exit"
	"ivxv.ee/conf"
	"ivxv.ee/errors"
	"ivxv.ee/log"
	//ivxv:modules container
)

func main() {
	code, err := verifierMain()
	if err != nil {
		// nolint: gas, if writing to stderr returns an error, then we
		// do not have any way to report it anyway.
		fmt.Fprintln(os.Stderr, "error:", err)
	}
	os.Exit(code)
}

func verifierMain() (int, error) {
	flag.Usage = func() {
		// nolint: gas, if writing to stderr returns an error, then we
		// do not have any way to report it anyway.
		fmt.Fprintln(os.Stderr, "Usage: "+os.Args[0]+` [options] <container file>

verifier uses the trust root given in order to verify a container's signatures
and return the signers and signing times.

The trust container must contain a single file. This is the trust configuration
which must be identified by the key "trust.yaml".

The container to be verified must have an extension corresponding to the
container type it is, e.g., foo.bdoc.

options:`)
		flag.PrintDefaults()
	}

	trust := flag.String("trust", "/etc/ivxv/trust.bdoc",
		"`path` to the trust container. Must have an extension\n"+
			"    \tcorresponding to the container type it is, e.g., trust.bdoc.\n"+
			"    \t")
	flag.Parse()

	if len(flag.Args()) != 1 {
		flag.Usage()
		return exit.Usage, nil
	}
	path := flag.Arg(0)

	// We do not want the verifier application to log anything, but it is
	// still assumed that the context has a logger. Use TestContext which
	// provides a test logger that does nothing.
	ctx := log.TestContext(context.Background())

	cfg, code, err := conf.New(ctx, *trust, "", "")
	if err != nil {
		return code, fmt.Errorf("failed to load trust root: %v", err)
	}

	c, err := cfg.Container.OpenFile(path)
	if err != nil {
		code = exit.DataErr
		if perr := errors.CausedBy(err, new(os.PathError)); perr != nil {
			if os.IsNotExist(perr) {
				code = exit.NoInput
			}
		}
		return code, fmt.Errorf("failed to open container: %v", err)
	}
	defer c.Close() // nolint: errcheck, ignore close failure of read-only container.

	for _, s := range c.Signatures() {
		fmt.Println(s.Signer.Subject.CommonName, s.SigningTime.Format(time.RFC3339))
	}

	return exit.OK, nil
}
