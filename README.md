##Trudy

A transparent proxy that can modify and drop traffic for arbitrary TCP connections. Can be used to programmatically modify TCP traffic for proxy-unaware clients. At the moment, no application-layer support is built-in.

###Simple Setup

0. Configure a virtual machine (This was tested on a 64-bit Debian 8 VM) to shove all traffic through Trudy. I personally use a Vagrant VM that sets this up for me (Vagrantfile coming soon).

    `iptables -t nat -A PREROUTING -i eth1 -p tcp -m tcp -j REDIRECT --to-ports 6666`
    
    `ip route del 0/0`
    
    `route add default gw 192.168.1.1 dev eth1`
    
    `sysctl -w net.ipv4.ip_forward=1`

1. Clone the repo on the virutal machine and build the Trudy binary.

    `git clone https://github.com/kelbyludwig/trudy.git`
    
    `cd trudy`
    
    `go install`

2. Run the Trudy binary as root. This starts the listener.

3. Setup your host machine to use the virtual machine as its router. You should see connections being made in Trudy's console but not notice any traffic issues on the host.

###TODO
* Make application-layer modules.
* Signing SSL/TLS traffic with a Trudy certficate.
* Implement a UDP pipe as well.
* Instead of PrettyPrint, define serialize and deserialize. This could allow plug and play for other interfaces.
* Ncurses intercept modification.
