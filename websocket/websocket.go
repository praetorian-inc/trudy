package websocket

import (
    "github.com/gorilla/websocket"
    "net/http"
    "io/ioutil"
    "io"
)

var editor string
var wsConn *websocket.Conn

func Listen() *websocket.Conn {
    html,_ := ioutil.ReadFile("editor.html")
    editor = string(html)
    http.HandleFunc("/", EditorServer)
    http.HandleFunc("/ws", WebSocketServer)
    err := http.ListenAndServe(":8090", nil)
    if err != nil {
        panic(err)
    }
    return wsConn
}

func EditorServer(w http.ResponseWriter, req *http.Request) {
    io.WriteString(w, editor)
}

func WebSocketServer(w http.ResponseWriter, req *http.Request) {
    upgrader := websocket.Upgrader{
        ReadBufferSize:  65535,
        WriteBufferSize: 65535,
    }
    conn, err := upgrader.Upgrade(w, req, nil)
    if err != nil {
        panic(err)
    }
    wsConn = conn
}

//Sends packets to the web browser if the packet should be intercepted.
func InterceptWriter(input []byte) {
}

//Retrieves packets from the web browser and sends them back through Trudy.
func InterceptReader(input []byte) {

}
