package main

import (
    "net"
    "log"
    "encoding/hex"
    "syscall"
)

const SO_ORIGINAL_DST = 80
var connection_count uint

type TCPPipe struct {
    id uint
    //TODO: destination_ip is a temporary struct. Using connections is all that is necessary.
    destination_ip [16]byte
    destination net.TCPConn
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

func newTCPPipe(id uint, conn net.TCPConn) TCPPipe {
    f, err := conn.File()
    if err != nil {
        log.Printf("[ERR] Failed to read connection file descriptor")
    }
    //TODO: Investigate this more. This seems arbitrary. If a linux machine: syscall.SOL_IP
    original_destination,err := syscall.GetsockoptIPv6Mreq(int(f.Fd()), syscall.IPPROTO_IP, SO_ORIGINAL_DST)
    if err != nil {
        log.Println("[ERR] Getsockopt failed. Error below:")
        log.Printf("\t%v\n",err)
    }

    //TODO: Construct the destination end of the pipe here too.
    tcppipe := TCPPipe{id : id, source: conn, destination_ip: original_destination.Multiaddr}

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
        log.Printf("[DEBUG] Packet bytes\n%v", hex.Dump(buffer[:n]))
    }
}

func filter(input []byte) bool {
    return true
}

func mangle(input []byte) []byte {
    return input
}
