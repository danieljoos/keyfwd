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

type NotifyIcon interface {
	Start() error
	Stop()
	OnClick() chan NotifyIconButton
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Missing argument")
	}

	var action Runnable
	var notifyIcon NotifyIcon
	var err error

	switch os.Args[1] {
	case "client":
		action = NewClient(LoadClientConfiguration())
		notifyIcon, err = NewNotifyIcon("Key Forwarding (client)", IconClient)
	case "server":
		action = NewServer(LoadServerConfiguration())
		notifyIcon, err = NewNotifyIcon("Key Forwarding (server)", IconServer)
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

	if err != nil {
		log.Fatal(err)
	}

	quit := make(chan bool)

	// Shutdown signal handler
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		for {
			select {
			case <-c:
				quit <- true
			case <-quit:
			}
			log.Println("Shutdown signal received")
			action.Stop()
			notifyIcon.Stop()
			os.Exit(0)
		}
	}()

	// Keyboard interception
	go func() {
		err := action.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Notify icon click handler
	go func() {
		for {
			select {
			case button := <-notifyIcon.OnClick():
				switch button {
				case LeftMouseButton:
					ToggleShowConsoleWindow()
				case RightMouseButton:
					ShowConsoleWindow()
					quit <- true
					return
				}
			case <-quit:
				return
			}
		}
	}()

	// Notify icon run
	err = notifyIcon.Start()
	if err != nil {
		log.Fatal(err)
	}
}
