/*
Package conf provides common code for accessing the configuration of collector
services.
*/
package conf // import "ivxv.ee/conf"

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"ivxv.ee/age"
	"ivxv.ee/auth"
	"ivxv.ee/command/exit"
	"ivxv.ee/conf/version"
	"ivxv.ee/container"
	"ivxv.ee/dds"
	"ivxv.ee/errors"
	"ivxv.ee/identity"
	"ivxv.ee/log"
	"ivxv.ee/mid"
	"ivxv.ee/q11n"
	"ivxv.ee/server"
	"ivxv.ee/storage"
	"ivxv.ee/yaml"
)

// C wraps the configuration container parser, election and technical
// configuration into one.
type C struct {
	Version   version.V
	Container container.Opener
	Election  *Election
	Technical *Technical
}

// Election contains the election parameters.
type Election struct {
	Identifier string   // The election identifier.
	Questions  []string // Identifiers of questions asked during this election.

	// Election period. All times must be formatted according to RFC3339.
	Period struct {
		ServiceStart  string // The time when to start serving requests.
		ElectionStart string // The start time of the election: votes before this are not counted.
		ElectionStop  string // The end time of the election: choices are no longer distributed.
		ServiceStop   string // The time when to stop serving requests.
	}

	Voting struct {
		RateLimitStart   uint64 // After how many votes do we rate limit submissions? 0 means immediately.
		RateLimitMinutes uint64 // How many minutes between submission when limiting? 0 disables rate limiting.
	}

	Verification struct {
		Count      uint64 // How many times a vote can be verified? 0 means unlimited.
		Minutes    uint64 // How much time is given to verify a vote? 0 means unlimited.
		LatestOnly bool   // If true, then only the latest vote of a voter can be verified.
	}

	// VoterForeignEHAK specifies the administrative unit code (EHAK) to
	// use for determining voter districts if the voter is foreign. If
	// VoterForeignEHAK is empty, then the default value "0000" is used.
	VoterForeignEHAK string

	// IgnoreVoterList is an option used for public testing and should NOT
	// be set during production! If not empty, then voter lists will be
	// ignored, meaning that everybody who can authenticate and produce a
	// valid signed container can vote, and the specified choices list will
	// be presented to all users.
	IgnoreVoterList string

	VoterList struct {
		Key string // PEM-encoding of the public key used to verify voter list signatures.
	}

	// Composited configuration structures defined in other packages.
	Auth          auth.Conf
	Identity      identity.Type
	Age           age.Conf
	Vote          container.Conf
	DDS           dds.Conf
	MID           mid.Conf
	Qualification q11n.Conf
}

// ServiceStartTime parses c.Period.ServiceStart and returns the result.
func (e Election) ServiceStartTime() (time.Time, error) {
	return time.Parse(time.RFC3339, e.Period.ServiceStart)
}

// ElectionStartTime parses c.Period.ElectionStart and returns the result.
func (e Election) ElectionStartTime() (time.Time, error) {
	return time.Parse(time.RFC3339, e.Period.ElectionStart)
}

// ElectionStopTime parses c.Period.ElectionStop and returns the result.
func (e Election) ElectionStopTime() (time.Time, error) {
	return time.Parse(time.RFC3339, e.Period.ElectionStop)
}

// ServiceStopTime parses c.Period.ElectionStop and returns the result.
func (e Election) ServiceStopTime() (time.Time, error) {
	return time.Parse(time.RFC3339, e.Period.ServiceStop)
}

// VoterForeignEHAKDefault returns e.VoterForeignEHAK or its default value if
// it is not specified.
func (e Election) VoterForeignEHAKDefault() string {
	if e.VoterForeignEHAK != "" {
		return e.VoterForeignEHAK
	}
	return "0000"
}

// Technical contains the collector services' technical parameters.
type Technical struct {
	Debug bool // Should debug logging be enabled?

	Network []struct {
		ID       string   // Network segment identifier.
		Services Services // Configured services in this segment.
	}

	// Composited configuration structures defined in other packages.
	Filter  server.FilterConf
	Storage storage.Conf
}

// Services is a block of configured services for each service type.
type Services struct {
	Proxy        []*Service
	DDS          []*Service
	MID          []*Service
	Choices      []*Service
	Voting       []*Service
	Verification []*Service
	Storage      []*Service
}

// Services finds the configured services for the requested network segment.
func (t Technical) Services(network string) *Services {
	for _, n := range t.Network {
		if n.ID == network {
			return &n.Services
		}
	}
	return nil
}

// Service is the configuration for a single service instance.
type Service struct {
	ID          string // Unique identifier of the service instance.
	Address     string // The host:port to listen on for requests.
	PeerAddress string // The host:port to listen on for peer messages.
}

// Service finds the network and configuration for a service instance with id.
func (t Technical) Service(id string) (network string, service *Service) {
	fields := reflect.TypeOf(new(Services)).Elem().NumField()
	for _, n := range t.Network {
		services := reflect.ValueOf(n.Services)
		for i := 0; i < fields; i++ {
			for _, s := range services.Field(i).Interface().([]*Service) {
				if s.ID == id {
					return n.ID, s
				}
			}
		}
	}
	return "", nil
}

// Sensitive returns the path to the service directory for the service instance
// with ID. This directory contains sensitive service information which can not
// be included in the configuration, e.g., secret keys.
func Sensitive(id string) string {
	return filepath.Join("/var/lib/ivxv/service", id)
}

// TLS returns the path to the TLS-certificate and corresponding private key
// given the service directory.
func TLS(sensitive string) (cert, key string) {
	return filepath.Join(sensitive, "tls.pem"),
		filepath.Join(sensitive, "tls.key")
}

// New opens and parses configuration files. All files must be signature
// containers and the names must have extensions corresponding to the type of
// container that they are. If err is not nil, then code is the code that the
// caller should exit with.
//
// New first opens trust without verification and parses the file named
// "trust.yaml" in the container to obtain verification parameters to construct
// a configuration container parser. It uses this to verify the signature
// provided on the trust container itself to check for errors in the
// self-signed trust container.
//
// Next, election and technical are verified and opened and the files
// "election.yaml" and "technical.yaml" are parsed for the election and
// technical configuration, respectively.
//
// The configuration container parser, election and technical configurations
// are returned nested into C.
//
// If election or technical are empty strings, then parsing the election or
// technical configuration is skipped, respectively. Trust must not be empty,
// because it is needed for verifying other configurations.
func New(ctx context.Context, trust, election, technical string) (c *C, code int, err error) {
	log.Log(ctx, Parsing{Trust: trust, Election: election, Technical: technical})

	c = new(C)
	if code, err = c.trust(ctx, trust); err != nil {
		return nil, code, ParseTrustError{Path: trust, Err: err}
	}

	if len(election) > 0 {
		c.Election = new(Election)
		if c.Version.Election, code, err = c.parse(
			ctx, election, "election.yaml", &c.Election); err != nil {

			return nil, code, ParseElectionError{Path: election, Err: err}
		}
	}

	if len(technical) > 0 {
		c.Technical = new(Technical)
		if c.Version.Technical, code, err = c.parse(
			ctx, technical, "technical.yaml", &c.Technical); err != nil {

			return nil, code, ParseTechnicalError{Path: technical, Err: err}
		}
		log.SetDebug(ctx, c.Technical.Debug)
	}
	return
}

func (c *C) trust(ctx context.Context, path string) (code int, err error) {
	// First open trust without verifying signatures.
	fp, err := os.Open(path)
	if err != nil {
		return exit.NoInput, OpenTrustFileError{Err: err}
	}
	defer fp.Close()

	// Check the extension after we have ensured that the file even exists.
	t := container.Type(strings.TrimPrefix(filepath.Ext(path), "."))
	if len(t) == 0 {
		return exit.DataErr, TrustMissingExtensionError{}
	}

	cnt, err := container.UnverifiedOpen(t, fp)
	if err != nil {
		return exit.DataErr, OpenTrustError{Err: err}
	}
	defer cnt.Close()

	var conf struct {
		Container container.Conf
	}
	if err = unmarshal("trust.yaml", cnt.Data(), &conf); err != nil {
		return exit.DataErr, UnmarshalTrustError{Err: err}
	}

	// Then configure the configuration container parser and verify the
	// trust container.
	if c.Container, err = container.Configure(conf.Container); err != nil {
		return exit.DataErr, ConfigureContainerOpenerError{Err: err}
	}
	if _, err = fp.Seek(0, io.SeekStart); err != nil {
		return exit.IOErr, RewindTrustError{Err: err}
	}

	cnt, err = c.Container.Open(t, fp)
	if err != nil {
		return exit.DataErr, VerifyTrustError{Err: err}
	}
	defer cnt.Close()

	c.Version.Trust = version.Signatures(cnt.Signatures())
	if len(c.Version.Trust) == 0 {
		return exit.DataErr, UnsignedTrustError{}
	}
	for _, s := range c.Version.Trust {
		log.Log(ctx, TrustSignature{Signer: s.Signer, SigningTime: s.SigningTime})
	}
	return
}

func (c *C) parse(ctx context.Context, path, key string, v interface{}) (
	signatures version.Signatures, code int, err error) {

	cnt, err := c.Container.OpenFile(path)
	if err != nil {
		code = exit.DataErr
		if perr := errors.CausedBy(err, new(os.PathError)); perr != nil {
			if os.IsNotExist(perr) {
				code = exit.NoInput
			}
		}
		return nil, code, ConfigurationContainerError{Err: err}
	}
	defer cnt.Close()

	signatures = version.Signatures(cnt.Signatures())
	if len(signatures) == 0 {
		return nil, exit.DataErr, UnsignedConfigurationError{}
	}
	for _, s := range signatures {
		log.Log(ctx, ConfigurationSignature{
			Path:        path,
			Signer:      s.Signer,
			SigningTime: s.SigningTime,
		})
	}

	if err = unmarshal(key, cnt.Data(), v); err != nil {
		return nil, exit.DataErr, UnmarshalConfigurationError{Err: err}
	}
	return
}

// unmarshal searches a container for the an entry with the given key or that
// ends in ".<key>" and unmarshals its contents within into v.
func unmarshal(key string, data map[string][]byte, v interface{}) (err error) {
	for name, content := range data {
		if name == key || strings.HasSuffix(name, "."+key) {
			if err = yaml.Unmarshal(bytes.NewReader(content), data, v); err != nil {
				return UnmarhsalConfError{Err: err}
			}
			return nil
		}
	}
	return MissingContainerKeyError{Key: key}
}
