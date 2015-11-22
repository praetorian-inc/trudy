package main

import (
    "io"
    "net"
    "net/http"
    "log"
    "crypto/tls"
    "github.com/kelbyludwig/trudy/pipe"
    "github.com/kelbyludwig/trudy/module"
    "github.com/kelbyludwig/trudy/listener"
    "github.com/gorilla/websocket"
)

var connection_count uint
var websocketConn *websocket.Conn

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

    go websocketHandler()
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

        if data.DoIntercept() {
            if err := websocketConn.WriteMessage(websocket.TextMessage, data.Bytes); err != nil {
                log.Printf("[ERR] Failed to write to websocket.\n")
                continue
            }
        }

        if data.DoMangle() {
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

func websocketHandler() {
    listener,err := net.Listen("tcp", ":8080")
    if err != nil {
        panic(err)
    }
    upgrader := websocket.Upgrader{ ReadBufferSize: 65535, WriteBufferSize: 65535 }
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        io.WriteString(w, editor)
    })
    http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        websocketConn, err = upgrader.Upgrade(w, r, nil)
        if err != nil {
            log.Printf("[ERR] Could not upgrade websocket connection.")
            return
        }
    })
    err = http.Serve(listener,nil)
    if err != nil {
        panic(err)
    }
}

const editor string = `<!-- this wonderful page was found here: https://github.com/xem/hex -->
<body onload='
// Reset the textarea value
m.value="00";

// Init the top cell content
for(i=0;i<16;i++)
  t.innerHTML+=(0+i.toString(16)).slice(-2)+" ";
'>

<!-- TRUDY SPECIFIC CODE ADDED FOR THIS PROJECT -->
<h1> ~ Trudy Intercept ~ </h1>
<script>
    //TODO: This will have to be updated. Need to pull the address of the VM from the DOM.
    var url = window.location.href
    var arr = url.split("/");
    var ws_url = "ws://" + arr[2] + "/ws"
    var socket = new WebSocket(ws_url)
    socket.onmessage = function (event) {
        document.getElementById('m').value = event.data
        document.getElementById('m').oninput()
    }
</script>
<!-- END TRUDY SPECIFIC CODE -->
</body>
<table border><td><pre><td id=t><tr><td id=l width=80>00000000<td><textarea spellcheck=false id=m oninput='
// On input, store the length of clean hex before the textarea caret in b
b=value
.substr(0,selectionStart)
.replace(/[^0-9A-F]/ig,"")
.replace(/(..)/g,"$1 ")
.length;

// Clean the textarea value
value=value
.replace(/[^0-9A-F]/ig,"")
.replace(/(..)/g,"$1 ")
.replace(/ $/,"")
.toUpperCase();

// Set the height of the textarea according to its length
style.height=(1.5+value.length/47)+"em";

// Reset h
h="";

// Loop on textarea lines
for(i=0;i<value.length/48;i++)
  
  // Add line number to h
  h+=(1E7+(16*i).toString(16)).slice(-8)+" ";

// Write h on the left column
l.innerHTML=h;

// Reset h
h="";

// Loop on the hex values
for(i=0;i<value.length;i+=3)
  
  // Convert them in numbers
  c=parseInt(value.substr(i,2),16),
  
  // Convert in chars (if the charCode is in [64-126] (maybe more later)) or ".".
  h=63<c&&127>c?h+String.fromCharCode(c):h+".";
  
// Write h in the right column (with line breaks every 16 chars)
r.innerHTML=h.replace(/(.{16})/g,"$1 ");

// If the caret position is after a space or a line break, place it at the previous index so we can use backspace to erase hex code
if(value[b]==" ")
  b--;

// Put the textarea caret at the right place
setSelectionRange(b,b)'
cols=48></textarea><td width=160 id=r>.</td>
</table>
<style>
*{margin:0;padding:0;vertical-align:top;font:1em/1em courier}
#m{height:1.5em;resize:none;overflow:hidden}
#t{padding:0 2px}
#w{position:absolute;opacity:.001}
</style>
<button onclick="alert('sending does not work')">send</button>
`
