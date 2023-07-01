package main

import (
	"encoding/binary"
	"math"
	"net"
	"time"

	"golang.org/x/crypto/salsa20"
)

// https://github.com/Nenkai/PDTools/blob/master/SimulatorInterface/SimulatorInterface.cs
const (
	SendDelaySeconds = 10
	ReceivePort      = 33740
	SendPort         = 33739

	KEY = "Simulator Interface Packet GT7 ver 0.0"
)

var Key32 [32]byte

func init() {
	copy(Key32[:], KEY[0:32])
}

func salsa20Dec(dat []byte) (ddata []byte) {
	oiv := dat[0x40:0x44]
	iv1 := binary.LittleEndian.Uint32(oiv)
	iv2 := iv1 ^ 0xDEADBEAF
	iv := make([]byte, 8)
	binary.LittleEndian.PutUint32(iv, uint32(iv2))
	binary.LittleEndian.PutUint32(iv[4:], uint32(iv1))

	ddata = make([]byte, len(dat))
	salsa20.XORKeyStream(ddata, dat, iv, &Key32)

	magic := int32(binary.LittleEndian.Uint32(ddata[:4]))
	if magic != 0x47375330 {
		return []byte{}
	}
	return
}

func sendHB(s *net.UDPConn, ip string) {
	s.Write([]byte("A"))
	println("send heartbeat")
}

func Float32frombytes(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)
	float := math.Float32frombits(bits)
	return float
}

// Listen open udp connection to given ps5 ip/network name
// listen and decode messages
func Listen(ip string) (err error) {
	var (
		s *net.UDPConn
		n int
	)
	if s, err = net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: ReceivePort}); err != nil {
		return
	}
	defer s.Close()
	s.SetDeadline(time.Now().Add(time.Second * 10))

	sendHB(s, ip)

	pknt := 0
	data := make([]byte, 4096)
	for {
		n, _, err = s.ReadFromUDP(data)
		if err != nil {
			sendHB(s, ip)
			pknt = 0
			continue
		}

		pknt++
		println("received:", n, "bytes")

		ddata := salsa20Dec(data[:n])
		if len(ddata) == 0 {
			continue
		}

		rpm := Float32frombytes(ddata[15*4 : 15*4+4])
		println("RPM:", int(rpm))

		if pknt > 100 {
			sendHB(s, ip)
			pknt = 0
		}
	}
	return
}
