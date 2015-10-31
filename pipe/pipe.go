package pipe

import (
    "net"
    "log"
    "syscall"
    "strconv"
)

//Netfilter/iptables adds a tcp header to identify original destination. 
//Since all traffic is routed through trudy, we need to retrieve the original 
//intended destination (i.e. _not_ trudy)
const SO_ORIGINAL_DST = 80

type TCPPipe struct {
    id uint
    destination net.Conn
    source net.TCPConn
}

func (t *TCPPipe) Id() uint {
    return t.id
}

func (t *TCPPipe) Close() {
    log.Printf("[INFO] ( %v ) Closing connection.")
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

func NewTCPPipe(id uint, sourceConn net.TCPConn) (pipe TCPPipe, err error) {
    tcppipe := new(TCPPipe)
    f, err := sourceConn.File()
    if err != nil {
        log.Println("[DEBUG] Failed to read connection file descriptor.")
        sourceConn.Close()
        return *tcppipe, err
    }
    //TODO: Make the second argument system-dependent. E.g. If a linux machine: syscall.SOL_IP
    originalAddrBytes,err := syscall.GetsockoptIPv6Mreq(int(f.Fd()), syscall.IPPROTO_IP, SO_ORIGINAL_DST)
    if err != nil {
        log.Println("[DEBUG] Getsockopt failed.")
        sourceConn.Close()
        return *tcppipe, err
    }
    destConn,err := net.Dial("tcp", ByteToConnString(originalAddrBytes.Multiaddr))
    if err != nil {
        log.Printf("[ERR] Unable to connect to destination. Closing connection %v.\n", id)
        sourceConn.Close()
        return *tcppipe, err
    }
    tcppipe = &TCPPipe{id : id, source: sourceConn, destination: destConn}
    return *tcppipe, nil
}

