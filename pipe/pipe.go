package pipe

import (
	"crypto/tls"
	"log"
	"net"
	"strconv"
	"syscall"
	"time"
)

//Netfilter/iptables adds a tcp header to identify original destination.
//Since all traffic is routed through trudy, we need to retrieve the original
//intended destination (i.e. _not_ trudy)
const SO_ORIGINAL_DST = 80

//TrudyPipe is the primary interface that handles connections. TrudyPipe creates a full-duplex
//pipe that passes data from the client to the server and vice versa. A pipe is compromised of
//two connections. The client transparently connects to Trudy, and Trudy accepts the connection.
//Trudy will then make a connection with the client's intended destination and just pass traffic
//back-and-forth between the two connections. All modifications and drops to the packet happen
//to data between the two ends of the pipe.
type TrudyPipe interface {
	ReadSource(buffer []byte) (n int, err error)
	WriteSource(buffer []byte) (n int, err error)
	ReadDestination(buffer []byte) (n int, err error)
	WriteDestination(buffer []byte) (n int, err error)
	DestinationInfo() (addr net.Addr)
	SourceInfo() (addr net.Addr)
	New(uint, int, net.Conn) (err error)
	Close()
}

//TCPPipe implements the TrudyPipe interface and can be used to proxy generic TCP connections.
type TCPPipe struct {
	id          uint
	destination net.Conn
	source      net.Conn
}

//DestinationInfo returns the net.Addr of the destination.
func (t *TCPPipe) DestinationInfo() (addr net.Addr) {
	addr = t.destination.RemoteAddr()
	return
}

func (t *TCPPipe) SourceInfo() (addr net.Addr) {
	addr = t.source.RemoteAddr()
	return
}

//Close closes both ends of a TCPPipe.
func (t *TCPPipe) Close() {
	log.Printf("[INFO] ( %v ) Closing TCP connection.\n", t.id)
	t.source.Close()
	t.destination.Close()
}

//ReadSource reads data from the source end of the pipe. This is typically the proxy-unaware client.
func (t *TCPPipe) ReadSource(buffer []byte) (n int, err error) {
	t.source.SetReadDeadline(time.Now().Add(15 * time.Second))
	return t.source.Read(buffer)
}

//WriteSource writes data to the source end of the pipe. This is typically the proxy-unaware client.
func (t *TCPPipe) WriteSource(buffer []byte) (n int, err error) {
	t.source.SetWriteDeadline(time.Now().Add(15 * time.Second))
	return t.source.Write(buffer)
}

//ReadDestination reads data from the destination end of the pipe. The destination is the proxy-unaware
//client's intended destination.
func (t *TCPPipe) ReadDestination(buffer []byte) (n int, err error) {
	t.destination.SetReadDeadline(time.Now().Add(15 * time.Second))
	return t.destination.Read(buffer)
}

//WriteDestination writes data to the destination end of the pipe. The destination is the proxy-unaware
//client's intended destination.
func (t *TCPPipe) WriteDestination(buffer []byte) (n int, err error) {
	t.destination.SetWriteDeadline(time.Now().Add(15 * time.Second))
	return t.destination.Write(buffer)
}

//byteToConnString converts the Multiaddr bytestring returned by Getsockopt into a "host:port" connection string.
func byteToConnString(multiaddr [16]byte) string {
	ip := multiaddr[4:8]
	ipString := net.IPv4(ip[0], ip[1], ip[2], ip[3]).String()
	port := multiaddr[2:4]
	portUint := int64((uint32(port[0]) << 8) + uint32(port[1]))
	portString := strconv.FormatInt(portUint, 10)
	return (ipString + ":" + portString)
}

//New creates a new TCPPipe.
func (t *TCPPipe) New(id uint, fd int, sourceConn net.Conn) (err error) {
	//TODO: Make the second argument system-dependent. E.g. If a linux machine: syscall.SOL_IP
	originalAddrBytes, err := syscall.GetsockoptIPv6Mreq(fd, syscall.IPPROTO_IP, SO_ORIGINAL_DST)
	if err != nil {
		log.Println("[DEBUG] Getsockopt failed.")
		sourceConn.Close()
		return err
	}
	destConn, err := net.Dial("tcp", byteToConnString(originalAddrBytes.Multiaddr))
	if err != nil {
		log.Printf("[ERR] Unable to connect to destination. Closing connection %v.\n", id)
		sourceConn.Close()
		return err
	}
	t.id = id
	t.source = sourceConn
	t.destination = destConn
	return nil
}

//TLSPipe implements the TrudyPipe interface. TLSPipe is used to transparently proxy TLS traffic. Trudy
//currently just accepts TLS connections and poses as the destination. Obviously, TLS should stop this,
//so a reasonable well-designed client should _not_ allow this. But sometimes that is possible.
type TLSPipe struct {
	id          uint
	destination net.Conn
	source      net.Conn
}

//New creates a new TLSPipe.
func (t *TLSPipe) New(id uint, fd int, sourceConn net.Conn) (err error) {
	//TODO: Make the second argument system-dependent. E.g. If a linux machine: syscall.SOL_IP
	originalAddrBytes, err := syscall.GetsockoptIPv6Mreq(fd, syscall.IPPROTO_IP, SO_ORIGINAL_DST)
	if err != nil {
		log.Println("[DEBUG] Getsockopt failed.")
		sourceConn.Close()
		return err
	}
	tlsconfig := &tls.Config{InsecureSkipVerify: true}
	destConn, err := tls.Dial("tcp", byteToConnString(originalAddrBytes.Multiaddr), tlsconfig)
	if err != nil {
		log.Printf("[ERR] Unable to connect to destination. Closing connection %v.\n", id)
		sourceConn.Close()
		return err
	}
	t.id = id
	t.source = sourceConn
	t.destination = destConn
	return nil
}

//DestinationInfo returns the net.Addr of the destination.
func (t *TLSPipe) DestinationInfo() (addr net.Addr) {
	addr = t.destination.RemoteAddr()
	return
}

func (t *TLSPipe) SourceInfo() (addr net.Addr) {
	addr = t.source.RemoteAddr()
	return
}

//Close closes both ends of a TLSPipe.
func (t *TLSPipe) Close() {
	log.Printf("[INFO] ( %v ) Closing TLS connection.", t.id)
	t.source.Close()
	t.destination.Close()
}

//ReadSource reads data from the source end of the pipe. This is typically the proxy-unaware client.
func (t *TLSPipe) ReadSource(buffer []byte) (n int, err error) {
	return t.source.Read(buffer)
}

//WriteSource writes data to the source end of the pipe. This is typically the proxy-unaware client.
func (t *TLSPipe) WriteSource(buffer []byte) (n int, err error) {
	return t.source.Write(buffer)
}

//ReadDestination reads data from the destination end of the pipe. The destination is the proxy-unaware
//client's intended destination.
func (t *TLSPipe) ReadDestination(buffer []byte) (n int, err error) {
	return t.destination.Read(buffer)
}

//WriteDestination writes data to the destination end of the pipe. The destination is the proxy-unaware
//client's intended destination.
func (t *TLSPipe) WriteDestination(buffer []byte) (n int, err error) {
	return t.destination.Write(buffer)
}

type UDPPipe struct {
	id          uint
	destination net.Conn
	source      net.Conn
}

func (u *UDPPipe) DestinationInfo() (addr net.Addr) {
	addr = u.destination.RemoteAddr()
	return
}

func (u *UDPPipe) SourceInfo() (addr net.Addr) {
	addr = u.source.RemoteAddr()
	return
}

func (u *UDPPipe) Close() {
	u.source.Close()
	u.destination.Close()
}

func (u *UDPPipe) ReadSource(buffer []byte) (n int, err error) {
	u.source.SetReadDeadline(time.Now().Add(15 * time.Second))
	return u.source.Read(buffer)
}

func (u *UDPPipe) WriteSource(buffer []byte) (n int, err error) {
	u.source.SetWriteDeadline(time.Now().Add(15 * time.Second))
	return u.source.Write(buffer)
}

func (u *UDPPipe) ReadDestination(buffer []byte) (n int, err error) {
	u.destination.SetReadDeadline(time.Now().Add(15 * time.Second))
	return u.destination.Read(buffer)
}

func (u *UDPPipe) WriteDestination(buffer []byte) (n int, err error) {
	u.destination.SetWriteDeadline(time.Now().Add(15 * time.Second))
	return u.destination.Write(buffer)
}

func (u *UDPPipe) New(id uint, fd int, sourceConn net.Conn) (err error) {
	//TODO: Make the second argument system-dependent. E.g. If a linux machine: syscall.SOL_IP
	originalAddrBytes, err := syscall.GetsockoptIPv6Mreq(fd, syscall.IPPROTO_IP, SO_ORIGINAL_DST)
	if err != nil {
		log.Println("[DEBUG] Getsockopt failed.")
		sourceConn.Close()
		return err
	}
	destConn, err := net.Dial("udp", byteToConnString(originalAddrBytes.Multiaddr))
	if err != nil {
		log.Printf("[ERR] Unable to connect to destination. Closing connection %v.\n", id)
		sourceConn.Close()
		return err
	}
	u.id = id
	u.source = sourceConn
	u.destination = destConn
	return nil
}
