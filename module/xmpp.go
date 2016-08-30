package module

import (
	"bytes"
	"crypto/tls"
	"encoding/hex"
	"github.com/praetorian-inc/trudy/pipe"
	"log"
	"net"
	"strings"
)

//Data is a wrapper that provides metadata that may be useful when mangling bytes on the network.
type Data struct {
	FromClient bool                   //FromClient is true is the data sent is coming from the client (the device you are proxying)
	KV         map[string]interface{} //KV is a map that can be used to pass data between different module calls.
	Bytes      []byte                 //Bytes is a byte slice that contians the TCP data
	TLSConfig  *tls.Config            //TLSConfig is a TLS server config that contains Trudy's TLS server certficiate.
	ServerAddr net.Addr               //ServerAddr is net.Addr of the server
	ClientAddr net.Addr               //ClientAddr is the net.Addr of the client (the device you are proxying)
}

var startTLSElementSingle string = "<starttls xmlns='urn:ietf:params:xml:ns:xmpp-tls'>"
var startTLSElementDouble string = "<starttls xmlns=\"urn:ietf:params:xml:ns:xmpp-tls\">"
var proceedElementDouble string = "<proceed xmlns=\"urn:ietf:params:xml:ns:xmpp-tls\"/>"
var proceedElementSingle string = "<proceed xmlns='urn:ietf:params:xml:ns:xmpp-tls'/>"

//DoPrint will return true if the PrettyPrinted version of the Data struct
//needs to be logged to the console.
func (input Data) DoPrint() bool {
	//Only print client/server data sent over XMPP Ports.
	return strings.Contains(input.ServerAddr.String(), ":5225") || strings.Contains(input.ClientAddr.String(), ":5225")
}

//AfterWriteToServer is a function that will be called after data is sent to
//a server.
func (input *Data) AfterWriteToServer(p pipe.TrudyPipe) {

	if input.FromClient && (bytes.Contains(input.Bytes, []byte(startTLSElementDouble)) ||
		bytes.Contains(input.Bytes, []byte(startTLSElementSingle))) {
		//The client has sent StartTLS response to the server's StartTLS
		//requestion. Trudy will handle that upgrade.

		//Tell the client to proceed.
		p.WriteToClient([]byte(proceedElementDouble))
		//Upgrade the connection and prepare for a ClientHandshake
		tlsConn := tls.Server(p.ClientConn(), input.TLSConfig)
		log.Printf("[INFO] (%v) Upgrading client-side connection.\n", p.Id())
		err := tlsConn.Handshake()
		if err != nil {
			log.Printf("[ERR] (%v) Failure in upgrading client-side connection: %v\n", p.Id(), err)
			p.Close()
			return
		}
		p.SetClientConn(tlsConn)
		log.Printf("[INFO] (%v) Succesfully upgraded client-side connection.\n", p.Id())
	}
}

//BeforeWriteToClient is a function that will be called before data is sent to
//a client.
func (input *Data) BeforeWriteToClient(p pipe.TrudyPipe) {

	if !input.FromClient && (bytes.Contains(input.Bytes, []byte(proceedElementDouble)) ||
		bytes.Contains(input.Bytes, []byte(proceedElementSingle))) {

		//We have recieved a proceed from the server. Trudy will
		//now upgrade the server-side connection.

		tlsConn := tls.Client(p.ServerConn(), input.TLSConfig)
		log.Printf("[INFO] (%v) Sending handshake to server.\n", p.Id())
		err := tlsConn.Handshake()
		if err != nil {
			log.Printf("[ERR] (%v) Failure in upgrading server-side connection: %v\n", p.Id(), err)
			p.Close()
			return
		}
		log.Printf("[INFO] (%v) Succesfully upgraded server-side connection\n", p.Id())
		p.SetServerConn(tlsConn)

		//Lets drop the proceed message so its not sent to the client. (Since Trudy already sent one)
		input.Bytes = []byte{}
	}

}

//AfterWriteToClient is a function that will be called after data is sent to
//a client.
func (input *Data) AfterWriteToClient(p pipe.TrudyPipe) {

}

//BeforeWriteToServer is a function that will be called before data is sent to
//a server.
func (input *Data) BeforeWriteToServer(p pipe.TrudyPipe) {

}

//
//
// Unmodified module methods. All methods past this point are using the default implementation.
//
//

//DoIntercept returns true if data should be sent to the Trudy interceptor.
func (input Data) DoIntercept() bool {
	return false
}

//Mangle can modify/replace the Bytes values within the Data struct. This can
//be empty if no programmatic mangling needs to be done.
func (input *Data) Mangle() {

}

//PrettyPrint returns the string representation of the data. This string will
//be the value that is logged to the console.
func (input Data) PrettyPrint() string {
	return hex.Dump(input.Bytes)
}

//Deserialize should replace the Data struct's Bytes with a deserialized bytes.
//For example, unpacking a HTTP/2 frame would be deserialization.
func (input *Data) Deserialize() {

}

//Serialize should replace the Data struct's Bytes with the serialized form of
//the bytes. The serialized bytes will be sent over the wire.
func (input *Data) Serialize() {

}

//DoMangle will return true if Data needs to be sent to the Mangle function.
func (input Data) DoMangle() bool {
	return false
}

//Drop will return true if the Data needs to be dropped before going through
//the pipe.
func (input Data) Drop() bool {
	return false
}
