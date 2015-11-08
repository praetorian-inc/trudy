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
            continue
        }
        var p pipe.TrudyPipe
        if name == "TLS" {
            p = new(pipe.TLSPipe)
            err = p.New(connection_count, fd, conn)
        } else {
            p = new(pipe.TCPPipe)
            err = p.New(connection_count, fd, conn)
        }
        if err != nil {
            log.Println("[ERR] Error creating new pipe.")
            continue
        }
        log.Printf("[INFO] ( %v ) %v Connection accepted!\n", connection_count, name)
        go clientHandler(p)
        go serverHandler(p)
        connection_count++
    }
}

func errHandler(err error) {
    if err != nil {
        panic(err)
    }
}

func clientHandler(pipe pipe.TrudyPipe) {
    defer pipe.Close()

    buffer := make([]byte, 65535)

    //TODO: Timeouts!
    for {
        bytesRead,err := pipe.ReadSource(buffer)
        if err != nil {
            break
        }

        data := module.Data{FromClient: true, Bytes: buffer[:bytesRead]}
        //if module.Drop(buffer[:bytesRead]){
        //    continue
        //}

        //if !module.Pass(buffer[:bytesRead]) {
        //    //TODO: This won't work when Mangle returns a different sized buffer.
        //    buffer = module.Mangle(buffer)
        //}

        log.Printf("Client -> Server: \n%v\n", module.PrettyPrint(data))

        _, err = pipe.WriteDestination(data.Bytes)
        if err != nil {
            break
        }
    }
}

func serverHandler(pipe pipe.TrudyPipe) {
    defer pipe.Close()

    buffer := make([]byte, 65535)

    //TODO: Timeouts!
    for {
        bytesRead,err := pipe.ReadDestination(buffer)
        if err != nil {
            break
        }
        data := module.Data{FromClient: false, Bytes: buffer[:bytesRead]}
        log.Printf("Server -> Client: \n%v\n", module.PrettyPrint(data))
        _,err = pipe.WriteSource(data.Bytes)
        if err != nil {
            break
        }
    }
}
