package main

import (
	"fmt"
	"log"
	"net"

	"github.com/miekg/dns"
)

var (
	client   dns.Client
	upstream string

	blockIPv4 = net.ParseIP("0.0.0.0")
	blockIPv6 = net.ParseIP("0:0:0:0:0:0:0:0")
	blockTTL  = uint32(60)
)

func handler(rw dns.ResponseWriter, req *dns.Msg) {
	defer rw.Close()

	log.Println("req", req.Question[0].Name)

	if isBlocked(req) {
		if err := block(rw, req); err != nil {
			log.Println(fmt.Errorf("failed to block request: %w", err))
		}

		return
	}

	resp, _, err := client.Exchange(req, upstream)
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to exchange: %w", err))
		return
	}

	// log.Println("resp", resp)

	if err = rw.WriteMsg(resp); err != nil {
		log.Println(fmt.Errorf("failed to reply: %w", err))
	}
}

func block(rw dns.ResponseWriter, req *dns.Msg) error {
	resp := &dns.Msg{}
	resp.SetReply(req)

	q := req.Question[0]

	header := dns.RR_Header{
		Name:   q.Name,
		Rrtype: q.Qtype,
		Class:  q.Qclass,
		Ttl:    blockTTL,
	}

	var answer dns.RR

	switch q.Qtype {
	case dns.TypeA:
		answer = &dns.A{
			Hdr: header,
			A:   blockIPv4,
		}
	case dns.TypeAAAA:
		answer = &dns.AAAA{
			Hdr:  header,
			AAAA: blockIPv6,
		}
	}

	resp.Answer = append(resp.Answer, answer)

	return rw.WriteMsg(resp)
}
