#DNS Amplification Attack: Proof of Concept

##To start, first install all dependencies 

```
sudo apt-get update
sudo apt-get install -y libpcap-dev
sudo apt-get install python3
sudo apt-get install build-essential python-dev libnetfilter-queue-dev
pip3 install scapy
```

##To perform the attack:

**First run all of the DNS servers as shown in the Miss2 ReadMe. You can use the -s flag to run DNSSEC**

Run the amplAttack file as root:
```
sudo python3 amplAttack.py
```
You should see a print output with packet sizes and an amplification factor. 