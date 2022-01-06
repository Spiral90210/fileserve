package main

import (
	_ "embed"
	"flag"
	"log"
	"os"

	"github.com/Spiral90210/fileserve/pkg/server"
)

//go:embed favicon.ico
var favicon []byte

func main() {

	listenFlag := flag.String("listen", ":8007", "listen address")
	dataDirFlag := flag.String("data-dir", "/var/data", "data directory")
	includeHiddenFlag := flag.Bool("include-hidden", false, "include hidden files")

	flag.Parse()

	if *listenFlag == "" {
		log.Fatal("--listen address required e.g. 127.0.0.1:8007")
	}

	if *dataDirFlag == "" {
		log.Fatal("--data-dir required")
	}

	info, err := os.Stat(*dataDirFlag)

	if err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("configured data directory does not exist: %s\n", *dataDirFlag)
		}
		log.Fatalf("could not stat configured data dir: %s\n", *dataDirFlag)
	}

	if !info.IsDir() {
		log.Fatalf("configured data directory is not a directory: %s\n", *dataDirFlag)
	}

	s := &server.Server{
		BindAddr:      *listenFlag,
		Datadir:       *dataDirFlag,
		IncludeHidden: *includeHiddenFlag,
		Favicon:       favicon,
	}

	_ = s.ListenAndServe()
}
