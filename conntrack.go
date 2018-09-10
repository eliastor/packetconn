package packetconn

import (
	"log"
	"net"
	"sync"
)

type key string

const minConns = 16
const pendingNewConns = 12

/*type key struct {
	addr [net.IPv6len + 2]byte
}

func (k *key) Set(addr string) *key {
	copy(k.addr[:], addr.IP)
	binary.LittleEndian.PutUint16(k.addr[net.IPv6len:], uint16(addr.Port))
	return k
}

func (k key) Addr() *net.UDPAddr {
	a := new(net.UDPAddr)
	a.IP = k.addr[:net.IPv6len]
	a.Port = int(binary.LittleEndian.Uint16(k.addr[net.IPv6len:]))
	return a
}

func (k key) String() string { return k.Addr().String() }

func (k key) Network() string { return k.Addr().Network() }
*/

//Tracker implements net.Tracker and server for udp streams tracking. Stream are track by destination ip & port
type Tracker struct {
	//net.Dialer - it works actually)
	net.PacketConn
	net.Listener
	conns   map[key]*PacketConn
	lock    sync.Mutex
	new     chan *PacketConn
	network string
}

//Accept method create net.Conn every time new connection is found
func (l *Tracker) Accept() (net.Conn, error) {
	return <-l.new, nil
}

/*func (l *Tracker) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	return l.PacketConn.WriteTo(p, addr)
}*/

func (l *Tracker) closeConn(conn net.Conn) {
	k := conn.RemoteAddr().String()

	if conn.LocalAddr() == l.LocalAddr() {
		l.lock.Lock()
		l.conns[key(k)].parent = nil
		l.conns[key(k)].Close()
		delete(l.conns, key(k))
		l.lock.Unlock()
	}
}

//Close current tracker. Also closes all current connections
func (l *Tracker) Close() error {
	l.lock.Lock()
	defer l.lock.Unlock()
	for k := range l.conns {
		l.conns[k].Close()
	}
	l.conns = nil
	close(l.new)
	///Should be closed externally
	//l.PacketConn.Close()
	return nil
}

func (l *Tracker) readHandler() {
	for {
		rPacket := make([]byte, Size)
		n, addr, err := l.ReadFrom(rPacket)
		if err != nil {
			log.Println(err)
			return
		}
		k := key(addr.String())
		l.lock.Lock()
		c, exists := l.conns[k]
		if !exists {
			if len(l.new) < cap(l.new) {
				c = newPacketConn(l, addr, 10, 3)

				l.conns[k] = c
				l.new <- c
			}
		} else if c.state == StateNew {
			l.conns[k].state = StateEstablished
		}

		l.lock.Unlock()

		///We can d it via select{default:}
		if len(c.rc) < cap(c.rc) {
			c.rc <- rPacket[:n]
		}
		//drop packet

	}
}

//Dial using current packet connection. If there is no such connection, new one will be created, otherwise old one would be returned
func (l *Tracker) Dial(network, address string) (conn net.Conn, err error) {
	if network != l.network {
		return nil, ErrWrongNetwork
	}
	var addr net.Addr
	switch network {
	case "udp":
		addr, err = net.ResolveUDPAddr(network, address)
	case "unix":
		addr, err = net.ResolveUnixAddr(network, address)
	}
	k := key(addr.String())
	c, exists := l.conns[k]
	if !exists {
		if len(l.new) < cap(l.new) {
			c = newPacketConn(l, addr, 10, 3)
			l.conns[k] = c
			l.new <- c
		} else {

			return nil, ErrTooManyConns
		}
	}
	return c, nil
}

// LocalAddr returns local address of current PacketConn
func (l *Tracker) LocalAddr() net.Addr {
	return l.PacketConn.LocalAddr()
}

// Addr returns local address of current PacketConn
func (l *Tracker) Addr() net.Addr {
	return l.PacketConn.LocalAddr()
}

//Track creates connection tracker on specific PacketConn
func Track(p net.PacketConn) (*Tracker, error) {
	l := new(Tracker)
	l.PacketConn = p
	l.network = p.LocalAddr().Network()
	l.conns = make(map[key]*PacketConn, minConns)
	l.new = make(chan *PacketConn, pendingNewConns)
	go l.readHandler()
	//go l.writeHandler()
	return l, nil
}
