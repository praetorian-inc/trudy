package module

import (
	"encoding/hex"
	"net"
)

//Data is a thin wrapper that provides metadata that may be useful when mangling bytes on the network.
type Data struct {
	FromClient bool     //FromClient is true is the data sent is coming from the client (the device you are proxying)
	Bytes      []byte   //Bytes is a byte slice that contians the TCP data
	DestAddr   net.Addr //DestAddr is net.Addr of the server
	SrcAddr    net.Addr //SrcAddr is the net.Addr of the client (the device you are proxying)
}

//DoMangle will return true if Data needs to be sent to the Mangle function.
func (input Data) DoMangle() bool {
	return true
}

//Mangle can modify/replace the Bytes values within the Data struct. This can be empty if no
//programmatic mangling needs to be done.
func (input *Data) Mangle() {

}

//Drop will return true if the Data needs to be dropped before going through the pipe.
func (input Data) Drop() bool {
	return false
}

//PrettyPrint returns the string representation of the data. This string will be the value that is logged to the console.
func (input Data) PrettyPrint() string {
	return hex.Dump(input.Bytes)
}

//DoPrint will return true if the PrettyPrinted version of the Data struct needs to be logged to the console.
func (input Data) DoPrint() bool {
	return true
}

//DoIntercept returns true if data should be sent to the Trudy interceptor.
func (input Data) DoIntercept() bool {
	return false
}
