package main

import (
    "net"
    "log"
    "crypto/tls"
    "github.com/kelbyludwig/trudy/pipe"
    "github.com/kelbyludwig/trudy/module"
    "github.com/kelbyludwig/trudy/listener"
)

var connection_count uint

func main(){
    tcpAddr,_ := net.ResolveTCPAddr("tcp", ":6666")
    tcpListener := new(listener.TCPListener)
    tcpListener.Listen("tcp", tcpAddr, tls.Config{})

    cert,_ := tls.LoadX509KeyPair("./certificate/trudy.crt", "./certificate/trudy.key")
    config := tls.Config{Certificates: []tls.Certificate{cert}}
    tlsAddr,_ := net.ResolveTCPAddr("tcp", ":6443")
    tlsListener := new(listener.TLSListener)
    tlsListener.Listen("tcp", tlsAddr, config)

    log.Println("[INFO] Trudy lives!")

    go ConnectionDispatcher(tlsListener, "TLS")
    ConnectionDispatcher(tcpListener, "TCP")
}

func ConnectionDispatcher(listener listener.TrudyListener, name string) {
    defer listener.Close()
    for {
        fd, conn, err := listener.Accept()
        if err != nil {
            log.Printf("[ERR] Error accepting %v connection. Moving along.", name)
            continue
        }
        tcppipe,err := pipe.NewTCPPipe(connection_count, fd, conn)
        if err != nil {
            log.Println("[ERR] Error creating new TCPPipe.")
            continue
        }
        log.Printf("[INFO] %v Connection %v accepted!\n", name, connection_count)
        go clientHandler(tcppipe)
        go serverHandler(tcppipe)
        connection_count++
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

        //log.Println(module.PrettyPrint(buffer))

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
