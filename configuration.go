package main

type ClientConfiguration struct {
	Hostname      string
	Port          uint64
	Secret        []byte
	ForwardedKeys []int
}

type ServerConfiguration struct {
	Port   uint64
	Secret []byte
}
