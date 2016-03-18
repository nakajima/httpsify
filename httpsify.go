// httpsify is a transparent blazing fast https offloader with auto certificates renewal .
// this software is published under MIT License .
// by Mohammed Al ashaal <alash3al.xyz> with the help of those opensource libraries [github.com/xenolf/lego, github.com/dkumor/acmewrapper] .
package main

import (
	"io"
	"net"
	"log"
	"flag"
	"strings"
	"net/http"
	"crypto/tls"
	"path/filepath"
	"github.com/dkumor/acmewrapper"
)

// --------------

const VERSION = "httpsify/v1"

var (
	port	*string		=	flag.String("port", "443", "the port that will serve the https requests")
	cert	*string 	=	flag.String("cert", "./cert.pem", "the cert.pem save-path")
	key		*string 	=	flag.String("key", "./key.pem", "the key.pem save-path")
	domains	*string 	=	flag.String("domains", "", "a comma separated list of your site(s) domain(s)")
	backend	*string 	=	flag.String("backend", "", "the backend http server that will serve the terminated requests")
	info 	*string 	=	flag.String("info", "yes", "whether to send information about httpsify or not ^_^")
)

// --------------

func init() {
	flag.Parse()
	if ( *domains == "" ) {
		log.Fatal("err> Please enter your site(s) domain(s)")
	}
	if ( *backend == "" ) {
		log.Fatal("err> Please enter the backend http server")
	}
}

// --------------

func main() {
	acme, err := acmewrapper.New(acmewrapper.Config{
	    Domains: strings.Split(*domains, ","),
	    Address: ":" + *port,
	    TLSCertFile: *cert,
	    TLSKeyFile:  *key,
	    RegistrationFile: filepath.Dir(*cert) + "/lets-encrypt-user.reg",
	    PrivateKeyFile:   filepath.Dir(*cert) + "/lets-encrypt-user.pem",
	    TOSCallback: acmewrapper.TOSAgree,
	})
	if err!=nil {
	    log.Fatal("err> "+ err.Error())
	}
	listener, err := tls.Listen("tcp", ":" + *port, acme.TLSConfig())
	if err != nil {
		log.Fatal("err> " + err.Error())
	}
	log.Fatal(http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		defer r.Body.Close()
		req, err := http.NewRequest(r.Method, *backend, r.Body)
		if err != nil {
			http.Error(w, http.StatusText(504), 504)
			return
		}
		for k, v := range r.Header {
			for i := 0; i < len(v); i ++ {
				if i == 0 {
					req.Header.Set(k, v[i])
				} else {
					req.Header.Add(k, v[i])
				}
			}
		}
		uip, uport, _ := net.SplitHostPort(r.RemoteAddr)
		req.Host = r.Host
		req.Header.Set("Host", r.Host)
		req.Header.Set("X-Real-IP", uip)
		req.Header.Set("X-Remote-IP", uip)
		req.Header.Set("X-Remote-Port", uport)
		req.Header.Set("X-Forwarded-For", uip)
		req.Header.Set("X-Forwarded-Proto", "https")
		req.Header.Set("X-Forwarded-Host", r.Host)
		req.Header.Set("X-Forwarded-Port", *port)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, http.StatusText(504), 504)
			return
		}
		defer res.Body.Close()
		for k, v := range res.Header {
			for i := 0; i < len(v); i ++ {
				if i == 0 {
					w.Header().Set(k, v[i])
				} else {
					w.Header().Add(k, v[i])
				}
			}
		}
		if *info == "yes" {
			w.Header().Set("Server", VERSION)
		}
		w.WriteHeader(res.StatusCode)
		io.Copy(w, res.Body)
	})))
}
