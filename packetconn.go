package packetconn

import (
	"net"
)

//States of packet connection
const (
	StateNew uint8 = iota
	StateEstablished
)

//Size of data payload
const Size = (1280 - 40)

var stateStrings = map[uint8]string{
	StateNew:         "new",
	StateEstablished: "established",
}

//StateString return human readable connection state
func StateString(s uint8) string {
	return stateStrings[s]
}

//PacketConn adds net.Conn functionality to net.PacketConn (PacketConn in fact)
type PacketConn struct {
	net.PacketConn
	parent *Tracker
	addr   net.Addr
	state  uint8
	rc     chan []byte
}

//Creates newPacketConnection
func newPacketConn(pc net.PacketConn, addr net.Addr, readBuffer, writeBuffer int) *PacketConn {
	u := new(PacketConn)
	u.PacketConn = pc
	u.addr = addr
	u.rc = make(chan []byte, readBuffer)
	return u
}

//State returns current state of connection. For human-readable values use StateString()
func (u *PacketConn) State() uint8 {
	return u.state
}

//StateString return human-readable state of connection
func (u *PacketConn) StateString() string {
	return stateStrings[u.state]
}

//Close connetion. there is different behaviour
func (u *PacketConn) Close() error {
	if u.parent != nil {
		u.parent.closeConn(u)
	} else {
		close(u.rc)
		u = nil
	}
	return nil
}

func (u *PacketConn) Read(b []byte) (n int, err error) {
	///TODO timeout
	n = copy(b, <-u.rc)
	return
}

func (u *PacketConn) Write(b []byte) (n int, err error) {
	u.PacketConn.WriteTo(b, u.RemoteAddr())
	return
}

//RemoteAddr return remote address of current connection
func (u *PacketConn) RemoteAddr() net.Addr {
	return u.addr
}
