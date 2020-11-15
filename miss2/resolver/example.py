
import socket
import glob
import json
# import dns
# import dns.resolver

# result = dns.resolver.resolve('tutorialspoint.com', 'A')
# for ipval in result:
    # print('IP', ipval.to_text())


port = 53
ip = '127.0.0.1' # loopback

sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
sock.bind((ip, port))

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


def getQuestionDomain(data):
    domain = ""
    bLen = ord(data[:1])
    while bLen > 0:
        domain += (data[1:bLen+1].decode('utf-8'))
        domain += '.'
        data = data[bLen+1:]
        bLen = ord(data[:1])
    q = ''
    recType = ord(data[2:3])
    if recType == 1:
        q = 'a'
    
    return (domain, q)

def getRecs(data):
    domain, question = getQuestionDomain(data)
    zone = getZone(domain)
    return (zone[question], question, domain)

# handles resource records
def rectobytes(domain, rectype, recttl, recval):
    rBytes = b'\xc0\x0c'
    # type
    if rectype == 'a':
        rBytes += bytes([0]) + bytes([1])
    else:
        rBytes += bytes([0]) + bytes([1])
    # class
    rBytes += bytes([0]) + bytes([1])
    # ttl
    rBytes += int(recttl).to_bytes(4, byteorder='big')
    # rdlength
    if rectype == 'a':
        rBytes += bytes([0]) + bytes([4])
    else:
        rBytes += bytes([0]) + bytes([4])
    # data
    for part in recval.split('.'):
        rBytes += bytes([int(part)])

    return rBytes

def buildQuest(domain, recType):
    qBytes = b''
    for part in domain.split("."):
        length = len(part)
        qBytes += bytes([length])

        for char in part:
            qBytes += ord(char).to_bytes(1, byteorder='big')
    
    if recType == 'a':
        qBytes += (1).to_bytes(2, byteorder='big')
    else: # other records
        qBytes += (1).to_bytes(2, byteorder='big')

    qBytes += (1).to_bytes(2, byteorder='big')

    return qBytes

def buildResp(data):
    # build header
    tID = ''
    TransID = data[0:2]
    for b in TransID:
        tID += hex(b)[2:]
    
    Flags = getFlags(data[2:4])
    QDCOUNT = b'\x00\x01'
    rec, recType, domainName = getRecs(data[12:])
    ANCOUNT = len(rec).to_bytes(2, byteorder='big')
    NSCOUNT = (0).to_bytes(2, byteorder='big')
    ARCOUNT = (0).to_bytes(2, byteorder='big')
    dnsHeader = TransID+Flags+QDCOUNT+ANCOUNT+NSCOUNT+ARCOUNT
    # build question
    dnsQuestion = buildQuest(domainName, recType)
    # build body
    dnsBody = b''
    for record in rec:
        dnsBody += rectobytes(domainName, recType, record["ttl"], record["value"])


    return dnsHeader+dnsQuestion+dnsBody

def getFlags(flags):
    QR = '1'
    b1 = bytes(flags[0:1])
    OPCODE = ''
    for bit in range(1,5):
        OPCODE += str(ord(b1)&(1<<bit))
    
    AA = '1'
    TC = '0'
    RD = '0'
    RA = '0'
    Z = '000'
    RCODE ='0000'
    return int(QR+OPCODE+AA+TC+RD, 2).to_bytes(1, byteorder='big')+int(RA+Z+RCODE,2).to_bytes(1, byteorder='big')


while True:
    data, addr = sock.recvfrom(512)
    r = buildResp(data)
    sock.sendto(r, addr)