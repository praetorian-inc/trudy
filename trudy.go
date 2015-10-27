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
    original_destination_ip [16]byte
    connection net.TCPConn
}

func (t *TCPPipe) Id() uint {
    return t.id
}

func (t *TCPPipe) Close() {
    t.connection.Close()
}

func (t *TCPPipe) Read(buffer []byte) (n int, err error) {
    return t.connection.Read(buffer)
}

func newTCPPipe(id uint, conn net.TCPConn) TCPPipe {
    f, err := conn.File()
    if err != nil {
        log.Printf("[ERR] Failed to read connection file descriptor")
    }
    fd := f.Fd()
    log.Printf("[DEBUG] FD: %v\n", fd)
    //TODO: Investigate this more. This seems arbitrary. If a linux machine: syscall.SOL_IP
    original,err := syscall.GetsockoptIPv6Mreq(int(fd), syscall.IPPROTO_IP, SO_ORIGINAL_DST)
    log.Printf("[DEBUG] ORIGINAL %v\n", original.Multiaddr)
    tcppipe := TCPPipe{id : id, connection: conn, original_destination_ip: original.Multiaddr}
    if err != nil {
        log.Println("[ERR] Getting sockoption failed.")
        log.Println(err)
    }
    return tcppipe
}

func main(){
    addr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:6666")
    listener,err := net.ListenTCP("tcp", addr)

    defer listener.Close()

    if err != nil {
        log.Println("[FATAl] Failed to setup listener. Did you run this as root? (You should!)")
        panic(err)
    }

    log.Println("[INFO] Trudy lives!")
    for {
        conn, err := listener.AcceptTCP()
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
    for {
        n, err := tcppipe.Read(buffer)
        if err != nil {
            break
        }
        log.Printf("[DEBUG] Packet bytes\n%v", hex.Dump(buffer[:n]))
    }
}
