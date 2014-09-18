package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
)

type _Client struct {
	keyboardCapture *KeyboardCapture
	encryption      Encryption
	connection      *net.UDPConn
	configuration   *ClientConfiguration
}

func NewClient(config *ClientConfiguration) *_Client {
	ret := new(_Client)
	ret.keyboardCapture = NewKeyboardCapture(config.ForwardedKeys)
	ret.configuration = config
	return ret
}

// Starts the client.
// The function starts intercepting keys. A configurable set of keys cause an encrypted JSON
// message to be sent to the configured remote host via UDP.
// The function blocks until the Client.Stop() function was called.
// Returns an error in case the keyboard interception initialization or UDP client initialization failed.
func (t *_Client) Start() error {
	quit := make(chan bool)
	go func() {
		addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", t.configuration.Hostname, t.configuration.Port))
		if err != nil {
			log.Fatal(err)
		}
		t.connection, err = net.DialUDP("udp", nil, addr)
		if err != nil {
			log.Fatal(err)
		}
		defer t.connection.Close()

		t.encryption.Initialize(t.configuration.Secret)
		for {
			select {
			case k := <-t.keyboardCapture.KeyPressed:
				msg := Message{k}
				jsonData, _ := json.Marshal(msg)
				log.Printf("Sending key %d to remote host\n", k)
				t.connection.Write(t.encryption.Encrypt([]byte(jsonData)))
			case <-quit:
				return
			}
		}
	}()

	log.Println("Starting keyboard interception")
	err := t.keyboardCapture.SyncReceive()
	quit <- true

	return err
}

// Stops the key-press interception and causes the Client.Start() function to return.
// Intended to be called from another 'thread' (goroutine) as Client.Start().
func (t *_Client) Stop() {
	log.Println("Stopping keyboard interception")
	t.keyboardCapture.Stop()
}
