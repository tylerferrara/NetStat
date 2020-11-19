package dnsutils

import (
	"io/ioutil"
	"strings"

	"github.com/bwesterb/go-zonefile"
	"github.com/miekg/dns"
)

// Holds zone information after call to loadZones
var zoneData *zonefile.Zonefile

// classToInt converts class string to it's number complement
func classToInt(c string) uint16 {
	// Defined by RFC-1035
	// IN              1 the Internet
	// CS              2 the CSNET class (Obsolete)
	// CH              3 the CHAOS class
	// HS              4 Hesiod [Dyer 87]
	switch strings.ToUpper(c) {
	case "IN":
		return 1
	case "CS":
		return 2
	case "CH":
		return 3
	case "HS":
		return 4
	}
	return 0
}

// typeToInt converts type string to it's number complement
func typeToInt(t string) uint16 {
	// Defined by RFC-1035
	// TYPE            value and meaning
	// -------------------------------------------
	// A               1 a host address
	// NS              2 an authoritative name server
	// MD              3 a mail destination (Obsolete - use MX)
	// MF              4 a mail forwarder (Obsolete - use MX)
	// CNAME           5 the canonical name for an alias
	// SOA             6 marks the start of a zone of authority
	// MB              7 a mailbox domain name (EXPERIMENTAL)
	// MG              8 a mail group member (EXPERIMENTAL)
	// MR              9 a mail rename domain name (EXPERIMENTAL)
	// NULL            10 a null RR (EXPERIMENTAL)
	// WKS             11 a well known service description
	// PTR             12 a domain name pointer
	// HINFO           13 host information
	// MINFO           14 mailbox or mail list information
	// MX              15 mail exchange
	// TXT             16 text strings
	switch strings.ToUpper(t) {
	case "A":
		return 1
	case "NS":
		return 2
	case "MD":
		return 3
	case "MF":
		return 4
	case "CNAME":
		return 5
	case "SOA":
		return 6
	case "MB":
		return 7
	case "MG":
		return 8
	case "MR":
		return 9
	case "NULL":
		return 10
	case "WKS":
		return 11
	case "PTR":
		return 12
	case "HINFO":
		return 13
	case "MINFO":
		return 14
	case "MX":
		return 15
	case "TXT":
		return 16
	}
	return 0
}

// GetResolutionList returns the last portion of the entry record
func GetResolutionList(entry zonefile.Entry) (result []string) {
	s := strings.Index(entry.String(), "[")
	e := strings.Index(entry.String(), "]")
	// capture everything between brackets [ ... ]
	substr := entry.String()[s+1 : e]
	// remove all double quotes
	substr = strings.ReplaceAll(substr, "\"", "")
	// parse string to list
	result = strings.Split(substr, " ")
	return result
}

// LoadZones initializes the given zoneFile into memory.
// This method must be called before using GetZones
func LoadZones(zoneFile string) error {
	data, err := ioutil.ReadFile(zoneFile)
	if err != nil {
		return err
	}
	zf, err := zonefile.Load(data)
	if err != nil {
		return err
	}
	zoneData = zf
	return nil
}

// GetZones finds all entries that match the given question
// Performance O(N) - can improve with hashmap/dictionary
func GetZones(q dns.Question) (r []zonefile.Entry) {
	for _, e := range zoneData.Entries() {
		if string(e.Domain()) == q.Name &&
			classToInt(string(e.Class())) == q.Qclass &&
			typeToInt(string(e.Type())) == q.Qtype {
			r = append(r, e)
		}
	}
	return r
}
