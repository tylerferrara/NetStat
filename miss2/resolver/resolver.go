package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"netsec/dnsutils"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// Max cache limit
const cacheCapacity = 200

const sigKey = "reskey."
const sigVal = "92kslfjwlOWPk0s=99="

// DNSSEC
var tsigMap = map[string]string{
	"rootkey.": "BB6zGir4GPAqINNh9U5c3A==", // known root key
	"tldkey.":  "cB6zGir4GPAqINNh9U5c3A==",
	"authkey.": "tt6zGir4GPAqINNh9U5c3A==",
	sigKey:     sigVal,
}

// UDP Packet Size
const udpSize = 4096

// Network info
const staticIP = "127.0.0.1" // "10.21.4.1"
const staticPort = 53
const extraPort = 8071
const rootIP = "127.0.0.2" // "10.21.4.2"
const rootPort = 8082      // 53

// Flags
var verbose bool
var dnssec bool

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

// returns the org. from "domain.org.""
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
	// create client
	c := new(dns.Client)
	c.Dialer = &net.Dialer{
		Timeout: 200 * time.Millisecond,
		LocalAddr: &net.UDPAddr{
			IP:   net.ParseIP(staticIP),
			Port: extraPort,
			Zone: "",
		},
	}
	if dnssec {
		msg.SetEdns0(udpSize, true)
		c.TsigSecret = tsigMap
		msg.SetTsig("rootkey.", dns.HmacSHA512, 3000, time.Now().Unix())
	}
	in, _, err := c.Exchange(msg, rootAddr)
	if err != nil {
		return "", err
	}
	if dnssec {
		if in.IsTsig() == nil {
			if verbose {
				printDate()
				fmt.Println("=== NO TSIG FROM ROOT")
			}
			return "", errors.New("Root responded without TSIG when DNSSEC is enabled")
		}
		if verbose {
			printDate()
			fmt.Println("=== VALID TSIG RESPONSE FROM ROOT")
		}
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
	port := 8083 // 53
	msg := new(dns.Msg)
	tldAddr := fmt.Sprintf("%s:%d", ip, port)
	// fetch domain with SOA
	msg.SetQuestion(q.Name, dns.TypeSOA)
	// create client
	c := new(dns.Client)
	c.Dialer = &net.Dialer{
		Timeout: 200 * time.Millisecond,
		LocalAddr: &net.UDPAddr{
			IP:   net.ParseIP(staticIP),
			Port: extraPort,
			Zone: "",
		},
	}
	if dnssec {
		msg.SetEdns0(udpSize, true)
		c.TsigSecret = tsigMap
		msg.SetTsig("tldkey.", dns.HmacSHA512, 3000, time.Now().Unix())
	}
	in, _, err := c.Exchange(msg, tldAddr)
	if err != nil {
		return "", err
	}
	if dnssec {
		if in.IsTsig() == nil {
			if verbose {
				printDate()
				fmt.Println("=== NO TSIG FROM TLD")
			}
			return "", errors.New("TLD responded without TSIG when DNSSEC is enabled")
		}
		if verbose {
			printDate()
			fmt.Println("=== VALID TSIG RESPONSE FROM TLD")
		}
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
	if dnssec {
		msg.SetEdns0(udpSize, true)
		c.TsigSecret = tsigMap
		msg.SetTsig("tldkey.", dns.HmacSHA512, 3000, time.Now().Unix())
	}
	ina, _, err := c.Exchange(msg, tldAddr)
	if err != nil {
		return "", nil
	}
	if dnssec {
		if ina.IsTsig() == nil {
			if verbose {
				printDate()
				fmt.Println("=== NO TSIG FROM TLD")
			}
			return "", errors.New("TLD responded without TSIG when DNSSEC is enabled")
		}
		if verbose {
			printDate()
			fmt.Println("=== VALID TSIG RESPONSE FROM TLD")
		}
	}
	if len(ina.Answer) == 0 {
		return "", errors.New("TLD Nameserver gave empty answer to A request")
	}
	authIP, err = getIPFromRR(ina.Answer[0])
	return authIP, err
}

// cunsult Authouritative DNS to get the requested record
func queryAuth(id uint16, ip string, q dns.Question) (res *dns.Msg, err error) {
	// ****************************************
	// ***********************************************
	// NOTE: ***********************************************
	// When we deploy this, the default port will be 53
	msg := new(dns.Msg)
	port := 8084 // 53
	authAddr := fmt.Sprintf("%s:%d", ip, port)
	msg.SetQuestion(q.Name, q.Qtype)
	msg.Id = id
	// create client
	c := new(dns.Client)
	c.Dialer = &net.Dialer{
		Timeout: 200 * time.Millisecond,
		LocalAddr: &net.UDPAddr{
			IP:   net.ParseIP(staticIP),
			Port: extraPort,
			Zone: "",
		},
	}
	if dnssec {
		msg.SetEdns0(udpSize, true)
		c.TsigSecret = tsigMap
		msg.SetTsig("authkey.", dns.HmacSHA512, 3000, time.Now().Unix())
	}
	res, _, err = c.Exchange(msg, authAddr)
	if dnssec {
		if res.IsTsig() == nil {
			if verbose {
				printDate()
				fmt.Println("=== NO TSIG FROM AUTH")
			}
			return res, errors.New("TLD responded without TSIG when DNSSEC is enabled")
		}
		if verbose {
			printDate()
			fmt.Println("=== VALID TSIG RESPONSE FROM TLD")
		}
	}
	return res, err
}

func handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	if verbose {
		printDate()
		fmt.Printf("Got request:\n%s\n", r.String())
	}
	// Look in cache
	if res, hit := dnsutils.GetCacheVal(r); hit {
		// cache hit
		res.SetReply(r)
		if verbose {
			printDate()
			fmt.Printf("Cache hit:\n%s\n", res.String())
		}
		w.WriteMsg(&res)
		return
	} else if verbose {
		printDate()
		fmt.Printf("Cache miss:\n%s\n", r.String())
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
	result, err = queryAuth(r.Id, authIP, r.Question[0])
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
	// populate cache
	dnsutils.PushCache(r, result)
	w.WriteMsg(result)
}

// handle flags
func parseFlags() {
	flag.BoolVar(&verbose, "v", false, "verbose debug output")
	flag.BoolVar(&dnssec, "s", false, "enable dnssec")
	flag.Parse()
}

func main() {
	parseFlags()
	if verbose {
		fmt.Println("Starting RESOLVER DNS")
		fmt.Printf("IP: %s\tPORT: %d\n", staticIP, staticPort)
		if dnssec {
			fmt.Println("=== DNSSEC ENABLED ===")
		}
		fmt.Println("Listening...")
	}
	// init cache
	dnsutils.InitCache(cacheCapacity)
	// Define server configurations
	addr := fmt.Sprintf("%s:%d", staticIP, staticPort)
	server := &dns.Server{Addr: addr, Net: "udp"}
	if dnssec {
		server.TsigSecret = tsigMap
	}
	// Bind handler
	dns.HandleFunc(".", handleRequest)
	// Listen
	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("Failed to start RESOLVER! ERROR: %s\n", err.Error())
	}
}
