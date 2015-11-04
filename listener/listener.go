package listener

import (
    "crypto/tls"
    "net"
    "errors"
)

type TrudyListener interface {
    //TODO: Listen should take two strings: "tcp" or "udp" and a port to listen on.
    //This parameter could create a Listener for both tcp and udp.
    Listen(string, *net.TCPAddr, tls.Config)

    //Accept returns a generic net.Conn and the file descriptor of the socket.
    Accept() (int, net.Conn, error)

    //Close shuts down the listener.
    Close() error
}

//The TCPListener struct implements the TrudyListener interface and handles TCP connections.
type TCPListener struct {
	Listener *net.TCPListener
}

func (tl *TCPListener) Listen(nets string, tcpAddr *net.TCPAddr, _ tls.Config) {
	tcpListener,err := net.ListenTCP(nets, tcpAddr)
    if err != nil {
        panic(err)
    }
    tl.Listener = tcpListener
}

func (tl *TCPListener) Accept() (fd int, conn net.Conn, err error) {
    cpointer, err := tl.Listener.AcceptTCP()
	if err != nil {
		return
	}
    file,err := cpointer.File()
    fd = int(file.Fd())
	conn,err = net.FileConn(file)
	return
}

func (tl *TCPListener) Close() error {
    return tl.Listener.Close()
}


//TLSListener struct implements the TrudyListener interface and handles TCP connections over TLS.
type TLSListener struct {
	Listener *net.TCPListener
	Config *tls.Config
}

func (tl *TLSListener) Accept() (fd int, conn net.Conn, err error) {
    cpointer, err := tl.Listener.AcceptTCP()
	if err != nil {
		return
	}
	file,err := cpointer.File()
	fd = int(file.Fd())
    fconn,err := net.FileConn(file)
    conn = tls.Server(fconn, tl.Config)
	return
}

func (tl *TLSListener) Listen(nets string, laddr *net.TCPAddr, config tls.Config) {
	if len(config.Certificates) == 0 {
		panic(errors.New("tls.Listen: no certificates in configuration"))
	}
	tcpListener, err := net.ListenTCP(nets, laddr)
    if err != nil {
        panic(err)
    }
	tl.Listener = tcpListener
	tl.Config = &config
}

func (tl *TLSListener) Close() error {
	return tl.Listener.Close()
}
