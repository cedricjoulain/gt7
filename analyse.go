package main

import (
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"os"
)

type Message [296]byte

func (m Message) RPM() float32 {
	return math.Float32frombits(binary.LittleEndian.Uint32(m[15*4:]))
}

// Analyse open raw decoded data and parse
func Analyse(filename string) (err error) {
	var (
		nbr     int
		r       io.Reader
		f       *os.File
		header  [2]byte
		message Message
	)
	if f, err = os.Open(filename); err != nil {
		err = fmt.Errorf("unable to open %s:%s", filename, err)
		return
	}
	defer f.Close()
	// is it a gzip file ?
	// use magic number
	if _, err = io.ReadFull(f, header[:]); err != nil {
		err = fmt.Errorf("unable to read head os %s:%s", filename, err)
		return
	}
	// back to beginning of file
	if _, err = f.Seek(0, 0); err != nil {
		return
	}
	if header[0] == 0x1f && header[1] == 0x8b {
		// this is a gzip
		if r, err = gzip.NewReader(f); err != nil {
			err = fmt.Errorf("unable to create gzip reader for %s:%s", filename, err)
			return
		}
	} else {
		r = f
	}
	for err == nil {
		if _, err = io.ReadFull(r, message[:]); err != nil {
			break
		}
		nbr++
		log.Println("message", nbr, "RPM", message.RPM())
	}
	if err == io.EOF {
		// just the end
		err = nil
	}
	return
}
