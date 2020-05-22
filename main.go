package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/miekg/dns"
)

var (
	httpClient = &http.Client{Timeout: 30 * time.Second}
)

func main() {
	var (
		addr         string
		blacklistURL string
	)

	defaultBlacklistURL := "https://v.firebog.net/hosts/AdguardDNS.txt"

	flag.StringVar(&addr, "addr", ":5353", "addr")
	flag.StringVar(&upstream, "upstream", "1.1.1.1:53", "upstream")
	flag.StringVar(&blacklistURL, "blacklist", defaultBlacklistURL, "blacklist URL")
	flag.Parse()

	if err := fetchBlacklist(blacklistURL); err != nil {
		log.Fatalln(fmt.Errorf("failed to fetch blacklist: %w", err))
	}

	dns.HandleFunc(".", handler)
	log.Fatalln(dns.ListenAndServe(addr, "udp", nil))
}
