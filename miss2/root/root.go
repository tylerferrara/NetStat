package main

import (
	"fmt"
	"net"
	"netsec/dnsutils"
	"os"
	"strconv"
	"time"

	"github.com/miekg/dns"
)

// Network info
const staticIP = "127.0.0.2"
const staticPort = 8020

// Flag
var verbose bool

// Zone location
const zoneFile = "zones/root.zone"

func printDate() {
	fmt.Printf("--------- %s ---------\n", time.Now().Format("2006-01-02 15:04:05.000000"))
}

func answerQuestion(q dns.Question, resp *dns.Msg) {
	// look at zone data for domain (name)
	found := dnsutils.GetZones(q)
	if len(found) == 0 {
		if verbose {
			printDate()
			fmt.Printf("No record found for Question: %s\n", q.String())
		}
		return
	}
	for _, entry := range found {
		if verbose {
			fmt.Println("Entry found:")
			fmt.Println(entry)
		}

		switch q.Qtype {
		case dns.TypeA:
			rec := new(dns.A)
			rec.Hdr = dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    0,
			}

			if lst := dnsutils.GetResolutionList(entry); len(lst) > 0 {
				rec.A = net.ParseIP(lst[0])
			}
			resp.Answer = append(resp.Answer, rec)
		case dns.TypeNS:
			rec := new(dns.NS)
			rec.Hdr = dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypeNS,
				Class:  dns.ClassINET,
				Ttl:    0,
			}
			if lst := dnsutils.GetResolutionList(entry); len(lst) > 0 {
				rec.Ns = lst[0]
			}
			resp.Answer = append(resp.Answer, rec)
		case dns.TypeSOA:
			rec := new(dns.SOA)
			rec.Hdr = dns.RR_Header{
				Name:   q.Name,
				Rrtype: dns.TypeSOA,
				Class:  dns.ClassINET,
				Ttl:    0,
			}
			if lst := dnsutils.GetResolutionList(entry); len(lst) == 7 {
				rec.Ns, rec.Mbox = lst[0], lst[1]
				serial, _ := strconv.ParseUint(lst[2], 10, 32)
				refresh, _ := strconv.ParseUint(lst[3], 10, 32)
				retry, _ := strconv.ParseUint(lst[4], 10, 32)
				expire, _ := strconv.ParseUint(lst[5], 10, 32)
				minttl, _ := strconv.ParseUint(lst[6], 10, 32)
				rec.Serial, rec.Refresh, rec.Retry, rec.Expire, rec.Minttl = uint32(serial), uint32(refresh), uint32(retry), uint32(expire), uint32(minttl)
			}
			resp.Answer = append(resp.Answer, rec)
		default:
			if verbose {
				fmt.Printf("Unsupported record entry: %d\n", q.Qtype)
			}
		}
	}
}

func handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	if verbose {
		printDate()
		fmt.Printf("Got request: \n%s\n", r.String())
	}
	// setup reply
	result := new(dns.Msg)
	result.SetReply(r)
	// only expect one question
	if len(r.Question) != 1 {
		if verbose {
			printDate()
			println("[WARNING] Recieved %d questions in DNS query\nResponding with no answer.", len(r.Question))
		}
		w.WriteMsg(result)
		return
	}
	// handle question
	answerQuestion(r.Question[0], result)
	// send back response
	if verbose {
		printDate()
		fmt.Printf("Sending response: \n%s\n", result.String())
	}
	w.WriteMsg(result)
}

func main() {
	if err := dnsutils.LoadZones(zoneFile); err != nil {
		printDate()
		fmt.Printf("Failed to load zonefile! ERROR: %s\n", err.Error())
		os.Exit(1)
	}
	if len(os.Args) > 1 && os.Args[1] == "-v" {
		verbose = true
	}
	if verbose {
		fmt.Println("Starting AUTH DNS")
		fmt.Printf("IP: %s\tPORT: %d\n", staticIP, staticPort)
		fmt.Println("Listening...")
	}
	// Define server configurations
	addr := fmt.Sprintf("%s:%d", staticIP, staticPort)
	udpServer := &dns.Server{Addr: addr, Net: "udp"}
	dns.HandleFunc(".", handleRequest)
	//  Run UDP server
	if err := udpServer.ListenAndServe(); err != nil {
		fmt.Printf("Failed to run udpServer! ERROR: %s\n", err.Error())
	}
}
