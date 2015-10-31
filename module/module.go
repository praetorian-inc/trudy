package module

import (
    "encoding/hex"
)

//TODO: Make modules aware of dataflow (client->server) & (server->client)

func Pass(input []byte) bool {
    return true
}

func Mangle(input []byte) []byte {
    return input
}

func Drop(input []byte) bool {
    return false
}

func PrettyPrint(input []byte) string {
    return hex.Dump(input)
}


