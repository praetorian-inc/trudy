package main

import (
    "net"
    "log"
    //"crypto/tls"
    "github.com/kelbyludwig/trudy/pipe"
    "github.com/kelbyludwig/trudy/module"
)

var connection_count uint

func main(){
    tcpAddr,err := net.ResolveTCPAddr("tcp", ":6443")
    errHandler(err)
    tcpListener,err := net.ListenTCP("tcp", tcpAddr)
    errHandler(err)

    //cert,err := tls.LoadX509KeyPair("./certificate/trudy.crt", "./certificate/trudy.key")
    //errHandler(err)
    //config := tls.Config{Certificates: []tls.Certificate{cert}}
    //tlsListener,err := tls.Listen("tcp", ":6443", &config)
    //errHandler(err)

    log.Println("[INFO] Trudy lives!")

    //go ConnectionDispatcher(tlsListener, "TLS")
    ConnectionDispatcher(tcpListener, "TCP")
}

func ConnectionDispatcher(listener *net.TCPListener, name string) {
    defer listener.Close()
    for {
        conn, err := listener.AcceptTCP()
        if err != nil {
            log.Printf("[ERR] Error accepting %v connection. Moving along.", name)
            continue
        }
        tcppipe,err := pipe.NewTCPPipe(connection_count, *conn)
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
