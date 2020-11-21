#DNS Spoofing/DNS Cache Poisoning Attack 

##To start, first install all dependencies 

```
sudo apt-get update
sudo apt-get install -y libpcap-dev
sudo apt-get install python3
sudo apt-get install build-essential python-dev libnetfilter-queue-dev
pip3 install netfilterqueue scapy
```

##To perform the attack:

**First run all of the DNS servers as shown in the Miss2 ReadMe**

Run the attackMe file as root:
```
sudo python3 attackMe.py
```

Now send a DNS request to the recursor from any machine:
```
dig fakesite.com @127.0.0.1 A
```
The recursor should now return a bogus IP (123.4.5.6) for fakesite.com rather than the one inside the DNS zone file (127.0.0.5)


