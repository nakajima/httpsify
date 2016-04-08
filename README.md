# Intro
A transparent HTTPS proxy with automatic certificate renewal
using https://letsencrypt.org/

# How it works ?
httpsify is a https reverse proxy ...
[https request] --> httpsify --> [apache/nginx/nodejs/... etc]
but this isn't the point because there are many https offloaders,
but httpsify uses letsencrypt (https://letsencrypt.org/)
for automatically generating free and valid ssl certificates, as well as auto renewal of certs,
this web server by default uses HTTP/2 .
you can say that httpsify is just a http/2 & letsencrypt wrapper for any http web server with no hassle, it just works .

# Features
* SSL Offloader.
* HTTP/2 support.
* Multi-Core support.
* Auto-Renewal for generated certificates.
* Blazing fast.
* Very light.
* Portable and small `~ 2 MB`
* No system requirements.
* No configurations required, just `httpsify --domains="domain.com,www.domain.com,sub.domain.com"`
* Passes `X-Forwarded-*` headers, `X-Real-IP` header and `X-Remote-IP`/`X-Remote-Port` to the backend server.

# Installation
> Currently the only available binaries are built for `linux` `386/amd64` and you can download them from [here](https://github.com/alash3al/httpsify/releases) .  

# Building from source :
* Make sure you have `Golang` installed .
* `go get github.com/alash3al/httpsify`.
* `go install github.com/alash3al/httpsify`.
*  make sure that `$GOPATH/bin` in your `$PATH` .

# Quick Usage
> lets say that you have extracted/built httpsify in the current working directory .
```bash
# this is the simplest way to run httpsify
# this will run a httpsify instance listening on port 443 and passing the incoming requests to http://localhost
# and building valid signed cerificates for the specified domains [they must be valid domain names]
./httpsify --domains="domain.tld,www.domain.tld,another.domain.tld"
```

# Author
I'm [Mohammed Al Ashaal](https://www.alash3al.xyz)

# Thanks
I must thank the following awesome libraries

* [github.com/xenolf/lego](https://github.com/xenolf/lego)
* [github.com/dkumor/acmewrapper](https://github.com/dkumor/acmewrapper)
