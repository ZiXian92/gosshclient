package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/zixian92/gossh"
)

var localUser, remoteUser, host string
var port uint

func main() {
	// Declare runtime configurable flags to get details for the SSH connection
	flag.StringVar(&localUser, "user", os.Getenv("USER"), "Name of local user to log in as")
	flag.StringVar(&remoteUser, "ruser", os.Getenv("USER"), "Name of remote user to log in as")
	flag.StringVar(&host, "host", "localhost", "Hostname of remote server to log in to")
	flag.UintVar(&port, "port", 22, "Port of remote server to connect to")
	flag.Parse()

	clientConfig := gossh.ClientConfig{
		Host:          fmt.Sprintf("%s:%d", host, port),
		LocalUser:     localUser,
		RemoteUser:    remoteUser,
		HostKeyPolicy: gossh.StrictHostKeyChecking,
	}
	client, err := gossh.ConnectWithKeyCert(clientConfig)

	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Fatalln(err)
	}
	defer session.Close()

	r, err := session.StdoutPipe()
	if err != nil {
		log.Fatalln(err)
	}

	if err := session.Run("ls -lah"); err != nil {
		log.Fatalln(err)
	}

	buf, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(string(buf[0:]))
}
