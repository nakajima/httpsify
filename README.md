# Intro
a transparent HTTPS termination proxy using letsencrypt with auto certification renewal, you may need to read more about LetsEncrypt from [here](https://letsencrypt.org/)

# Features
* SSL Offloader .
* HTTP/2 support .
* Multi-Core support .
* Auto-Renewal for generated certificates .
* Blazing fast .
* Very light .
* Portable and small `~ 2 MB`
* No system requirements .
* No configurations required, just `httpsify --backend=http://127.0.0.1`
* Passes `X-Forwarded-*` headers, `X-Real-IP` header and `X-Remote-IP`/`X-Remote-Port` to the backend server .

# Installation
> Currently the only available binaries are built for `linux` `386/amd64` and you can download them from [here](https://github.com/alash3al/httpsify/releases) .  
> Building from source :  
--  MAke sure you have `Golang` installed .  
--  `go get github.com/alash3al/httpsify`.  
--  `go install github.com/alash3al/httpsify`.  
--  make sure that `$GOPATH/bin` in your `$PATH` .

# Quick Usage
> lets say that you have extracted/built httpsify in the current working directory .  
```bash
# this is the simplest way to run httpsify
# this will run a httpsify instance listening on port 443 and passing the incoming requests to http://localhost
# and building valid signed cerificates for the specified domains [they must be valid domain names]
./httpsify --backend=http://localhost --domains="domain.tld,www.domain.tld,another.domain.tld"
# this command will tell you more ...
./httpsify --help
```

# Author
I'm [Mohammed Al Ashaal](https://www.alash3al.xyz)
