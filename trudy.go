package main

import (
    "bufio"
    "net"
    "log"
    "encoding/hex"
)

var connection_counter uint

func main(){
    tcp_recvr, err := net.Listen("tcp", ":6666")
    if err != nil {
        log.Fatal("Failed to setup TCP listener!")
    }
    for {
        tcp_conn, err := tcp_recvr.Accept()
        if err != nil {
            log.Println("[ERR] Error accepting a connection. Moving along.")
            continue
        }
        connection_counter++
        log.Printf("[INFO] Connection %v accepted!\n", connection_counter)
        go tcpHandle(tcp_conn, connection_counter)
    }
}

func tcpHandle(conn net.Conn, ctr uint) {
    defer conn.Close()
    reader := bufio.NewReader(conn)
    buffer := make([]byte, 1024)
    for {
        n,err := reader.Read(buffer)
        if err != nil {
            log.Printf("[INFO] Terminating connection %v!\n", ctr)
            break
        }
        log.Printf("[DEBUG]: Input hexdump from connection %v\n%v", ctr, hex.Dump(buffer[:n]))
    }
}
