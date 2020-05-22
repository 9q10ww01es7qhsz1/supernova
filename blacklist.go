package main

import (
	"strings"

	"github.com/9q10ww01es7qhsz1/supernova/blacklist"
	"github.com/miekg/dns"
)

var privBlacklist = blacklist.New(httpClient)

func isBlacklisted(req *dns.Msg) bool {
	if req.Opcode != dns.OpcodeQuery {
		return false
	}

	if len(req.Question) != 1 {
		return false
	}

	q := req.Question[0]

	switch q.Qtype {
	case dns.TypeA:
	case dns.TypeAAAA:
	default:
		return false
	}

	return privBlacklist.Contains(strings.TrimRight(q.Name, "."))
}
