##Trudy

A transparent proxy that can modify and drop traffic for arbitrary TCP connections. Can be used to programmatically modify TCP traffic for proxy-unaware clients.

Trudy was designed for monitoring and modifying proxy-unaware devices that use non-HTTP protocols. If you want to intercept and modify HTTP(S) traffic, Burp Suite is probably the better option. 

## Author

Written by Kelby Ludwig ([@kelbyludwig](https://twitter.com/kelbyludwig))


###Why I Built This
I have done security research that invovled sitting between a embedded device and a server and modifying some custom binary protocol on the fly. This usually is a slow process that involves sniffing legitimate traffic, and then rebuilding packets programmatically. Trudy enables Burp-like features for generalized TCP traffic.

###Simple Setup

0. Configure a virtual machine (Trudy has been tested on a 64-bit Debian 8 VM) to shove all traffic through Trudy. I personally use a Vagrant VM that sets this up for me (Vagrantfile coming soon).

    `iptables -t nat -A PREROUTING -i eth1 -p tcp --dport 8888 -m tcp -j REDIRECT --to-ports 8080`

    `iptables -t nat -A PREROUTING -i eth1 -p tcp --dport 443 -m tcp -j REDIRECT --to-ports 6443`

    `iptables -t nat -A PREROUTING -i eth1 -p tcp -m tcp -j REDIRECT --to-ports 6666`
    
    `ip route del 0/0`
    
    `route add default gw 192.168.1.1 dev eth1`
    
    `sysctl -w net.ipv4.ip_forward=1`

1. Clone the repo on the virutal machine and build the Trudy binary.

    `git clone https://github.com/kelbyludwig/trudy.git`
    
    `cd trudy`
    
    `go install`

2. Run the Trudy binary as root. This starts the listeners. If you ran the iptables commands above, iptables will forward traffic destined for port 443 to port 6443. Trudy listens on this port and expects traffic coming into this port to be TLS. All other TCP connections will be forwarded through port 6666.

    `sudo $GOPATH/bin/trudy`

3. Setup your host machine to use the virtual machine as its router. You should see connections being made in Trudy's console but not notice any traffic issues on the host.

4. In order to manipulate data, just implement the following functions within the `module` package. 

    `DoMangle`

    `Mangle`

    `Drop`

    `PrettyPrint`

    `DoPrint`

    `DoIntercept`

5. To access the interceptor, visit `http://<IP ADDRESS OF VM>:8888/` in your web browser.

###Coming soon
* Instead of PrettyPrint, define serialize and deserialize.
* Implement a UDP pipe.
* Command line flags for easy customization.

###Coming at some point
* On-the-fly TLS certificate generation.
