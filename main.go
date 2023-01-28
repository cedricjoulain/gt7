package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math"
	"net"

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

func SendHeartBeat(conn net.Conn) {
	fmt.Fprintf(conn, "A")
}

func Salsa20Decode(in []byte) (out []byte) {
	// Seed IV is always located there
	iv1 := binary.LittleEndian.Uint32(in[0x40:0x44])
	iv2 := iv1 ^ 0xDEADBEAF // Notice DEADBEAF, not DEADBEEF TODO check for endianess

	IV := make([]byte, 8)
	binary.LittleEndian.PutUint32(IV[0:], iv2)
	binary.LittleEndian.PutUint32(IV[4:], iv1)
	// Magic should be "G7S0" when decrypted
	out = make([]byte, len(in))
	salsa20.XORKeyStream(out, in, IV, &Key32)
	//check magic number
	magic := binary.LittleEndian.Uint32(out[0:4])
	if magic != 0x47375330 {
		return []byte{}
	}
	return
}

func Float64frombytes(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
}

func Float32frombytes(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)
	float := math.Float32frombits(bits)
	return float
}

func main() {
	ptrHost := flag.String("host", "PS5-89A73E", "PS5 IP or name on LAN")
	flag.Parse()

	rconn, err := net.Dial("udp", fmt.Sprintf("%s:%d", *ptrHost, ReceivePort))
	if err != nil {
		log.Fatal("unable to open receiver", err)
	}
	defer rconn.Close()
	wconn, err := net.Dial("udp", fmt.Sprintf("%s:%d", *ptrHost, SendPort))
	if err != nil {
		log.Fatal("unable to open sender", err)
	}
	defer wconn.Close()

	SendHeartBeat(wconn)

	pknt := 0
	buff := make([]byte, 4096)
	r := bufio.NewReader(rconn)
	for {
		size, err := r.Read(buff)
		if err != nil {
			SendHeartBeat(wconn)
			pknt = 0
		}
		pknt++
		log.Println("received", size, "bytes")
		if size > 0x44 {
			message := Salsa20Decode(buff[0:size])
			rpm := Float32frombytes(message[15*4 : 15*4+4])
			log.Println("RPM:", rpm)
		}
		if pknt > 100 {
			SendHeartBeat(wconn)
			pknt = 0
		}
	}
	return
}
