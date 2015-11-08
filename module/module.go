package module

import (
    "encoding/hex"
)

//A wrapper that provides metadata that may be useful when mangling bytes on the network.
type Data struct {
    FromClient bool
    Bytes      []byte
}

func Pass(input Data) bool {
    return true
}

func Mangle(input Data) Data {
    return input
}

func Drop(input Data) bool {
    return false
}

func PrettyPrint(input Data) string {
    return hex.Dump(input.Bytes)
}


