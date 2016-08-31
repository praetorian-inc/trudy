package pipe

import (
	"crypto/tls"
	"log"
	"net"
	"strconv"
	"sync"
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
	ReadFromClient(buffer []byte) (n int, err error)
	WriteToClient(buffer []byte) (n int, err error)
	ReadFromServer(buffer []byte) (n int, err error)
	WriteToServer(buffer []byte) (n int, err error)
	ServerInfo() (addr net.Addr)
	ClientInfo() (addr net.Addr)
	ServerConn() net.Conn
	ClientConn() net.Conn
	SetServerConn(net.Conn)
	SetClientConn(net.Conn)
	New(uint, int, net.Conn) (err error)
	Close()
	Id() uint
	Lock()
	Unlock()
	AddContext(key string, value interface{})
	GetContext(key string) (value interface{}, ok bool)
	DeleteContext(key string)
}

//TCPPipe implements the TrudyPipe interface and can be used to proxy generic TCP connections.
type TCPPipe struct {
	id         uint
	serverConn net.Conn
	clientConn net.Conn
	pipeMutex  *sync.Mutex
	userMutex  *sync.Mutex
	KV         map[string]interface{}
}

func (t TCPPipe) Lock() {
	t.userMutex.Lock()
}

func (t TCPPipe) Unlock() {
	t.userMutex.Unlock()
}

func (t *TCPPipe) AddContext(key string, value interface{}) {
	t.pipeMutex.Lock()
	t.KV[key] = value
	t.pipeMutex.Unlock()
}

func (t *TCPPipe) GetContext(key string) (retval interface{}, ok bool) {
	retval, ok = t.KV[key]
	return
}

func (t *TCPPipe) DeleteContext(key string) {
	t.pipeMutex.Lock()
	delete(t.KV, key)
	t.pipeMutex.Unlock()
}

func (t TCPPipe) ClientConn() net.Conn {
	return t.clientConn
}

func (t TCPPipe) ServerConn() net.Conn {
	return t.serverConn
}

func (t *TCPPipe) SetClientConn(c net.Conn) {
	t.pipeMutex.Lock()
	t.clientConn = c
	t.pipeMutex.Unlock()
}

func (t *TCPPipe) SetServerConn(s net.Conn) {
	t.pipeMutex.Lock()
	t.serverConn = s
	t.pipeMutex.Unlock()
}

//Id returns a TCPPipe identifier
func (t TCPPipe) Id() uint {
	return t.id
}

//ServerInfo returns the net.Addr of the server.
func (t TCPPipe) ServerInfo() (addr net.Addr) {
	addr = t.serverConn.RemoteAddr()
	return
}

//ClientInfo returns the net.Addr of the client.
func (t TCPPipe) ClientInfo() (addr net.Addr) {
	addr = t.clientConn.RemoteAddr()
	return
}

//Close closes both ends of a TCPPipe.
func (t *TCPPipe) Close() {
	t.serverConn.Close()
	t.clientConn.Close()
}

//ReadFromClient reads data from the client end of the pipe. This is typically the proxy-unaware client.
func (t *TCPPipe) ReadFromClient(buffer []byte) (n int, err error) {
	//TODO(kkl): Make timeouts configureable.
	t.clientConn.SetReadDeadline(time.Now().Add(15 * time.Second))
	return t.clientConn.Read(buffer)
}

//WriteToClient writes data to the client end of the pipe. This is typically the proxy-unaware client.
func (t *TCPPipe) WriteToClient(buffer []byte) (n int, err error) {
	//TODO(kkl): Make timeouts configureable.
	t.clientConn.SetWriteDeadline(time.Now().Add(15 * time.Second))
	return t.clientConn.Write(buffer)
}

//ReadFromServer reads data from the server end of the pipe. The server is the proxy-unaware
//client's intended destination.
func (t *TCPPipe) ReadFromServer(buffer []byte) (n int, err error) {
	t.serverConn.SetReadDeadline(time.Now().Add(15 * time.Second))
	return t.serverConn.Read(buffer)
}

//WriteToServer writes data to the server end of the pipe. The server is the proxy-unaware
//client's intended destination.
func (t *TCPPipe) WriteToServer(buffer []byte) (n int, err error) {
	t.serverConn.SetWriteDeadline(time.Now().Add(15 * time.Second))
	return t.serverConn.Write(buffer)
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

//New builds a new TCPPipe.
func (t *TCPPipe) New(id uint, fd int, clientConn net.Conn) (err error) {
	//TODO(kkl): Make the second argument system-dependent. E.g. If a linux machine: syscall.SOL_IP
	originalAddrBytes, err := syscall.GetsockoptIPv6Mreq(fd, syscall.IPPROTO_IP, SO_ORIGINAL_DST)
	if err != nil {
		log.Println("[DEBUG] Getsockopt failed.")
		clientConn.Close()
		return err
	}
	serverConn, err := net.Dial("tcp", byteToConnString(originalAddrBytes.Multiaddr))
	if err != nil {
		log.Printf("[ERR] Unable to connect to destination. Closing pipe.\n", id)
		clientConn.Close()
		return err
	}
	t.id = id
	t.clientConn = clientConn
	t.serverConn = serverConn
	t.pipeMutex = new(sync.Mutex)
	t.userMutex = new(sync.Mutex)
	return nil
}

//TLSPipe implements the TrudyPipe interface. TLSPipe is used to transparently proxy TLS traffic. Trudy
//currently just accepts TLS connections and poses as the destination. Obviously, TLS should stop this,
//so a reasonable well-designed client should _not_ allow this. But sometimes that is possible.
type TLSPipe struct {
	id         uint
	serverConn net.Conn
	clientConn net.Conn
	pipeMutex  *sync.Mutex
	userMutex  *sync.Mutex
	KV         map[string]interface{}
}

func (t TLSPipe) Lock() {
	t.userMutex.Lock()
}

func (t TLSPipe) Unlock() {
	t.userMutex.Unlock()
}

func (t *TLSPipe) AddContext(key string, value interface{}) {
	t.pipeMutex.Lock()
	t.KV[key] = value
	t.pipeMutex.Unlock()
}

func (t *TLSPipe) GetContext(key string) (retval interface{}, ok bool) {
	retval, ok = t.KV[key]
	return
}

func (t *TLSPipe) DeleteContext(key string) {
	t.pipeMutex.Lock()
	delete(t.KV, key)
	t.pipeMutex.Unlock()
}

func (t TLSPipe) ClientConn() net.Conn {
	return t.clientConn
}

func (t TLSPipe) ServerConn() net.Conn {
	return t.serverConn
}

func (t *TLSPipe) SetClientConn(c net.Conn) {
	t.pipeMutex.Lock()
	t.clientConn = c
	t.pipeMutex.Unlock()
}

func (t *TLSPipe) SetServerConn(s net.Conn) {
	t.pipeMutex.Lock()
	t.serverConn = s
	t.pipeMutex.Unlock()
}

//Id returns a TLSPipe identifier
func (t TLSPipe) Id() uint {
	return t.id
}

//New creates a new TLSPipe.
func (t *TLSPipe) New(id uint, fd int, clientConn net.Conn) (err error) {
	//TODO: Make the second argument system-dependent. E.g. If a linux machine: syscall.SOL_IP
	originalAddrBytes, err := syscall.GetsockoptIPv6Mreq(fd, syscall.IPPROTO_IP, SO_ORIGINAL_DST)
	if err != nil {
		log.Println("[DEBUG] Getsockopt failed.")
		clientConn.Close()
		return err
	}
	tlsconfig := &tls.Config{InsecureSkipVerify: true}
	serverConn, err := tls.Dial("tcp", byteToConnString(originalAddrBytes.Multiaddr), tlsconfig)
	if err != nil {
		log.Printf("[ERR] Unable to connect to destination. Closing connection %v.\n", id)
		clientConn.Close()
		return err
	}
	t.id = id
	t.clientConn = clientConn
	t.serverConn = serverConn
	t.pipeMutex = new(sync.Mutex)
	t.userMutex = new(sync.Mutex)
	return nil
}

//ServerInfo returns the net.Addr of the server.
func (t TLSPipe) ServerInfo() (addr net.Addr) {
	addr = t.serverConn.RemoteAddr()
	return
}

//ClientInfo returns the net.Addr of the client.
func (t TLSPipe) ClientInfo() (addr net.Addr) {
	addr = t.clientConn.RemoteAddr()
	return
}

//Close closes both ends of a TLSPipe.
func (t *TLSPipe) Close() {
	log.Printf("[INFO] ( %v ) Closing TLS connection.", t.id)
	t.clientConn.Close()
	t.serverConn.Close()
}

//ReadFromClient reads data from the source end of the pipe.
func (t *TLSPipe) ReadFromClient(buffer []byte) (n int, err error) {
	return t.clientConn.Read(buffer)
}

//WriteToClient writes data to the client end of the pipe.
func (t *TLSPipe) WriteToClient(buffer []byte) (n int, err error) {
	return t.clientConn.Write(buffer)
}

//ReadFromServer reads data from the server end of the pipe.
func (t *TLSPipe) ReadFromServer(buffer []byte) (n int, err error) {
	return t.serverConn.Read(buffer)
}

//WriteToServer writes data to the server end of the pipe.
func (t *TLSPipe) WriteToServer(buffer []byte) (n int, err error) {
	return t.serverConn.Write(buffer)
}
