module auth

go 1.15

replace netsec/dnsutils => ../dnsutils

require (
	github.com/bwesterb/go-zonefile v1.0.0
	github.com/miekg/dns v1.1.35
	netsec/dnsutils v0.0.0-00010101000000-000000000000
)
