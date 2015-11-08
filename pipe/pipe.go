package pipe

import (
    "net"
    "log"
    "syscall"
    "strconv"
    "crypto/tls"
)

//Netfilter/iptables adds a tcp header to identify original destination. 
//Since all traffic is routed through trudy, we need to retrieve the original 
//intended destination (i.e. _not_ trudy)
const SO_ORIGINAL_DST = 80

//TODO: This makes sense but not currently. The way connectinos are currently dispatched don't cooperate with an interface without reflection.
type TrudyPipe interface {
    ReadSource(buffer []byte)       (n int, err error)
    WriteSource(buffer []byte)      (n int, err error)
    ReadDestination(buffer []byte)  (n int, err error)
    WriteDestination(buffer []byte) (n int, err error)
    New(uint, int, net.Conn)    (err error)
    Close()
}

type TCPPipe struct {
    id uint
    destination net.Conn
    source net.Conn
}

func (t *TCPPipe) Id() uint {
    return t.id
}

func (t *TCPPipe) Close() {
    log.Printf("[INFO] ( %v ) Closing TCP connection.")
    t.source.Close()
    t.destination.Close()
}

func (t *TCPPipe) ReadSource(buffer []byte) (n int, err error) {
    return t.source.Read(buffer)
}

func (t *TCPPipe) WriteSource(buffer []byte) (n int, err error) {
    return t.source.Write(buffer)
}

func (t *TCPPipe) ReadDestination(buffer []byte) (n int, err error) {
    return t.destination.Read(buffer)
}

func (t *TCPPipe) WriteDestination(buffer []byte) (n int, err error) {
    return t.destination.Write(buffer)
}

//Converts the Multiaddr bytestring returned by Getsockopt into a "host:port" connection string.
func ByteToConnString(multiaddr [16]byte) string {
    ip := multiaddr[4:8]
    ip_string := net.IPv4(ip[0], ip[1], ip[2], ip[3]).String()
    port := multiaddr[2:4]
    port_uint := int64((uint32(port[0]) << 8) + uint32(port[1]))
    port_string := strconv.FormatInt(port_uint,10)
    return (ip_string + ":" + port_string)
}

func (tcppipe *TCPPipe) New(id uint, fd int, sourceConn net.Conn) (err error) {
    //TODO: Make the second argument system-dependent. E.g. If a linux machine: syscall.SOL_IP
    originalAddrBytes,err := syscall.GetsockoptIPv6Mreq(fd, syscall.IPPROTO_IP, SO_ORIGINAL_DST)
    if err != nil {
        log.Println("[DEBUG] Getsockopt failed.")
        log.Println(err)
        sourceConn.Close()
        return err
    }
    destConn,err := net.Dial("tcp", ByteToConnString(originalAddrBytes.Multiaddr))
    if err != nil {
        log.Printf("[ERR] Unable to connect to destination. Closing connection %v.\n", id)
        sourceConn.Close()
        return err
    }
    tcppipe = &TCPPipe{id : id, source: sourceConn, destination: destConn}
    tcppipe.id = id
    tcppipe.source = sourceConn
    tcppipe.destination = destConn
    return nil
}

type TLSPipe struct {
    id uint
    destination net.Conn
    source net.Conn
}

func (tlspipe *TLSPipe) New(id uint, fd int, sourceConn net.Conn) (err error) {
    //TODO: Make the second argument system-dependent. E.g. If a linux machine: syscall.SOL_IP
    originalAddrBytes,err := syscall.GetsockoptIPv6Mreq(fd, syscall.IPPROTO_IP, SO_ORIGINAL_DST)
    if err != nil {
        log.Println("[DEBUG] Getsockopt failed.")
        log.Println(err)
        sourceConn.Close()
        return err
    }
    tlsconfig := &tls.Config { InsecureSkipVerify: true }
    destConn,err := tls.Dial("tcp", ByteToConnString(originalAddrBytes.Multiaddr), tlsconfig)
    if err != nil {
        log.Printf("[ERR] Unable to connect to destination. Closing connection %v.\n", id)
        sourceConn.Close()
        return err
    }
    tlspipe.id = id
    tlspipe.source = sourceConn
    tlspipe.destination = destConn
    return nil
}

func (t *TLSPipe) Close() {
    log.Printf("[INFO] ( %v ) Closing TLS connection.")
    t.source.Close()
    t.destination.Close()
}

func (t *TLSPipe) ReadSource(buffer []byte) (n int, err error) {
    return t.source.Read(buffer)
}

func (t *TLSPipe) WriteSource(buffer []byte) (n int, err error) {
    return t.source.Write(buffer)
}

func (t *TLSPipe) ReadDestination(buffer []byte) (n int, err error) {
    return t.destination.Read(buffer)
}

func (t *TLSPipe) WriteDestination(buffer []byte) (n int, err error) {
    return t.destination.Write(buffer)
}
