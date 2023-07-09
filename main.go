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
	data1     [][]float64
	data2     [][]float64
)

func main() {
	ipPtr := flag.String("ip", "PS5-89A73E", "ip or network name of PS5 GranTurismo7 to listen")
	file1Ptr := flag.String("file1", "", "GT7 raw data ref file to analyse (small point)")
	lap1Ptr := flag.Int("lap1", 1, "lap number to look at")
	file2Ptr := flag.String("file2", "", "GT7 raw data file to compare (big point)")
	lap2Ptr := flag.Int("lap2", 1, "lap number to look at")
	recordPtr := flag.Bool("rec", false, "record mode, listening at ip")
	portPtr := flag.Int("port", 8080, "an int")
	flag.Parse()

	if *recordPtr {
		if err := Listen(*ipPtr); err != nil {
			log.Fatal(err)
		}
		// end
		os.Exit(0)
	} else {
		packets, err := Analyse(*file1Ptr, *lap1Ptr)
		if err != nil {
			log.Fatal(err)
		}
		EchartsData(packets, &data1)
		packets, err = Analyse(*file2Ptr, *lap2Ptr)
		if err != nil {
			log.Fatal(err)
		}
		EchartsData(packets, &data2)
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
