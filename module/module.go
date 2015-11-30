package module

import (
	"encoding/hex"
	"net"
)

//Data is a thin wrapper that provides metadata that may be useful when mangling bytes on the network.
type Data struct {
	FromClient bool
	Bytes      []byte
	DestAddr   net.Addr
	SrcAddr    net.Addr
}

//DoMangle will return true, if Data needs to be sent to the Mangle function.
func (input Data) DoMangle() bool {
	return true
}

//Mangle can modify/replace the Bytes values within the Data struct. This can be empty if no
//programmatic mangling needs to be done.
func (input *Data) Mangle() {

}

//Drop will return true, if the Data needs to be dropped between one side of the pipe (i.e. Client->Server, Server->Client)
func (input Data) Drop() bool {
	return false
}

//PrettyPrint returns the string representation of the data. This string will be value logged to output.
func (input Data) PrettyPrint() string {
	return hex.Dump(input.Bytes)
}

//DoPrint will return true, if the PrettyPrinted version of the Data struct needs to be logged to the output.
func (input Data) DoPrint() bool {
	return true
}

//DoIntercept returns true if data should be sent to the Trudy interceptor.
func (input Data) DoIntercept() bool {
	return false
}
