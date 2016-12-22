// httpsify is a transparent blazing fast https offloader with auto certificates renewal .
// this software is published under MIT License .
// by Mohammed Al ashaal <alash3al.xyz> with the help of those opensource libraries [github.com/xenolf/lego, github.com/dkumor/acmewrapper] .
package main

import (
	"crypto/tls"
	"flag"
	"github.com/dkumor/acmewrapper"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/scottjg/go-nat"
)

// --------------

const version = "httpsify/holepuncher/v1"

var (
	port    = flag.String("port", "443", "the port that will serve the https requests")
	cert    = flag.String("cert", "./cert.pem", "the cert.pem save-path")
	key     = flag.String("key", "./key.pem", "the key.pem save-path")
	domains = flag.String("domains", "", "a comma separated list of your site(s) domain(s)")
	backend = flag.String("backend", "http://127.0.0.1:80", "the backend http server that will serve the terminated requests")
	info    = flag.String("info", "yes", "whether to send information about httpsify or not ^_^")
	natfwd  = flag.Bool("natfwd", false, "automatically setup a port forwarding rule on the upstream NAT router")
)

// --------------

func init() {
	flag.Parse()
	if *domains == "" {
		log.Fatal("err> Please enter your site(s) domain(s)")
	}
}

// --------------

func main() {
	portNum, err := strconv.Atoi(*port)
	if err != nil {
		log.Fatalf("bogus port: %s", err)
	}

	backendUrl, err := url.Parse(*backend)
	if err != nil {
		log.Fatalf("bogus backend url: %s", err)
	}

	if *natfwd {
		gw, err := nat.DiscoverGateway()
		if err != nil {
			log.Fatalf("error: %s", err)
		}

		err = gw.AddPortMapping("tcp", portNum, 443, "https", 60*time.Second)
		if err != nil {
			log.Fatalf("error: %s", err)
		}

		// unmap the port if you ctrl-C or if we finish running main().
		// in other situations, we may leak the mapping, but it will expire
		// after a minute, so it could be worse.
		defer gw.DeletePortMapping("tcp", portNum, 443)
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			for _ = range c {
				gw.DeletePortMapping("tcp", portNum, 443)
				os.Exit(1)
			}
		}()

		go func() {
			for {
				time.Sleep(30 * time.Second)

				err = gw.AddPortMapping("tcp", portNum, 443, "https", 60*time.Second)
				if err != nil {
					log.Fatalf("error: %s", err)
				}
			}
		}()
	}

	acme, err := acmewrapper.New(acmewrapper.Config{
		Domains:          strings.Split(*domains, ","),
		Address:          ":" + *port,
		TLSCertFile:      *cert,
		TLSKeyFile:       *key,
		RegistrationFile: filepath.Dir(*cert) + "/lets-encrypt-user.reg",
		PrivateKeyFile:   filepath.Dir(*cert) + "/lets-encrypt-user.pem",
		TOSCallback:      acmewrapper.TOSAgree,
	})
	if err != nil {
		log.Fatal("err> " + err.Error())
	}
	listener, err := tls.Listen("tcp", ":"+*port, acme.TLSConfig())
	if err != nil {
		log.Fatal("err> " + err.Error())
	}

	reverseProxy := httputil.NewSingleHostReverseProxy(backendUrl)
	log.Fatal(http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uip, uport, _ := net.SplitHostPort(r.RemoteAddr)

		r.Host = r.Host
		r.Header.Set("Host", r.Host)
		r.Header.Set("X-Real-IP", uip)
		r.Header.Set("X-Remote-IP", uip)
		r.Header.Set("X-Remote-Port", uport)
		r.Header.Set("X-Forwarded-For", uip)
		r.Header.Set("X-Forwarded-Proto", "https")
		r.Header.Set("X-Forwarded-Host", r.Host)
		r.Header.Set("X-Forwarded-Port", *port)
		reverseProxy.ServeHTTP(w, r)
	})))
}
