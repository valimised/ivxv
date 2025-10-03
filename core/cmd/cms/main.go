package main

import (
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"tivi.io/core/crypto/cades"
	"tivi.io/core/crypto/cms"
	"tivi.io/core/crypto/util"
)

const (
	success int = iota
	argError
	appError
)

const appName = "cms"

const (
	verifyCmd = "verify"
)

var cmds []string = []string{verifyCmd}

func parseFlag() ([]string, error) {
	flags := flag.NewFlagSet(appName, flag.ContinueOnError)
	if len(os.Args) < 2 {
		fmt.Fprintf(flags.Output(), "No command specified. Valid commands are: %s\n", strings.Join(cmds, ", "))
		return nil, errors.New("no command")
	}
	command := os.Args[1]
	if command != verifyCmd {
		fmt.Fprintf(flags.Output(), "'%s' is not valid command. Valid commands are: %s\n", command, strings.Join(cmds, ", "))
		return nil, errors.New("invalid command")
	}
	flags.Usage = func() {
		fmt.Fprintf(flags.Output(), "\nUsage:\n\n")
		if command == verifyCmd {
			fmt.Fprintf(flags.Output(), "\t%s %s [-data filepath] [-certs filepath(s)] [-cades level]\n", appName, verifyCmd)
			fmt.Fprint(flags.Output(), "\nDescription:\n\n\tVerifies the signed data with the certificates provided.\n")
		}
		fmt.Fprintln(flags.Output(), "\nFlags:")
		flags.PrintDefaults()
	}
	var data, certs, cades, detachedData, nonce string
	flags.StringVar(&data, "data", "", "filepath of the signed data in DER format to be verified")
	flags.StringVar(&certs, "certs", "", "filepath(s) (separated by ':') of the certificates to be matched")
	flags.StringVar(&cades, "cades", "", "CAdES conformance level (either 'b', 't' or 'lt')")
	flags.StringVar(&detachedData, "detached", "", "filepath of the original data, if detached")
	flags.StringVar(&nonce, "nonce", "", "nonce used for generating the timestamp (CAdES-t level)")
	if err := flags.Parse(os.Args[2:]); err != nil {
		return nil, err
	}
	if data == "" {
		fmt.Fprintf(flags.Output(), "%s %s: no data filepath provided\n", appName, command)
		flags.Usage()
		return nil, errors.New("argument missing")
	}
	if certs == "" {
		fmt.Fprintf(flags.Output(), "%s %s: no certificate filepath provided\n", appName, command)
		flags.Usage()
		return nil, errors.New("argument missing")
	}
	args := []string{command, data, certs, cades, detachedData, nonce}
	return args, nil
}

func verifySignedData(cadesLevel string, data []byte, certs []*x509.Certificate, detachedData, nonce []byte) error {
	switch cadesLevel {
	case "b":
		return cades.CAdESBVerifySignedData(data, certs, detachedData)
	case "t":
		return cades.CAdESTVerifySignedData(data, certs, detachedData, nonce)
	}
	return cms.VerifySignedData(data, certs, detachedData, map[string]cms.AttrToVerify{}, map[string]cms.AttrToVerify{})
}

func appmain() (result int) {
	args, err := parseFlag()
	if len(args) != 6 {
		fmt.Fprintf(os.Stderr, "not enough arguments\n")
		return argError
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed parsing arguments: %v\n", err)
		return argError
	}
	command, data, fcts, cades, detached, nc := args[0], args[1], args[2], args[3], args[4], args[5]
	bytes, err := ioutil.ReadFile(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed reading data file at %s: %v\n", data, err)
		return appError
	}
	fcerts := strings.Split(fcts, ":")
	certs := make([]*x509.Certificate, len(fcerts))
	for i, fcert := range fcerts {
		pem, err := ioutil.ReadFile(fcert)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed reading PEM file at '%s': %v\n", fcert, err)
			return appError
		}
		cert, err := util.GetPEMCertificate(pem)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed getting PEM certificate at '%s': %v\n", fcert, err)
			return appError
		}
		certs[i] = cert
	}
	var detachedData []byte
	if detached != "" {
		detachedData, err = ioutil.ReadFile(detached)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed reading detached data file at %s: %v\n", data, err)
			return appError
		}
	}
	var nonce []byte
	if nc == "" {
		nonce = nil
	} else {
		nonce = []byte(nc)
	}
	switch command {
	case verifyCmd:
		err := verifySignedData(cades, bytes, certs, detachedData, nonce)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed verifying signed data: %v\n", err)
			return appError
		}
		fmt.Fprintf(os.Stdout, "verification successful\n")
	}
	return success
}

func main() {
	os.Exit(appmain())
}
