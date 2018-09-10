package packetconn

import (
	"bytes"
	"encoding/hex"
	"net"
	"reflect"
	"strconv"
	"sync"
	"testing"
)

var defaultUDPAddress = &net.UDPAddr{IP: net.ParseIP("224.1.2.3"), Port: 8331}

func openPacketConn(addr net.Addr) (conn net.PacketConn) {
	var err error
	switch addr.Network() {
	case "udp":
		uaddr, _ := net.ResolveUDPAddr(addr.Network(), addr.String())
		conn, err = net.ListenMulticastUDP(addr.Network(), nil, uaddr)
		addr = uaddr

	case "unix":
		uaddr, _ := net.ResolveUnixAddr(addr.Network(), addr.String())
		conn, err = net.ListenUnixgram(addr.Network(), uaddr)
		addr = uaddr
	}
	if err != nil {
		panic(err)
	}
	return
}

func TestTrack(t *testing.T) {
	p := openPacketConn(defaultUDPAddress)
	defer p.Close()
	l, err := Track(p)
	if err != nil {
		t.Error(err)
	}
	if l.LocalAddr() != p.LocalAddr() {
		t.Error("Wrong addr in tracker")
	}

	t.Run("Input Connections", func(t *testing.T) {
		var i uint64
		for {
			if i == 5 {
				break
			}
			cSend := []byte("Client Send " + strconv.FormatUint(i, 10))
			cReceive := make([]byte, 1240)

			sSend := []byte("Server Send " + strconv.FormatUint(i, 10))
			sReceive := make([]byte, 1240)

			wg := sync.WaitGroup{}

			go func(send []byte, receive *[]byte) {
				wg.Add(1)
				defer wg.Done()
				c, err := net.Dial(l.LocalAddr().Network(), l.LocalAddr().String())
				if err != nil {
					t.Fatal(err)
				}
				_, err = c.Write(send)
				if err != nil {
					t.Fatal(err)
				}
				n, err := c.Read(*receive)
				if err != nil {
					t.Fatal(err)
				}
				*receive = (*receive)[:n]
				//fmt.Println(hex.Dump(*receive))

			}(cSend, &cReceive)
			conn, err := l.Accept()
			if err != nil {
				t.Error(err)
			}
			go func(conn net.Conn, send, receive []byte) {
				wg.Add(1)
				wg.Done()
				_, err = conn.Write(send)
			}(conn, sSend, sReceive)

			//it's needed for guaranteed packet receivnig
			wg.Wait()
			n, err := conn.Read(sReceive)
			if err != nil {
				t.Error(err)
			}
			if bytes.Compare(sReceive[:n], cSend) != 0 {
				t.Errorf("Data corrupted. Got: %v \n Expected: %v", hex.Dump(sReceive[:n]), hex.Dump(cSend))
			}

			if bytes.Compare(cReceive, sSend) != 0 {
				t.Errorf("Data corrupted. \nGot:\n%vExpected:\n%v", hex.Dump(cReceive), hex.Dump(sSend))
			}
			i++
		}
	})
}

func TestTracker_readHandler(t *testing.T) {
	tests := []struct {
		name string
		l    *Tracker
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.l.readHandler()
		})
	}
}

func TestTracker_Dial(t *testing.T) {
	type args struct {
		network string
		address string
	}
	tests := []struct {
		name     string
		l        *Tracker
		args     args
		wantConn net.Conn
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotConn, err := tt.l.Dial(tt.args.network, tt.args.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("Tracker.Dial() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotConn, tt.wantConn) {
				t.Errorf("Tracker.Dial() = %v, want %v", gotConn, tt.wantConn)
			}
		})
	}
}

func TestTracker_Addr(t *testing.T) {
	tests := []struct {
		name string
		l    *Tracker
		want net.Addr
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.l.LocalAddr(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Tracker.Addr() = %v, want %v", got, tt.want)
			}
		})
	}
}
