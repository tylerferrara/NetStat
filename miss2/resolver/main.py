
import socket
import os
import dns.resolver

"""
READ THIS: So I've figured out all the nameserver stuff and where things go 
I just can't figure out how to send back the responses and extract the IP address so we can set the TLD_IP and AUTH_IP 
"""

#List of all the IPs of our services 
#To keep it simple we kept them all on one machine 
PORT = 53
LOCAL_IP = '127.0.0.1'
DOT_IP = '127.0.0.2'
TLD_IP = '' #We don't know this!
AUTH_IP = '' #We don't know this!

# takes a dns.rrset.RRset and returns (domain, rrclass, rrtype)
def parseQuestion(q):
    arr = q.to_text().split(' ')
    return (arr[0], arr[1], arr[2])

#Gets the domain query portion for the root to handle (ex. .com)
def getRoot(domain):
    arr = domain.split('.')
    return arr[2]

#Hardcoded TLD nameserver to look for 
def getTLD(domain):
    return "allmyfakesites.com"

# takes dns.message.Message and returns a dns.message.Message
def resolveDot(msg):
    result = dns.message.Message(id=msg.id)
    res = dns.resolver.Resolver(configure=False)
    res.nameservers = [DOT_IP]
    res.port = 54
    # respond to only one question
    for q in msg.question:
        domain, rrclass, rrtype = parseQuestion(q)
        dom = getRoot(domain)
        try:
            answer = res.resolve(dom)
            print(answer.__str__)
        except:
            print("Failed to resolve Root DNS record")
        break
    return result

#Sends a dns.message.Message to the TLD and expects a dns.message.Message back 
def resolveTLD(msg):
    result = dns.message.Message(id=msg.id)
    res = dns.resolver.Resolver(configure=False)
    res.nameservers = [TLD_IP] #TODO: FIND A WAY TO ADD THIS IN
    res.port = 54
    # respond to only one question
    for q in msg.question:
        domain, rrclass, rrtype = parseQuestion(q)
        dom = getRoot(domain) # remove last dot
        try:
            answer = res.resolve(dom)
            print(answer.__str__)
        except:
            print("Failed to resolve Root DNS record")
        break
    return result

#Sends a dns.message.Message to the Authoratative Nameserver and expects a dns.message.Message back 
def resolveAuth(msg):
    result = dns.message.Message(id=msg.id)
    res = dns.resolver.Resolver(configure=False)
    res.nameservers = [AUTH_IP] #TODO: FIND A WAY TO ADD THIS IN
    res.port = 54
    # respond to only one question
    for q in msg.question:
        domain, rrclass, rrtype = parseQuestion(q)
        dom = getRoot(domain) # remove last dot
        try:
            answer = res.resolve(dom)
            print(answer.__str__)
        except:
            print("Failed to resolve Root DNS record")
        break
    return result

def recursor():

    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.bind((LOCAL_IP, PORT))

    # recursive resolver
    # talk to dot
    # talk to TLD
    # talk to Authoritative
    # pass back to client
    while True:
        print("Starting Resolver!")
        data, addr = sock.recvfrom(PORT)
        # get data from client
        msg = dns.message.from_wire(data)
        # talk to dot
        resp3 = resolveDot(msg)
        #TODO: take the response from dot, extract the IP, and fix the TLD IP address 

        # # talk to TLD
        # resp2 = resolveTLD(msg)
        # #TODO: take the response from dot, extract the IP, and fix the Auth IP address 

        # # talk to Auth
        # resp3 = resolveAuth(msg)

        # pass back to client
        sock.sendto(resp3.to_wire(), addr)

recursor()
