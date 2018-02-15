// httpsify is a transparent blazing fast https offloader with auto certificates renewal .
// this software is published under MIT License .
// by Mohammed Al ashaal <alash3al.xyz> with the help of those opensource libraries [github.com/xenolf/lego, github.com/dkumor/acmewrapper] .
package main

import (
	"crypto/tls"
	"flag"
	"fmt"
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

	"github.com/dkumor/acmewrapper"

	"github.com/scottjg/go-nat"
)

// --------------

const version = "httpsify/holepuncher/v2"

var (
	port       = flag.String("port", "4443", "the port that will serve the https requests")
	ddns       = flag.String("ddns", "", "specify provider (e.g. namecheap or iwantmyname) or update url")
	cert       = flag.String("cert", "./cert.pem", "the cert.pem save-path")
	key        = flag.String("key", "./key.pem", "the key.pem save-path")
	backend    = flag.String("backend", "http://127.0.0.1:80", "the backend http server that will serve the terminated requests")
	info       = flag.String("info", "yes", "whether to send information about httpsify or not ^_^")
	skipnatfwd = flag.Bool("skipnatfwd", false, "don't automatically setup a port forwarding rule on the upstream NAT router")
	natfwd     = flag.String("natfwd", "default", "comma separated list of internal:external ports to map, defaults to opening port 443 to point to your https server")
)

// --------------

var domain = ""
var portsToMap = map[int]int{}

func init() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		domain = ""
	} else {
		domain = args[0]
	}

	_, err := strconv.Atoi(*port)
	if err != nil {
		log.Fatalf("bogus port: %s", err)
	}

	if *skipnatfwd {
		return
	}

	if *natfwd == "default" {
		*natfwd = fmt.Sprintf("%s:443", *port)
	}

	for _, portRange := range strings.Split(*natfwd, ",") {
		p := strings.Split(portRange, ":")
		intPort, err := strconv.Atoi(p[0])
		if err != nil {
			log.Fatalf("bogus natfwd parameter: %v", err)
			return
		}
		extPort, err := strconv.Atoi(p[1])
		if err != nil {
			log.Fatalf("bogus natfwd parameter: %v", err)
			return
		}
		portsToMap[intPort] = extPort
	}
}

// --------------

func main() {
	backendUrl, err := url.Parse(*backend)
	if err != nil {
		log.Fatalf("bogus backend url: %s", err)
	}

	if len(portsToMap) > 0 {
		gw, err := nat.DiscoverGateway()
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		log.Printf("Detected gateway type: %v\n", gw.Type())
		for intPort, extPort := range portsToMap {
			gw.DeletePortMapping("tcp", intPort, extPort)
			err = gw.AddPortMapping("tcp", intPort, extPort, "httpsify", 60*time.Second)
			if err != nil {
				log.Fatalf("error: %s", err)
			}
			log.Printf("Mapped internal port %v to external port %v.\n", intPort, extPort)
		}

		// unmap the port if you ctrl-C or if we finish running main().
		// in other situations, we may leak the mapping, but it will expire
		// after a minute, so it could be worse.
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			for _ = range c {
				for intPort, extPort := range portsToMap {
					gw.DeletePortMapping("tcp", intPort, extPort)
				}
				os.Exit(1)
			}
		}()

		go func() {
			for {
				time.Sleep(30 * time.Second)
				for intPort, extPort := range portsToMap {
					err = gw.AddPortMapping("tcp", intPort, extPort, "httpsify", 60*time.Second)
					if err != nil {
						log.Printf("error: %s\n", err)
					}
				}
			}
		}()
	}

	if *ddns != "" {
		updateUrl := ""
		dnsUsername := os.Getenv("DNS_USERNAME")
		dnsPassword := os.Getenv("DNS_PASSWORD")

		var req *http.Request
		switch *ddns {
		case "namecheap":
			if dnsPassword == "" {
				log.Fatalf("Need to define DNS_PASSWORD in your environment")
			}

			domainLevels := strings.Split(domain, ".")
			host := "@"
			sld := domain
			if len(domainLevels) > 2 {
				host = strings.Join(domainLevels[0:(len(domainLevels)-2)], ".")
				sld = strings.Join(domainLevels[(len(domainLevels)-2):len(domainLevels)], ".")
			}
			updateUrl = "https://dynamicdns.park-your-domain.com/update?host=" + host + "&domain=" + sld + "&password=" + dnsPassword
			req, err = http.NewRequest("GET", updateUrl, nil)
			break
		case "iwantmyname":
			if dnsPassword == "" || dnsUsername == "" {
				log.Fatalf("Need to define DNS_USERNAME and DNS_PASSWORD in your environment")
			}
			updateUrl = "https://iwantmyname.com/basicauth/ddns?hostname=" + domain
			req, err = http.NewRequest("GET", updateUrl, nil)
			if req != nil {
				req.SetBasicAuth(dnsUsername, dnsPassword)
			}
			break

		default:
			if dnsPassword == "" || dnsUsername == "" {
				log.Print("Warning: ddns url specified without username/password")
			}

			req, err = http.NewRequest("GET", updateUrl, nil)
			if req != nil && dnsPassword != "" && dnsUsername != "" {
				req.SetBasicAuth(dnsUsername, dnsPassword)
			}
		}

		client := &http.Client{}

		go func() {
			for {
				resp, err := client.Do(req)
				if err != nil {
					log.Fatalf("ddns error: %v\n", err)
				}

				if resp.StatusCode != 200 {
					if resp.StatusCode == 429 {
						log.Println("ddns rate limited. trying again in ten minutes")
						time.Sleep(10 * time.Minute)
					} else {
						log.Fatalf("ddns error: got http status code %v from api\n", resp.StatusCode)
					}
				}

				time.Sleep(60 * time.Second)
			}
		}()
	}

	if domain != "" {
		acme, err := acmewrapper.New(acmewrapper.Config{
			Domains:          []string{domain},
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
	select {}
}
