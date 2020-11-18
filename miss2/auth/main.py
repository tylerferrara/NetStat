
import socket
import os
import json
import glob
import dns.resolver

import dns.zone
from dns.exception import DNSException

PORT = 54
LOCAL_IP = '127.0.0.4'
RESOLVER_IP = '127.0.0.1'

def load_zones(name):
    domain = name
    #zone file to check for the information
    zone_file = "zones/allmyfakesites.com.zone" 
    
    try:
        # Create a zone object from the file 
        zone = dns.zone.from_file(zone_file, domain)

        #Now find all the A records with the domain we are looking for
        mySet = zone.find_rrset(domain, 'A')

        #return that resource record set
        #TODO: Figure out if this is really the return that we need
        return mySet
    
    #if we fail, throw an exception
    except (DNSException):
        print("oof")
        print (DNSException)

# returns (domain, class, rrtype)
def parseQuestion(q):
    arr = q.to_text().split(' ')
    return (arr[0], arr[1], arr[2])

def recursor():
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.bind((LOCAL_IP, PORT))

    while True: 
        data, addr = sock.recvfrom(PORT)
        msg = dns.message.from_wire(data)

        for q in msg.question:
            domain, rrclass, rrtype = parseQuestion(q)
            data, addr = sock.recvfrom(PORT)
            msg = dns.message.from_wire(data)
            
            # Get the right resource record based on the domain
            zoneSet = load_zones(domain)

            #Create a response message from the query message 
            resp = dns.message.make_response(msg)

            #Send the answer back 
            #TODO: I think the way I did this is wrong because of the zoneSet but idk how to fix it
            sendAnswer = dns.resolver.Answer(msg.origin, 'A', zoneSet, resp, nameserver=RESOLVER_IP, port=53)

recursor()