package main

import (
    "net"
    "log"
    "crypto/tls"
    "github.com/gorilla/websocket"
    "github.com/kelbyludwig/trudy/pipe"
    "github.com/kelbyludwig/trudy/module"
    "github.com/kelbyludwig/trudy/listener"
    ws "github.com/kelbyludwig/trudy/websocket"
)

var connection_count uint
var interceptSendChannel chan []byte
var interceptRecvChannel chan []byte

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

    websocketReady := make(chan bool)
    go websocketHandler(websocketReady)
    <-websocketReady
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

        data := module.Data{FromClient: true,
            Bytes: buffer[:bytesRead],
            DestAddr: pipe.DestinationInfo()}
        if data.Drop() {
            continue
        }

        if data.DoInterceptFromClient() {
            interceptSendChannel <- data.Bytes[:bytesRead]
            data.Bytes = <-interceptRecvChannel
        } else if data.DoMangle() {
            //TODO: I think else if maybe a sane suggestion. Maybe it is not.
            data.Mangle()
            bytesRead = len(data.Bytes)
        }

        if data.DoPrint() {
            log.Printf("Client -> Server: \n%v\n", data.PrettyPrint())
        }

        _, err = pipe.WriteDestination(data.Bytes[:bytesRead])
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
        data := module.Data{FromClient: false,
            Bytes: buffer[:bytesRead],
            DestAddr: pipe.DestinationInfo()}

        if data.Drop() {
            continue
        }

        if data.DoMangle() {
            data.Mangle()
            bytesRead = len(data.Bytes)
        }

        if data.DoPrint() {
            log.Printf("Server -> Client: \n%v\n", data.PrettyPrint())
        }
        _,err = pipe.WriteSource(data.Bytes[:bytesRead])
        if err != nil {
            break
        }
    }
}

func websocketHandler(ready chan bool) {
    conn := ws.Listen()
    defer conn.Close()
    interceptSendChannel = make(chan []byte)
    interceptRecvChannel = make(chan []byte)
    ready <- true
    for {
        ibytes := <-interceptSendChannel
        conn.WriteMessage(websocket.TextMessage, ibytes)
        _,obytes,_ := conn.ReadMessage()
        interceptRecvChannel <- obytes
    }
}
