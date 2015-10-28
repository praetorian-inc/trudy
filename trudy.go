package main

import (
    "net"
    "log"
    "encoding/hex"
    "syscall"
    "strconv"
)

const SO_ORIGINAL_DST = 80
var connection_count uint

type TCPPipe struct {
    id uint
    destination net.Conn
    //TODO: Specifying TCPConn was arbitrary. Replace with just net.Conn struct.
    source net.TCPConn
}

func (t *TCPPipe) Id() uint {
    return t.id
}

func (t *TCPPipe) Close() {
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

func ByteToConnString(multiaddr [16]byte) string {
    ip := multiaddr[4:8]
    ip_string := net.IPv4(ip[0], ip[1], ip[2], ip[3]).String()
    port := multiaddr[2:4]
    port_uint := int64((uint32(port[0]) << 8) + uint32(port[1]))
    port_string := strconv.FormatInt(port_uint,10)
    log.Printf("[DEBUG] PORT STRING: %v\n", port_string)
    concat := ip_string + ":" + port_string
    log.Printf("[DEBUG] Connection to destination: %v\n", concat)
    return concat
}

//TODO: Effective Go would suggest removing the "new" and just naming this function TCPPipe.
func newTCPPipe(id uint, sourceConn net.TCPConn) TCPPipe {
    f, err := sourceConn.File()
    if err != nil {
        log.Printf("[ERR] Failed to read connection file descriptor")
    }
    //TODO: Investigate this more. This seems arbitrary. If a linux machine: syscall.SOL_IP
    originalAddrBytes,err := syscall.GetsockoptIPv6Mreq(int(f.Fd()), syscall.IPPROTO_IP, SO_ORIGINAL_DST)
    if err != nil {
        log.Println("[ERR] Getsockopt failed. Error below:")
        log.Printf("\t%v\n",err)
    }

    destConn,err := net.Dial("tcp", ByteToConnString(originalAddrBytes.Multiaddr))
    if err != nil {
        log.Printf("[ERR] Unable to connect to destination. Closing connection %v.\n", id)
        //TODO: Close connection. Also, this function should return an err value.
    }
    tcppipe := TCPPipe{id : id, source: sourceConn, destination: destConn}
    return tcppipe
}

func main(){
    tcpAddr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:6666")
    tcpListener,err := net.ListenTCP("tcp", tcpAddr)

    defer tcpListener.Close()

    if err != nil {
        log.Println("[FATAL] Failed to setup listeners. Did you run this as root? (You should!)")
        panic(err)
    }

    log.Println("[INFO] Trudy lives!")

    for {
        conn, err := tcpListener.AcceptTCP()
        if err != nil {
            log.Println("[ERR] Error reading from connection. Moving along.")
            continue
        }
        tcppipe := newTCPPipe(connection_count, *conn)
        log.Printf("[INFO] TCP Connection %v accepted!\n", connection_count)
        go tcpConnectionHandler(tcppipe)
        connection_count++
    }
}

func tcpConnectionHandler(tcppipe TCPPipe) {
    defer log.Printf("[INFO] Connection %v closed!\n", tcppipe.Id())
    defer tcppipe.Close()
    buffer := make([]byte, 65535)
    //TODO: A connection should run data through at most 2 functions. A filter (for packets that don't meet some criteria), and a modifier.
    for {
        n, err := tcppipe.ReadSource(buffer)
        if err != nil {
            break
        }
        if filter(buffer) {
            buffer = mangle(buffer)
        }
        n, err = tcppipe.WriteDestination(buffer)
        if err != nil {
            log.Println("[ERR] Unable to send data to destination. Error: ")
            log.Printf("\t%v\n", err)
        }
        //TODO: This loop should be more sophisticated. Src/Dst Read/Write should be concurrent.
        log.Printf("[DEBUG] Packet send bytes\n%v", hex.Dump(buffer[:n]))
        n, err = tcppipe.ReadDestination(buffer)
        if err != nil {
            log.Printf("[ERR] Unable to read data from destination. Bailing")
            break
        }
        log.Printf("[DEBUG] Packet recieve bytes\n%v", hex.Dump(buffer[:n]))
    }
}

func filter(input []byte) bool {
    return true
}

func mangle(input []byte) []byte {
    return input
}
