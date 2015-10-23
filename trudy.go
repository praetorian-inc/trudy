package main

import (
    "net"
    "log"
    "encoding/hex"
)

func main(){
    ip,_ := net.ResolveIPAddr("ip4", "0.0.0.0")
    ip_conn, err := net.ListenIP("ip4:icmp", ip)
    if err != nil {
        log.Println("[FATAL] Failed to setup listener. Did you run this as root? (You should!)")
        panic(err)
    }
    for {
        packet_buffer := make([]byte, 65535)
        n,dest_addr,err := ip_conn.ReadFrom(packet_buffer)
        if err != nil {
            log.Println("[ERR] Error reading from connection. Moving along.")
            continue
        }
        log.Printf("[DEBUG] %v\n", dest_addr)
        log.Printf("[INFO] Packet arrived!\n")
        go packetHandle(packet_buffer[:n])
    }
}

func packetHandle(packet []byte) {
    log.Printf("[DEBUG] Packet bytes\n%v", hex.Dump(packet))
}
