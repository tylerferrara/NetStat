
import socket
import os
import dns.resolver

PORT = 53
LOCAL_IP = '127.0.0.1'
DOT_IP = '127.0.0.2'

# takes a dns.rrset.RRset and returns (domain, rrclass, rrtype)
def parseQuestion(q):
    arr = q.to_text().split(' ')
    return (arr[0], arr[1], arr[2])


# takes dns.message.Message and returns a dns.message.Message
def resolveDot(msg):
    result = dns.message.Message(id=msg.id)
    res = dns.resolver.Resolver(configure=False)
    res.nameservers = [DOT_IP]
    res.port = 54
    # respond to only one question
    for q in msg.question:
        domain, rrclass, rrtype = parseQuestion(q)
        dom = domain[:len(domain)-1] # remove last dot
        try:
            answer = res.resolve(dom, 'NS', rrclass, raise_on_no_answer=False)
            result = answer.response
        except:
            print("Failed to resolve Root DNS record")
        break
    return result

def recursor():

    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.bind((LOCAL_IP, PORT))

    # recursive resolver
    # talk to dot
    # talk to LTD
    # talk to Authoritative
    # pass back to client
    while True:
        data, addr = sock.recvfrom(PORT)
        # get data from client
        msg = dns.message.from_wire(data)
        # talk to dot
        resp = resolveDot(msg)

        # talk to LTD

        # talk to Auth

        # pass back to client
        sock.sendto(resp.to_wire(), addr)


recursor()