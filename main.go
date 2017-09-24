package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
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

	// Read and parse private key
	log.Printf("Using private key /home/%s/.ssh/id_rsa", localUser)
	privateKeyBuf, err := ioutil.ReadFile("/home/" + localUser + "/.ssh/id_rsa")
	if err != nil {
		log.Fatalln("Failed to read private key")
	}

	privateKey, err := ssh.ParsePrivateKey(privateKeyBuf)
	if err != nil {
		log.Fatalf("Failed to parse private key")
	}

	// Load in, process and use public key certificate as public key.
	// This is important for SSH login using public key with principals.
	// Relevant functions derived from http://grokbase.com/t/gg/golang-nuts/157889kt3k/go-nuts-ssh-certificate-parseceritificate
	pubKeyCertBuf, err := ioutil.ReadFile("/home/" + localUser + "/.ssh/id_rsa-cert.pub")
	if err != nil {
		log.Println(err)
		log.Fatalln("Fail to read SSH key cert file")
	}

	pubKeyCert, _, _, _, err := ssh.ParseAuthorizedKey(pubKeyCertBuf)
	if err != nil {
		log.Println(err)
		log.Println("Failed to parse cert file")
	}

	cert := pubKeyCert.(*ssh.Certificate)
	log.Printf("Public key certificate: %v", cert)

	// Use private key to verify the signed certificate's public key
	// and get the signer that uses the certificate as public key
	// for the connection
	certSigner, err := ssh.NewCertSigner(cert, privateKey)
	if err != nil {
		log.Println(err)
		log.Println("Failed to make cert signer")
	}

	// Set up the client configuration
	config := ssh.ClientConfig{
		User: remoteUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(certSigner),
		},

		// This function verifies the remote host's host key against
		// global/local known_hosts file's public keys or host CA public key
		// Important for protecting against man-in-the-middle attacks
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			hostCert := key.(*ssh.Certificate)
			log.Printf("Host certificate: %v", hostCert)
			// TODO: Implement the actual verification
			return nil
		},
	}

	// Connect to the remote host
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), &config)
	if err != nil {
		log.Println(err)
		log.Fatalf("Failed to connect to remote server")
	} else {
		log.Println("Connection successful")
	}

	// Close connection after we are done
	defer conn.Close()
}
