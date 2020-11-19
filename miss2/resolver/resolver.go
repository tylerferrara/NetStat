package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// Network info
const staticIP = "127.0.0.1"
const staticPort = 53
const rootIP = "127.0.0.2"
const rootPort = 8020

var verbose bool

func printDate() {
	fmt.Printf("--------- %s ---------\n", time.Now().Format("2006-01-02 15:04:05.000000"))
}

// expects to run on SOA records
func getNsFromRR(a dns.RR) (r string, e error) {
	s := a.String()
	s = strings.ReplaceAll(s, "\t", " ")
	lst := strings.Split(s, " ")
	if len(lst) > 4 {
		return lst[4], nil
	}
	return "", errors.New("Malformed dns.RR expecting (expecting SOA record when calling getNsFromRR)")
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

// consult Root DNS for TLD Nameserver's IP address
func queryRoot(r dns.Question) (ip string, e error) {
	msg := new(dns.Msg)
	// get just the last part of the domain
	domain, err := getLastSubDomain(r.Name)
	if err != nil {
		return "", err
	}
	msg.SetQuestion(domain, dns.TypeA)
	rootAddr := fmt.Sprintf("%s:%d", rootIP, rootPort)
	in, err := dns.Exchange(msg, rootAddr)
	if err != nil {
		return "", err
	}
	if len(in.Answer) == 0 {
		return "", errors.New("Answer from Root DNS server has no answer")
	}
	// obtain TLD Nameserver address
	tldIP, err := getIPFromRR(in.Answer[0])
	if err != nil {
		return "", err
	}
	return tldIP, nil
}

// consult TLD Nameserver for Authoritative DNS IP
func queryTLD(ip string, q dns.Question) (authIP string, e error) {
	// ****************************************
	// ***********************************************
	// NOTE: ***********************************************
	// When we deploy this, the default port will be 53
	port := 8083
	msg := new(dns.Msg)
	tldAddr := fmt.Sprintf("%s:%d", ip, port)
	// fetch domain with SOA
	msg.SetQuestion(q.Name, dns.TypeSOA)
	in, err := dns.Exchange(msg, tldAddr)
	if err != nil {
		return "", err
	}
	if len(in.Answer) == 0 {
		return "", errors.New("TLD Nameserver gave empty answer to SOA request")
	}
	// obtain IP of AUTH server
	ns, err := getNsFromRR(in.Answer[0])
	if err != nil {
		return "", err
	}
	msg = new(dns.Msg)
	msg.SetQuestion(ns, dns.TypeA)
	ina, err := dns.Exchange(msg, tldAddr)
	if err != nil {
		return "", nil
	}
	if len(ina.Answer) == 0 {
		return "", errors.New("TLD Nameserver gave empty answer to A request")
	}
	authIP, err = getIPFromRR(ina.Answer[0])
	return authIP, err
}

// cunsult Authouritative DNS to get the requested record
func queryAuth(ip string, q dns.Question) (res *dns.Msg, err error) {
	// ****************************************
	// ***********************************************
	// NOTE: ***********************************************
	// When we deploy this, the default port will be 53
	msg := new(dns.Msg)
	port := 8082
	authAddr := fmt.Sprintf("%s:%d", ip, port)
	msg.SetQuestion(q.Name, q.Qtype)
	res, err = dns.Exchange(msg, authAddr)
	return res, err
}

func handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	if verbose {
		printDate()
		fmt.Printf("Got request:\n%s", r.String())
	}
	result := new(dns.Msg)
	result.SetReply(r)
	if len(r.Question) != 1 {
		printDate()
		fmt.Printf("Incoming request doesn't have a single question!\nReturning an empty reply")
		w.WriteMsg(result)
		return
	}
	// ROOT
	tldIP, err := queryRoot(r.Question[0])
	if err != nil {
		printDate()
		fmt.Printf("Stopped at Root DNS server\n%s\n", err.Error())
		w.WriteMsg(result)
		return
	}
	// TLD
	authIP, err := queryTLD(tldIP, r.Question[0])
	if err != nil {
		printDate()
		fmt.Printf("Stopped at TLD Nameserver\n%s\n", err.Error())
		w.WriteMsg(result)
		return
	}
	// AUTH
	result, err = queryAuth(authIP, r.Question[0])
	if err != nil {
		printDate()
		fmt.Printf("Stopped at Authoritative DNS\n%s\n", err.Error())
		w.WriteMsg(result)
		return
	}
	if verbose {
		printDate()
		fmt.Printf("Sending valid response:\n%s", result.String())
	}
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
