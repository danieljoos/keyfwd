package main

import (
	"bufio"
	"fmt"
	"github.com/howeyc/gopass"
	"os"
	"strconv"
	"strings"
)

// Interactive client configuration.
func ConfigureClient() {
	var configuration ClientConfiguration
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("%-10s: ", "Hostname")
	configuration.Hostname, _ = reader.ReadString(byte('\n'))
	configuration.Hostname = strings.Trim(configuration.Hostname, "\n\r\t ")

	fmt.Printf("%-10s: ", "Port")
	port, _ := reader.ReadString(byte('\n'))
	port = strings.Trim(port, "\n\r\t ")
	configuration.Port, _ = strconv.ParseUint(port, 10, 0)

	fmt.Printf("%-10s: ", "Password")
	configuration.Secret = gopass.GetPasswdMasked()

	configuration.ForwardedKeys = GetDefaultForwardedKeys()

	StoreClientConfiguration(&configuration)
}

// Interactive server configuration
func ConfigureServer() {
	var configuration ServerConfiguration

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("%-10s: ", "Port")
	port, _ := reader.ReadString(byte('\n'))
	port = strings.Trim(port, "\n\r\t ")
	configuration.Port, _ = strconv.ParseUint(port, 10, 0)

	fmt.Printf("%-10s: ", "Password")
	configuration.Secret = gopass.GetPasswdMasked()

	StoreServerConfiguration(&configuration)
}
