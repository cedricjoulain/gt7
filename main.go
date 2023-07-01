package main

import (
	"flag"
	"log"
)

func main() {
	ipPtr := flag.String("ip", "PS5-89A73E", "ip or network name of PS5 GranTurismo7 to listen")
	filePtr := flag.String("file", "", "GT7 raw data file to analyse")
	recordPtr := flag.Bool("rec", false, "record mode, listening at ip")
	flag.Parse()

	if *recordPtr {
		if err := Listen(*ipPtr); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := Analyse(*filePtr); err != nil {
			log.Fatal(err)
		}
	}
}
