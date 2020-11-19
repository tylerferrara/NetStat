# Mission 2 (DNS)

The infrastructure includes:
- Root Server
- TLD Namerserver
- 1 Authoritative Nameserver
- DNS Recursor

Each service depends on a few IP addresses being available to function. To set this up, check your active network devices:
```sh
ip addr
```
For this example, we will use to use **ens3** as our network device
#### Create IP Aliases
```sh
sudo ip addr add 127.0.0.2 dev ens3
sudo ip addr add 127.0.0.3 dev ens3
sudo ip addr add 127.0.0.4 dev ens3
#To delete them:
sudo ip addr del 127.0.0.2 dev ens3
sudo ip addr del 127.0.0.3 dev ens3
sudo ip addr del 127.0.0.4 dev ens3
```
### Start Services
Each service starts a UDP server and persists until a interupt SIG. Therefore, you should either run each service in the background or in a separate terminal.
```sh
# pass -v flag for debugging output
go run ./root/root.go
go run ./tld/tld.go
go run ./auth/auth.go
go run ./resolver/resolver.go
```

### Test Resolver
Currently, the only two domains available for succussful query are **fakesite.com** and **fakesite1.com**
```sh
# using dig or drill
dig fakesite.com @172.0.0.1 SOA
```
