package main

import "net/http"
import "io/ioutil"
import "io"
import "github.com/gorilla/websocket"

var editor string

func main() {
    html,_ := ioutil.ReadFile("editor.html")
    editor = string(html)
    http.HandleFunc("/", EditorServer)
    err := http.ListenAndServe(":8090", nil)
    if err != nil {
        panic(err)
    }
}

func EditorServer(w http.ResponseWriter, req *http.Request) {
    io.WriteString(w, editor)
}

func WebSocketServer(w http.ResponseWriter, req *http.Request) {
    var upgrader = websocket.Upgrader{
        ReadBufferSize:  65535,
        WriteBufferSize: 65535,
    }
    conn, err := ugrader.Upgrade(w, r, nil)
    if err != nil {
        panic(err)
    }
}

//Sends packets to the webrowser if the packet wants to be intercepted.
func InterceptWriter(conn *websocket.Conn) {
}

//Retrieves packets from the web browser and sends them back through Trudy.
func InterceptReader(conn *websocket.Conn) {

}

//document.getElementById('m').value = "42 42 42 42 42 42"
//document.getElementById('m').oninput()
