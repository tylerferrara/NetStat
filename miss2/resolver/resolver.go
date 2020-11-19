package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/miekg/dns"
)

// Network info
const staticIP = "127.0.0.1"
const staticPort = 53
const rootIP = "127.0.0.2"
const rootPort = 8020

var verbose bool

// expects to run on SOA records
func getNsFromRR(a dns.RR) (r string, e error) {
	s := a.String()
	lst := strings.Split(s, " ")
	fmt.Println(lst)
	return "", nil
}

// expects to run on A records
func getIPFromRR(a dns.RR) (r string, e error) {
	s := a.String()
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '\t' || s[i] == ' ' {
			return s[i+1:], nil
		}
	}
	return "", errors.New("Malformed RR without tab delimitor when converting to string")
}

// returns the org. from domain.org.
func getLastSubDomain(s string) (r string, e error) {
	dotCount := 0
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '.' {
			dotCount++
		}
		if dotCount == 2 {
			return s[i+1:], nil
		}
	}
	return "", errors.New("Malformed domain, containing less than two periods")
}

func queryRoot(r dns.Question) (m *dns.Msg, e error) {
	msg := new(dns.Msg)
	// get just the last part of the domain
	domain, err := getLastSubDomain(r.Name)
	if err != nil {
		return msg, err
	}
	msg.SetQuestion(domain, dns.TypeA)
	rootAddr := fmt.Sprintf("%s:%d", rootIP, rootPort)
	in, err := dns.Exchange(msg, rootAddr)
	if err != nil {
		return msg, err
	}
	return in, nil
}

func queryTLD(ip string, q dns.Question) (m *dns.Msg, e error) {
	// ****************************************
	// ***********************************************
	// NOTE: ***********************************************
	// When we deploy this, the default port will be 53
	port := 8083
	msg := new(dns.Msg)
	rootAddr := fmt.Sprintf("%s:%d", ip, port)
	// fetch domain with SOA
	msg.SetQuestion(q.Name, dns.TypeSOA)
	in, err := dns.Exchange(msg, rootAddr)
	if err != nil {
		return msg, err
	}
	if len(in.Answer) == 0 {
		return in, errors.New("TLD Nameserver gave empty answer to SOA request")
	}
	// obtain IP of AUTH server
	fmt.Println(in.Answer[0])
	ns, err := getNsFromRR(in.Answer[0])
	if err != nil {
		return msg, err
	}
	fmt.Println(ns)
	return nil, nil
}

func queryAuth(r *dns.Msg) *dns.Msg {
	// result := new(dns.Msg)
	return r
}

func handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	result := new(dns.Msg)
	result.SetReply(r)
	if len(r.Question) != 1 {
		fmt.Printf("Incoming request doesn't have a single question!\nReturning an empty reply")
		w.WriteMsg(result)
	}
	// ROOT
	ans, err := queryRoot(r.Question[0])
	if err != nil {
		fmt.Printf("Can't reach Root DNS server\n%s\n", err.Error())
		w.WriteMsg(result)
	}
	if len(ans.Answer) == 0 {
		fmt.Printf("Answer from Root DNS server has no answer")
		w.WriteMsg(result)
	}
	tldIP, err := getIPFromRR(ans.Answer[0])
	if err != nil {
		fmt.Printf("Unable to get IP for TLD server in record\n%s\n", err.Error())
	}
	// TLD
	ans, err = queryTLD(tldIP, r.Question[0])
	if err != nil {
		fmt.Printf("Can't reach TLD Nameserver\n%s\n", err.Error())
		w.WriteMsg(result)
	}
	// // AUTH
	// result = queryAuth(r)
	w.WriteMsg(result)
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-v" {
		verbose = true
	}
	if verbose {
		fmt.Println("Starting RESOLVER DNS")
		fmt.Printf("IP: %s\tPORT: %d\n", staticIP, staticPort)
		fmt.Println("Listening...")
	}
	// Define server configurations
	addr := fmt.Sprintf("%s:%d", staticIP, staticPort)
	server := &dns.Server{Addr: addr, Net: "udp"}
	// Bind handler
	dns.HandleFunc(".", handleRequest)
	// Listen
	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("Failed to start RESOLVER! ERROR: %s\n", err.Error())
	}
}
