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

//If this returns true, the data will be sent to the Mangle function.
func (input Data) DoMangle() bool {
	return true
}

//This function should modify/replace the Bytes values within the Data struct.
func (input *Data) Mangle() {

}

//If this returns true, the data will not be sent to the other end of the pipe.
func (input Data) Drop() bool {
	return false
}

//Returns the string representation of the data. This string will be value logged to output.
func (input Data) PrettyPrint() string {
	return hex.Dump(input.Bytes)
}

//If this returns true, the PrettyPrinted version of the Data struct will be logged to output.
func (input Data) DoPrint() bool {
	return true
}

//Returns true if data should be sent to the Trudy interceptor.
func (input Data) DoIntercept() bool {
	return false
}
