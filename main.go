package main

import (
    "net"
    "log"
    "crypto/tls"
    "encoding/hex"
    "github.com/kelbyludwig/trudy/pipe"
    "github.com/kelbyludwig/trudy/module"
)

var connection_count uint

func main(){
    tcpAddr,_ := net.ResolveTCPAddr("tcp", "0.0.0.0:6666")
    tcpListener,err := net.ListenTCP("tcp", tcpAddr)
    errHandler(err)

    udpAddr,_ := net.ResolveUDPAddr("udp", "0.0.0.0:6667")
    udpListener,err := net.ListenUDP("udp", udpAddr)
    errHandler(err)

    cert,err := tls.LoadX509KeyPair("./certificate/trudy.crt", "./certificate/trudy.key")
    errHandler(err)
    config := tls.Config{Certificates: []tls.Certificate{cert}}
    tlsListener,err := tls.Listen("tcp", "0.0.0.0:6443", &config)
    errHandler(err)

    defer tcpListener.Close()
    defer udpListener.Close()
    defer tlsListener.Close()

    log.Println("[INFO] Trudy lives!")

    go tlsHandler(tlsListener)

    for {
        conn, err := tcpListener.AcceptTCP()
        if err != nil {
            log.Println("[ERR] Error accepting TCP connection. Moving along.")
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

func tlsHandler(tlsListener net.Listener) {
    for {
        conn, err := tlsListener.Accept()
        if err != nil {
            log.Println("[ERR] Error accepting TLS connection. Moving along.")
            continue
        }
        buffer := make([]byte, 65535)
        n,_ := conn.Read(buffer)
        log.Printf("[DEBUG] Read from connection: \n%v\n", hex.Dump(buffer[:n]))
    }
}

func errHandler(err error) {
    if err != nil {
        panic(err)
    }
}

func clientHandler(tcppipe pipe.TCPPipe) {
    defer tcppipe.Close()

    buffer := make([]byte, 65535)

    //TODO: Timeouts!
    for {
        bytesRead,err := tcppipe.ReadSource(buffer)
        if err != nil {
            continue
        }

        if module.Drop(buffer[:bytesRead]){
            continue
        }

        if !module.Pass(buffer[:bytesRead]) {
            //TODO: This won't work when Mangle returns a different sized buffer.
            buffer = module.Mangle(buffer)
        }

        log.Println(module.PrettyPrint(buffer))

        _, err = tcppipe.WriteDestination(buffer[:bytesRead])
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
