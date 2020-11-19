package dnsutils

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"

	"github.com/bwesterb/go-zonefile"
	"github.com/miekg/dns"
)

// Holds zone information after call to loadZones
var zoneData *zonefile.Zonefile

// Cache datastructure
var cacheData map[string]dns.Msg
var cacheKeys []string // first in last out (OLDEST in front)

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

// zeroMsgID standardizes the ID so cache can find them later
func zeroMsgID(msg *dns.Msg) {
	msg.Id = 0
}

// GetCacheVal returns a msg found and true if cache hit
func GetCacheVal(key *dns.Msg) (val dns.Msg, hit bool) {
	// remove key ID
	id := key.Id
	zeroMsgID(key)
	// store key as string
	byteKey, err := key.Pack()
	if err != nil {
		fmt.Println("[WARNING] msg can't be packed into string format!")
	}
	strKey := string(byteKey)
	val, hit = cacheData[strKey]
	if hit {
		// re-order keys
		reord := false
		for idx, k := range cacheKeys {
			if reflect.DeepEqual(k, strKey) {
				reord = true
				if idx == 0 {
					// beginning
					cacheKeys = append(make([]string, 0, cap(cacheKeys)), cacheKeys[1:]...)
					cacheKeys = append(cacheKeys, strKey)
				} else if idx < len(cacheKeys)-1 {
					// middle
					left := cacheKeys[:idx]
					right := cacheKeys[idx+1:]
					cacheKeys = append(make([]string, 0, cap(cacheKeys)), left...)
					cacheKeys = append(cacheKeys, right...)
					cacheKeys = append(cacheKeys, strKey)
				}
			}
		}
		if !reord {
			fmt.Println("\n[WARNING] Cache re-ording was not accomplished!\nCheck equality")
		}
	} else {
		fmt.Println("KEYs don't match")
		fmt.Println("Given:")
		fmt.Println(strKey)
		if len(cacheKeys) > 0 {
			fmt.Println("Stored Tail:")
			fmt.Println(cacheKeys[len(cacheKeys)-1])

		}
		fmt.Println("map:")
	}
	// re-add key ID
	key.Id = id
	return val, hit
}

// PushCache adds a message to the cache
func PushCache(key *dns.Msg, val *dns.Msg) {
	max := cap(cacheKeys)
	// remove ID
	zeroMsgID(key)
	// store key as string
	byteKey, err := key.Pack()
	if err != nil {
		fmt.Println("[WARNING] msg can't be packed into string format!")
	}
	strKey := string(byteKey)
	// check length
	if len(cacheKeys) == max {
		// remove first
		last := cacheKeys[0]
		// adjust keys
		cacheKeys = append(make([]string, 0, max), cacheKeys[1:]...)
		// remove key from map
		delete(cacheData, last)
	}
	cacheKeys = append(cacheKeys, strKey)
	cacheData[strKey] = *val
}

// InitCache creates a map of given size
func InitCache(max int) {
	cacheData = make(map[string]dns.Msg)
	// length: 0, capacity: max
	cacheKeys = make([]string, 0, max)
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
