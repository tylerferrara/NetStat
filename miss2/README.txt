To run:

#First we need the dns library from source
git clone https://github.com/rthalley/dnspython
cd dnspython
python setup.py install

#Then we need to add more ips to our machine 

#For MacOS:
sudo ifconfig en0 alias 127.0.0.2 up 
sudo ifconfig en0 alias 127.0.0.3 up 
sudo ifconfig en0 alias 127.0.0.4 up 
#To delete it:
sudo ifconfig en0 -alias 127.0.0.2
sudo ifconfig en0 -alias 127.0.0.3
sudo ifconfig en0 -alias 127.0.0.4

#For Linux
sudo ip addr add 127.0.0.2 dev ens3
sudo ip addr add 127.0.0.3 dev ens3
sudo ip addr add 127.0.0.4 dev ens3
#To delete it:
sudo ip addr del 127.0.0.2 dev ens3
sudo ip addr del 127.0.0.3 dev ens3
sudo ip addr del 127.0.0.4 dev ens3

RUN THESE ON SEPERATE TERMINAL WINDOWS:

cd resolver 
#need to run as root user 
sudo python3 main.py

cd ..

cd root
#need to run as root user 
sudo python3 main.py

cd ..


cd tld
#need to run as root user 
sudo python3 main.py

cd ..

cd auth 
#need to run as root user 
sudo python3 main.py

#Now open a new terminal window and type in:
dig site.com @127.0.0.1