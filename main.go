package main

import (
    "net"
    "log"
    "github.com/kelbyludwig/trudy/pipe"
)

var connection_count uint

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
        tcppipe,err := pipe.NewTCPPipe(connection_count, *conn)
        if err != nil {
            log.Println("[ERR] Error creating new TCPPipe.")
            continue
        }
        log.Printf("[INFO] TCP Connection %v accepted!\n", connection_count)
        go clientHandler(tcppipe)
        go serverHandler(tcppipe)
        connection_count++
    }
}

func clientHandler(tcppipe pipe.TCPPipe) {
    defer tcppipe.Close()

    buffer := make([]byte, 65535)

    //TODO: Timeouts!
    for {
        bytesReadFromSource,err := tcppipe.ReadSource(buffer)
        if err != nil {
            continue
        }

        //if filter(buffer[:bytesReadFromSource]) {
        //    buffer = mangle(buffer[:bytesReadFromSource])
        //}

        _, err = tcppipe.WriteDestination(buffer[:bytesReadFromSource])
        if err != nil {
            continue
        }
    }
}

func serverHandler(tcppipe pipe.TCPPipe) {
    defer tcppipe.Close()

    buffer := make([]byte, 65535)

    //TODO: Timeouts!
    for {
        bytesReadFromDestination,err := tcppipe.ReadDestination(buffer)
        if err != nil {
            continue
        }
        _,err = tcppipe.WriteSource(buffer[:bytesReadFromDestination])
        if err != nil {
            continue
        }
    }
}

func pass(input []byte) bool {
    return true
}

func mangle(input []byte) []byte {
    return input
}

func drop(input []byte) bool {
    return false
}
