/*
The xroad-service communicates with X-road security server and IVXV system.
*/
package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"net/http"
	"os"

	"xroad/pkg/conf"
	"xroad/pkg/server"
	"xroad/pkg/service"

	log "github.com/sirupsen/logrus"
)

const appName = "xroad-service"

const (
	success int = iota
	argError
	confError
	serverShutdownError
)

func xroadmain() int {
	flags := flag.NewFlagSet(appName, flag.ContinueOnError)
	log.SetFormatter(&log.JSONFormatter{})
	confLoc := flags.String("conf", "/etc/xroad-service/xroad-service.json", "configuration file")
	if err := flags.Parse(os.Args[1:]); err != nil {
		log.Error("Parsing args", err)
		return argError
	}
	conf, err := conf.LoadJSON(*confLoc)
	if err != nil {
		log.Error("Loading conf: ", err)
		return confError
	}
	s := service.NewEHSService(conf.Elections, conf.Server.BatchMaxSize)
	handler := server.CreateHandler(s, conf.Server.OpenApiPath)

	cert, err := conf.Server.TLS.X509Certificate()
	if err != nil {
		log.Error("Loading certs: ", err)
		return confError
	}
	xcert, err := os.ReadFile(conf.Server.Xroad.Certificate)
	if err != nil {
		log.Error("Loading xroad cert: ", err)
		return confError
	}
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(xcert); !ok {
		log.Error("unable to parse cert from %s", cert)
		return confError
	}

	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{*cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}
	srv := http.Server{
		Addr:      conf.Server.Address,
		Handler:   handler,
		TLSConfig: tlsConf,
	}

	err = srv.ListenAndServeTLS("", "")
	if err != http.ErrServerClosed {
		log.Error("Serving ended:", err)
		return serverShutdownError
	}
	return success
}

func main() {
	os.Exit(xroadmain())

}
