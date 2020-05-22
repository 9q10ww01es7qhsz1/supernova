package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/miekg/dns"
)

var blacklist = map[string]struct{}{}

func isBlacklisted(req *dns.Msg) (blocked bool) {
	if req.Opcode != dns.OpcodeQuery {
		return
	}

	if len(req.Question) != 1 {
		return
	}

	q := req.Question[0]

	switch q.Qtype {
	case dns.TypeA:
	case dns.TypeAAAA:
	default:
		return
	}

	_, blocked = blacklist[strings.TrimRight(q.Name, ".")]

	if blocked {
		log.Println("blocked", q.Name)
	}

	return
}

func fetchBlacklist(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open blacklist: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		host := scanner.Text()

		if host == "" || strings.HasPrefix(host, "#") {
			continue
		}

		blacklist[host] = struct{}{}
	}

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan response body: %w", err)
	}

	log.Println("blacklisted", len(blacklist), "hosts")

	return nil
}
