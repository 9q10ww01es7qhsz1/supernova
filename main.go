package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/miekg/dns"
)

func main() {
	var (
		addr              string
		blacklistFilename string
	)

	flag.StringVar(&addr, "addr", ":53", "addr")
	flag.StringVar(&upstream, "upstream", "1.1.1.1:53", "upstream")
	flag.StringVar(&blacklistFilename, "blacklist filename", "black.list", "blacklist filename")
	flag.Parse()

	if err := fetchBlacklist(blacklistFilename); err != nil {
		log.Fatalln(fmt.Errorf("failed to fetch blacklist: %w", err))
	}

	dns.HandleFunc(".", handler)
	log.Fatalln(dns.ListenAndServe(addr, "udp", nil))
}
