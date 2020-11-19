module resolver

go 1.15

replace netsec/dnsutils => ../dnsutils

require (
	github.com/miekg/dns v1.1.35
	netsec/dnsutils v0.0.0-00010101000000-000000000000
)
