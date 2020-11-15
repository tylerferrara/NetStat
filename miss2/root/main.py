
import socket
import os
import json
import glob
import dns.resolver

PORT = 54
LOCAL_IP = '127.0.0.2'

def load_zones():
    zonefiles = glob.glob('zones/*.zone')
    z = {}
    for zone in zonefiles:
        with open(zone) as zonedata:
            data = json.load(zonedata)
            zonename = data["$origin"]
            z[zonename] = data
    return z

zonedata = load_zones()

def getZone(domain):
    global zonedata
    return zonedata[domain]

# domain.org. returns org.
def getOrg(domain):
    dotCount = 0
    i = len(domain) - 1
    while i >= 0:
        if domain[i] == '.':
            dotCount = dotCount + 1
        if dotCount == 2:
            return domain[i+1:]
        i = i - 1
    return None


# returns (domain, class, rrtype)
def parseQuestion(q):
    arr = q.to_text().split(' ')
    return (arr[0], arr[1], arr[2])

def recursor():
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.bind((LOCAL_IP, PORT))

    while True:
        data, addr = sock.recvfrom(PORT)
        # get data from client
        msg = dns.message.from_wire(data)
        # TODO: match with correct SERVER
        resp = dns.message.Message(id=msg.id)
        for q in msg.question:
            domain, rrclass, rrtype = parseQuestion(q)
            # answer = dns.resolver.resolve(domain, rrtype, rrclass, raise_on_no_answer=False)
            # resp = answer.response

            # zone = getZone(domain)
            # rrtype = rrtype.lower()
            # if zone is None:
            #     print("zone not found for domain: " + domain)
            # elif rrtype not in zone:
            #     print("zone found, but record type not in file domain:" + domain + " type: " + rrtype)
            # else:
            #     rec = zone[rrtype]

            #     name = getOrg(domain)
            #     if name == None:
            #         print("Unexpected Error!: could not get org from domain: " + domain)
                
                # for r in rec: # assuming ns record
                #     resp.authority.append(dns.rrset.RRset(r["host"], rrclass, rrtype))

            answer = dns.resolver.resolve(domain, rrtype, rrclass)
            resp = answer.response
            resp.id = msg.id
            break

        sock.sendto(resp.to_wire(), addr)

recursor()