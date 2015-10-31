package tlspipe

import (
    "net"
    "crypto/tls"
)

type TLSPipe struct {
    id uint
    destination tls.Conn
    source tls.Conn
}

func NewTLSPipe (id uint, sourceConn tls.Conn) (pipe TLSPipe, err error) {
    tlspipe := new(TLSPipe)
    f, err := sourcConn.(*net.TCPConn).File()
    if err != nil {
        log.Println("[DEBUG] Failed to read connection file descriptor.")
        sourceConn.Close()
        return *tlspipe, err
    }
}
