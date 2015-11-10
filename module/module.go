package module

import (
    "encoding/hex"
    "net"
)

//A wrapper that provides metadata that may be useful when mangling bytes on the network.
type Data struct {
    FromClient bool
    Bytes      []byte
    DestAddr   net.Addr
}

func (input Data) Ignore() bool {
    return true
}

func (input *Data) Mangle() {

}

func (input Data) Drop() bool {
    return false
}

func (input Data) PrettyPrint() string {
    return hex.Dump(input.Bytes)
}


