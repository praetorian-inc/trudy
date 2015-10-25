package main

import (
    "net"
    "log"
    "encoding/hex"
    "syscall"
)


const SO_ORIGINAL_DST = 80
var connection_count uint

//TODO: A connection-state holding struct will be nice :)

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
        log.Printf("[INFO] TCP Connection %v accepted!\n", connection_count)
        go connectionHandler(*conn, connection_count)
        connection_count++
    }
}

func connectionHandler(conn net.TCPConn, id uint) {
    defer log.Printf("[INFO] Connection %v closed!\n", id)
    defer conn.Close()
    f, err := conn.File()
    if err != nil {
        log.Printf("[ERR] Failed to read connection file descriptor")
    }
    fd := f.Fd()
    log.Printf("[DEBUG] FD: %v\n", fd)
    //TODO: Investigate this more. This seems arbitrary. If a linux machine: syscall.SOL_IP
    original,err := syscall.GetsockoptIPv6Mreq(int(fd), syscall.IPPROTO_IP, SO_ORIGINAL_DST)
    if err != nil {
        log.Println("[ERR] Getting sockoption failed.")
        log.Println(err)
    }
    log.Printf("[DEBUG] Original %v\n", original)
    buffer := make([]byte, 65535)
    for {
        n, err := conn.Read(buffer)
        if err != nil {
            break
        }
        log.Printf("[DEBUG] Packet bytes\n%v", hex.Dump(buffer[:n]))
    }
}
