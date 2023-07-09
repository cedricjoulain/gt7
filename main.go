package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var templates = template.Must(template.ParseGlob("templates/*"))

// will be used as Last-Modified for html/js... templates
var (
	startTime = time.Now()
	data      [][]float64
)

func main() {
	ipPtr := flag.String("ip", "PS5-89A73E", "ip or network name of PS5 GranTurismo7 to listen")
	filePtr := flag.String("file", "", "GT7 raw data file to analyse")
	recordPtr := flag.Bool("rec", false, "record mode, listening at ip")
	lapPtr := flag.Int("lap", 1, "lap number to look at")
	portPtr := flag.Int("port", 8080, "an int")
	flag.Parse()

	if *recordPtr {
		if err := Listen(*ipPtr); err != nil {
			log.Fatal(err)
		}
		// end
		os.Exit(0)
	} else {
		packets, err := Analyse(*filePtr, *lapPtr)
		if err != nil {
			log.Fatal(err)
		}
		data = EchartsData(packets)
	}
	router := mux.NewRouter()

	// handler for static files :
	StaticDir := handlers.CompressHandler(http.FileServer(http.Dir(".")))
	router.PathPrefix("/js/").Handler(StaticDir)
	router.HandleFunc("/", makeGzipHandler(HandleHome))

	// Serve Favory icon
	router.HandleFunc("/favicon.ico", makeGzipHandler(HandleFavicon))

	// Serve 404
	router.NotFoundHandler = router.HandleFunc("/", makeGzipHandler(HandleHome)).GetHandler()

	// Launch server
	srv, cRouter := startMuxHTTPServer(router, *portPtr)
	// Block until we receive our signal.
	<-cRouter
	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("HTTP server stopped")
	os.Exit(0)
}

// startMuxHTTPServer start Mux serveur return chan to Shutdown
func startMuxHTTPServer(router *mux.Router, port int) (srv *http.Server, c chan os.Signal) {
	c = make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	// Launch server
	srv = &http.Server{
		//timeouts to avoid Slowloris attacks
		//WriteTimeout: time.Second * 15,
		ReadTimeout: time.Second * 15,
		//IdleTimeout:  time.Second * 60,
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}
	go func() {
		log.Println("HTTP Server started on", port)
		if err := srv.ListenAndServe(); err != nil {
			log.Println("ListenAndServe() Error", err)
		}
	}()
	return
}
