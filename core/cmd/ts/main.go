package main

import (
	"context"
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"tivi.io/core/crypto/timestamp"
	"tivi.io/core/crypto/util"
)

const (
	success int = iota
	argError
	appError
)

const appName = "ts"

const (
	requestCmd = "request"
	verifyCmd  = "verify"
)

var cmds []string = []string{requestCmd, verifyCmd}

func parseFlag() (string, []string, string, uint64, uint64, string, error) {
	flags := flag.NewFlagSet(appName, flag.ContinueOnError)
	if len(os.Args) < 2 {
		fmt.Fprintf(flags.Output(), "No command specified. Valid commands are: %s\n", strings.Join(cmds, ", "))
		return "", nil, "", 0, 0, "", errors.New("no command")
	}
	command := os.Args[1]
	if command != requestCmd && command != verifyCmd {
		fmt.Fprintf(flags.Output(), "'%s' is not valid command. Valid commands are: %s\n", command, strings.Join(cmds, ", "))
		return "", nil, "", 0, 0, "", errors.New("invalid command")
	}
	flags.Usage = func() {
		fmt.Fprintf(flags.Output(), "\nUsage:\n\n")
		commonFlags := "[-data filepath] [-cert filepath] [-delay seconds] [-retries number] [-nonce number] [-out filepath]"
		if command == requestCmd {
			fmt.Fprintf(flags.Output(), "\t%s %s [-tsa url] %s\n", appName, requestCmd, commonFlags)
			fmt.Fprint(flags.Output(), "\nDescription:\n\n\tSends a timestamp request to the TSA and verifies the response against the data, nonce and certificate provided.\n")
		} else if command == verifyCmd {
			fmt.Fprintf(flags.Output(), "\t%s %s [-der filepath] %s\n", appName, verifyCmd, commonFlags)
			fmt.Fprint(flags.Output(), "\nDescription:\n\n\tVerifies a DER-format timestamp response against the data, nonce and certificate provided.\n")
		}
		fmt.Fprintln(flags.Output(), "\nFlags:")
		flags.PrintDefaults()
	}
	var tsa, derTimestamp, data, certs, nonce, out string
	var delay, retries uint64
	if command == requestCmd {
		flags.StringVar(&tsa, "tsa", "", "the url of the TSA")
	} else if command == verifyCmd {
		flags.StringVar(&derTimestamp, "der", "", "the timestamp response in DER format")
	}
	flags.StringVar(&data, "data", "", "filepath of the data to be timestamped")
	flags.StringVar(&certs, "certs", "", "filepath(s) (separated by ':') of the certificate(s) to be included in both the request and response")
	flags.StringVar(&nonce, "nonce", "", "if not provided, a pseudo-random number is included in the request")
	flags.Uint64Var(&delay, "delay", 1, "maximum time that gen time and sign time can differ")
	flags.Uint64Var(&retries, "retries", 2, "number of retries in case of failure")
	flags.StringVar(&out, "out", "", "filepath of the output file (default to stdout)")
	if err := flags.Parse(os.Args[2:]); err != nil {
		return "", nil, "", 0, 0, "", err
	}
	args := make([]string, 3)
	if command == requestCmd {
		if tsa == "" {
			fmt.Fprintf(flags.Output(), "%s %s: no TSA url provided\n", appName, command)
			flags.Usage()
			return "", nil, "", 0, 0, "", errors.New("argument missing")
		}
		args[0] = tsa
	} else if command == verifyCmd {
		if derTimestamp == "" {
			fmt.Fprintf(flags.Output(), "%s %s: no DER timestamp response filepath provided\n", appName, command)
			flags.Usage()
			return "", nil, "", 0, 0, "", errors.New("argument missing")
		}
		args[0] = derTimestamp
	}
	if data == "" {
		fmt.Fprintf(flags.Output(), "%s %s: no data filepath provided\n", appName, command)
		flags.Usage()
		return "", nil, "", 0, 0, "", errors.New("argument missing")
	}
	args[1] = data
	if certs == "" {
		fmt.Fprintf(flags.Output(), "%s %s: no certificate filepath provided\n", appName, command)
		flags.Usage()
		return "", nil, "", 0, 0, "", errors.New("argument missing")
	}
	args[2] = certs
	return command, args, nonce, delay, retries, out, nil
}

func appmain() (result int) {
	command, args, n, delay, retries, out, err := parseFlag()
	if err != nil {
		return argError
	}
	data, err := ioutil.ReadFile(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed reading data file at %s: %v\n", args[1], err)
		return appError
	}
	fcerts := strings.Split(args[2], ":")
	certs := make([]*x509.Certificate, len(fcerts))
	for i, fcert := range fcerts {
		pem, err := ioutil.ReadFile(fcert)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed reading PEM file at '%s': %v\n", fcert, err)
			return appError
		}
		cert, err := util.GetPEMCertificate(pem)
		if err != nil {
			fmt.Fprintln(os.Stderr, "get PEM certificate: %w", err)
			return appError
		}
		certs[i] = cert
	}
	var nonce []byte
	if n == "" {
		nonce = nil
	} else {
		nonce = []byte(n)
	}
	if out == "" {
		out = os.Stdout.Name()
	}
	conf := timestamp.Conf{URL: args[0], Signers: certs, DelayTime: delay, Retries: retries}
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed creating timestamp client: %v\n", err)
		return appError
	}
	switch command {
	case requestCmd:
		result, err := timestamp.RequestTimestamp(context.Background(), data, nonce, conf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed requesting timestamp: %v\n", err)
			return appError
		}
		fmt.Fprintf(os.Stdout, "request successful: ")
		if out != os.Stdout.Name() {
			fmt.Fprintf(os.Stdout, "output written to '%s'\n", out)
		} else {
			fmt.Fprintln(os.Stdout)
		}
		err = ioutil.WriteFile(out, result, 0600)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed writing output of request timestamp: %v\n", err)
			return appError
		}
	case verifyCmd:
		der, err := ioutil.ReadFile(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed reading DER response file at %s: %v\n", args[0], err)
			return appError
		}
		result, err := timestamp.VerifyTimestamp(der, data, nonce, conf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed verifying timestamp: %v\n", err)
			return appError
		}
		fmt.Fprintf(os.Stdout, "verification successful: ")
		if out != os.Stdout.Name() {
			fmt.Fprintf(os.Stdout, "output written to '%s'\n", out)
		}
		err = ioutil.WriteFile(out, []byte(result.String()), 0600)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed writing output of verify timestamp: %v\n", err)
			return appError
		}
	}
	return success
}

func main() {
	os.Exit(appmain())
}
