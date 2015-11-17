package main

import "net/http"
import "io/ioutil"
import "io"
//import "github.com/gorilla/websocket"

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


