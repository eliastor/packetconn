# packetconn
Creates **"virtual" net.Conn** connections with remote hosts for packets come from/to **net.PacketConn** to such hosts

# Details
If you've ever heard about **Linux conntrack** tldr this library is like conntrack but works with any **net.PacketConn** (unixgram and udp).

**net.PacketConn** connection is packet only connection without any real connection promoted by go stdlib and can be used with udp or unixgram.

Every connection have 2 sides: let's assume them local and remote hosts. PacketConn pass packets from different remote hosts to single local one. **Packetconn** spreads different remote hosts to multiple "virtual" connection (the implements net.Conn BTW) you can handle with as you always do using *Accept* method. Note that if you want to connect another node and answer need to be received by your tracked connection you should use *Dial* method

# Example

```Go
var defaultMulticast = net.UDPAddr{
	IP:   net.IPv4(224, 1, 1, 1),
	Port: 1111,
}

//Create anything implements net.PacketConn
l, _ := net.ListenMulticastUDP("udp", nil, &h.listen)

//Track connections on it
tTracker, _ := packetconn.Track(l)

//You can dial to remote hosts
connOut, _err := tListener.Dial("udp", "192.168.1.111")

//Or receive incoming connection
for {
    conn, err := tListener.Accept()
    //work with conn as you always do with net.Conn
}
```
