package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/miekg/dns"
)

var blacklist = map[string]struct{}{}

func isBlocked(req *dns.Msg) (blocked bool) {
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

func fetchBlacklist(blacklistURL string) error {
	req, err := http.NewRequest("GET", blacklistURL, nil)
	if err != nil {
		return err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(ioutil.Discard, resp.Body)
		return fmt.Errorf("unexpected response status code: %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)

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
