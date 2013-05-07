package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"net"
	"net/http"
	"net/http/cgi"
	"net/http/fcgi"
)

var cmd = flag.String("c", "", "CGI program to run")
var pwd = flag.String("w", "", "Working dir for CGI")
var serveFcgi = flag.Bool("f", false, "Run as a FCGI 'server' instead of HTTP")
var debug = flag.Bool("debug", false, "Print debug msgs to stderr.")
var address = flag.String("a", ":3333", "Listen address")

func main() {
	flag.Usage = usage
	flag.Parse()

	if *cmd == "" {
		usage()
	}

	// This is a hack to make p9p's rc happier for some unknown reason.
	c := *cmd
	if c[0] != '/' {
		c = "./" + c
	}

	os.Setenv("PATH", os.Getenv("PATH")+":.")

	h := &cgi.Handler{
		Path:       c,
		Root:       "/",
		Dir:        *pwd,
		InheritEnv: []string{"PATH", "PLAN9"},
	}

	var err error
	if *serveFcgi {
		socket := strings.Split(*address, ":")
		if socket[0] == "unix" {
			os.Remove(socket[1])
		} else {
			socket[0] = "tcp"
			socket[1] = *address
		}

		if l, err := net.Listen(socket[0], socket[1]); err == nil {
			log.Println("Starting FastCGI daemon listening on", *address)
			err = fcgi.Serve(l, h)
		}

	} else {
		log.Println("Starting HTTP server listening on", *address)
		err = http.ListenAndServe(*address, h)
	}

	if err != nil {
		log.Fatal(err)
	}
}

func usage() {
	os.Stderr.WriteString("usage: cgd [-f] -c prog [-w wdir] [-a :PORT|unix:SOCKET]\n")
	flag.PrintDefaults()
	os.Exit(2)
}
