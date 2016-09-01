//Package pipe defines the data structure used to manipulate, monitor, and create proxied connections.
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

//Pipe is the primary interface that handles connections. Pipe creates a
//full-duplex pipe that passes data from the client to the server and vice
//versa. A pipe is compromised of two connections. The client transparently
//connects to Trudy, and Trudy accepts the connection.  Trudy will then make a
//connection with the client's intended destination and just pass traffic
//back-and-forth between the two connections. All modifications and drops to
//the packet happen to data between the two ends of the pipe.
type Pipe interface {

	//Id returns a unique Pipe identifier
	Id() uint

	//ServerInfo returns the net.Addr of the server-end of the pipe.
	ServerInfo() (addr net.Addr)

	//ClientInfo returns the net.Addr of the client-end of the pipe.
	ClientInfo() (addr net.Addr)

	//ReadFromClient reads data into the buffer from the client-end of the
	//pipe. ReadFromClient returns the number of bytes read and an error
	//value if an error or EOF occurred. Note: ReadFromClient can read a
	//non-zero number of bytes and have a non-nil error value (e.g. EOF).
	ReadFromClient(buffer []byte) (n int, err error)

	//WriteToClient writes data to the client-end of the pipe. This is
	//typically the proxy-unaware client.
	WriteToClient(buffer []byte) (n int, err error)

	//ReadFromServer reads data into the buffer from the server-end of the
	//pipe. The server is the proxy-unaware client's intended destination.
	//ReadFromServer returns the number of bytes read and an error value if
	//an error or EOF occurred. Note: ReadFromServer can read a non-zero
	//number of bytes and have a non-nil error value (e.g. EOF).
	ReadFromServer(buffer []byte) (n int, err error)

	//WriteToServer writes buffer to the server-end of the pipe. The server
	//is the proxy-unaware client's intended destination.
	WriteToServer(buffer []byte) (n int, err error)

	//ServerConn returns the net.Conn responsible for server-end
	//communication.
	ServerConn() (conn net.Conn)

	//CilentConn returns the net.Conn responsible for client-end
	//communication.
	ClientConn() (conn net.Conn)

	//SetServerConn will replace the server-end of the pipe with the supplied
	//net.Conn parameter.
	SetServerConn(conn net.Conn)

	//SetClientConn will replace the client-end of the pipe with the supplied
	//net.Conn parameter.
	SetClientConn(conn net.Conn)

	//New builds a new Pipe.
	New(pipeID uint, clientConnFD int, clientConn net.Conn, useTLS bool) (err error)

	//Close closes both connections of the Pipe.
	Close()

	//Lock locks a per-Pipe mutex that can be used in modules for
	//synchronization.
	Lock()

	//Unlock unlocks a per-Pipe mutex that can be used in modules for
	//synchronization.
	Unlock()

	//AddContext adds a key/value pair to the Pipe.
	AddContext(key string, value interface{})

	//GetContext retrieves a value in a Pipe key/value data store.
	//GetContext returns the value and a bool indicating success.
	GetContext(key string) (value interface{}, ok bool)

	//DeleteContext removes a key/value pair from the Pipe.
	DeleteContext(key string)
}

//TODO(kkl): I don't think New needs to be part of the Pipe interface.
//Removing this very specific constructor will allow for other methods
//of getting trudy as a proxy (e.g. other transparent proxies, or
//non-transparent proxies like SOCKS).

//TrudyPipe implements the Pipe interface and can be used to proxy TCP connections.
type TrudyPipe struct {
	id         uint
	serverConn net.Conn
	clientConn net.Conn
	pipeMutex  *sync.Mutex
	userMutex  *sync.Mutex
	KV         map[string]interface{}
}

//Lock locks a mutex stored within TrudyPipe to allow for fine-grained
//synchronization within a module.
func (t *TrudyPipe) Lock() {
	t.userMutex.Lock()
}

//Unlock unlocks a mutex stored within TrudyPipe to allow for fine-grained
//synchronization within a module.
func (t *TrudyPipe) Unlock() {
	t.userMutex.Unlock()
}

//AddContext adds a key/value pair to the TrudyPipe. The key/value
//pair data store is per-TrudyPipe. AddContext is safe for use
//in multiple goroutines.
func (t *TrudyPipe) AddContext(key string, value interface{}) {
	t.pipeMutex.Lock()
	t.KV[key] = value
	t.pipeMutex.Unlock()
}

//GetContext retrieves a value in a TrudyPipe key/value data store.
//GetContext returns the value and a bool indicating success.
func (t *TrudyPipe) GetContext(key string) (retval interface{}, ok bool) {
	retval, ok = t.KV[key]
	return
}

//DeleteContext removes a key/value pair from the TrudyPipe. DeleteContext is
//safe for use in multiple goroutines.
func (t *TrudyPipe) DeleteContext(key string) {
	t.pipeMutex.Lock()
	delete(t.KV, key)
	t.pipeMutex.Unlock()
}

//CilentConn returns the net.Conn responsible for client-end communication.
func (t *TrudyPipe) ClientConn() net.Conn {
	return t.clientConn
}

//ServerConn returns the net.Conn responsible for server-end communication.
func (t *TrudyPipe) ServerConn() net.Conn {
	return t.serverConn
}

//SetClientConn will replace the client-end of the pipe with the supplied
//net.Conn parameter. SetClientConn is safe for use in multiple goroutines.
func (t *TrudyPipe) SetClientConn(c net.Conn) {
	t.pipeMutex.Lock()
	t.clientConn = c
	t.pipeMutex.Unlock()
}

//SetServerConn will replace the server-end of the pipe with the supplied
//net.Conn parameter. SetServerConn is safe for use in multiple goroutines.
func (t *TrudyPipe) SetServerConn(s net.Conn) {
	t.pipeMutex.Lock()
	t.serverConn = s
	t.pipeMutex.Unlock()
}

//Id returns a TrudyPipe identifier
func (t *TrudyPipe) Id() uint {
	return t.id
}

//ServerInfo returns the net.Addr of the server.
func (t *TrudyPipe) ServerInfo() (addr net.Addr) {
	addr = t.serverConn.RemoteAddr()
	return
}

//ClientInfo returns the net.Addr of the client.
func (t *TrudyPipe) ClientInfo() (addr net.Addr) {
	addr = t.clientConn.RemoteAddr()
	return
}

//Close closes both ends of a TrudyPipe.
func (t *TrudyPipe) Close() {
	t.serverConn.Close()
	t.clientConn.Close()
}

//ReadFromClient reads data from the client end of the pipe. This is typically the proxy-unaware client.
func (t *TrudyPipe) ReadFromClient(buffer []byte) (n int, err error) {
	//TODO(kkl): Make timeouts configureable.
	t.clientConn.SetReadDeadline(time.Now().Add(15 * time.Second))
	return t.clientConn.Read(buffer)
}

//WriteToClient writes data to the client end of the pipe. This is typically the proxy-unaware client.
func (t *TrudyPipe) WriteToClient(buffer []byte) (n int, err error) {
	//TODO(kkl): Make timeouts configureable.
	t.clientConn.SetWriteDeadline(time.Now().Add(15 * time.Second))
	return t.clientConn.Write(buffer)
}

//ReadFromServer reads data from the server end of the pipe. The server is the
//proxy-unaware client's intended destination.
func (t *TrudyPipe) ReadFromServer(buffer []byte) (n int, err error) {
	t.serverConn.SetReadDeadline(time.Now().Add(15 * time.Second))
	return t.serverConn.Read(buffer)
}

//WriteToServer writes data to the server end of the pipe. The server is the
//proxy-unaware client's intended destination.
func (t *TrudyPipe) WriteToServer(buffer []byte) (n int, err error) {
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

//New builds a new TrudyPipe. New will get the original destination of traffic
//that was mangled by iptables and get the original destination. New will then
//open a connection to that original destination and, upon success, will set
//all the internalf values needed for a TrudyPipe.
func (t *TrudyPipe) New(id uint, fd int, clientConn net.Conn, useTLS bool) (err error) {
	//TODO(kkl): Make the second argument system-dependent. E.g. If a linux machine: syscall.SOL_IP
	originalAddrBytes, err := syscall.GetsockoptIPv6Mreq(fd, syscall.IPPROTO_IP, SO_ORIGINAL_DST)
	if err != nil {
		log.Println("[DEBUG] Getsockopt failed.")
		clientConn.Close()
		return err
	}

	var serverConn net.Conn
	if useTLS {
		tlsconfig := &tls.Config{InsecureSkipVerify: true}
		serverConn, err = tls.Dial("tcp", byteToConnString(originalAddrBytes.Multiaddr), tlsconfig)
		if err != nil {
			log.Printf("[ERR] Unable to connect to destination. Closing connection %v.\n", id)
			clientConn.Close()
			return err
		}
	} else {
		serverConn, err = net.Dial("tcp", byteToConnString(originalAddrBytes.Multiaddr))
		if err != nil {
			log.Printf("[ERR] ( %v ) Unable to connect to destination. Closing pipe.\n", id)
			clientConn.Close()
			return err
		}
	}
	t.id = id
	t.clientConn = clientConn
	t.serverConn = serverConn
	t.pipeMutex = new(sync.Mutex)
	t.userMutex = new(sync.Mutex)
	return nil
}
