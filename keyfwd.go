package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Runnable interface {
	Start() error
	Stop()
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Missing argument")
	}

	var action Runnable

	switch os.Args[1] {
	case "client":
		action = NewClient(LoadClientConfiguration())
	case "server":
		action = NewServer(LoadServerConfiguration())
	case "configure":
		if len(os.Args) < 3 {
			log.Fatal("Missing argument")
		}
		switch os.Args[2] {
		case "client":
			ConfigureClient()
		case "server":
			ConfigureServer()
		default:
			log.Fatal("Unknown configuration target")
		}
		os.Exit(0)
	default:
		log.Fatal("Unknown action")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		for {
			<-c
			log.Println("Shutdown signal received")
			action.Stop()
			os.Exit(0)
		}
	}()
	err := action.Start()
	if err != nil {
		log.Fatal(err)
	}
}
