package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
)

type Server struct {
	configuration *ServerConfiguration
	encryption    Encryption
	emitter *KeyboardEmitter
}

func NewServer(config *ServerConfiguration) *Server {
	ret := new(Server)
	ret.configuration = config
	ret.emitter = NewKeyboardEmitter()
	return ret
}

func (t *Server) Start() error {
	t.encryption.Initialize(t.configuration.Secret)
	var buf [1024]byte
	addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", t.configuration.Port))
	log.Println(fmt.Sprintf("Listening on port %d", t.configuration.Port))
	sock, _ := net.ListenUDP("udp", addr)
	for {
		rlen, remote, err := sock.ReadFromUDP(buf[:])
		if err == nil {
			var msg Message
			if json.Unmarshal(t.encryption.Decrypt(buf[0:rlen]), &msg) == nil {
				log.Println(fmt.Sprintf("Received key from host '%s': %d ", remote.IP.String(), msg.VkCode))
				t.emitter.SendKey(msg.VkCode)
			}
		}
	}
}

func (t *Server) Stop() {
}
