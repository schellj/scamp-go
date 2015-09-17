package main

// import "scamp"
import "fmt"
import "flag"
import "io/ioutil"

import (
  "github.com/gudtech/scamp-go/scamp"
	"crypto/x509"
	"encoding/pem"
)

var announcePath string
var certPath string
var keyPath string

func main() {
	scamp.Initialize()

	flag.StringVar(&announcePath, "announcepath", "", "payload to be signed")
	flag.StringVar(&certPath, "certpath", "", "path to cert used for signing")
	flag.StringVar(&keyPath, "keypath", "", "path to service private key")
	flag.Parse()

	if len(keyPath) == 0 || len(announcePath) == 0 || len(certPath) == 0 {
		fmt.Println("must provide all 3: certpath, keypath, and announcepath")
		return
	}

	keyRawBytes,err := ioutil.ReadFile(keyPath)
	if err != nil {
		scamp.Error.Fatalf("could not read key at %s", keyPath)
	}

	block,_ := pem.Decode(keyRawBytes)

	if block == nil {
		scamp.Error.Fatalf("could not decode key data (%s)", block.Type)
		return
	} else if block.Type != "RSA PRIVATE KEY" {
		scamp.Error.Fatalf("expected key type 'RSA PRIVATE KEY' but got '%s'", block.Type)
	}

	privKey,err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		scamp.Error.Fatalf("could not parse key from %s (%s)", keyPath, block.Type)
	}

	announceData,err := ioutil.ReadFile(announcePath)
	if err != nil {
		scamp.Error.Fatalf("could not read announce data from %s", announcePath)
	}
	announceSig,err := scamp.SignSHA256( []byte(announceData), privKey)
	if err != nil {
		scamp.Error.Fatalf("could not sign announce data: %s", err)
	}

	certData,err := ioutil.ReadFile(certPath)
	if err != nil {
		scamp.Error.Fatalf("could not read cert from %s", certPath)
	}

	fmt.Printf("\n%%%%%%\n%s\n%s\n%s", announceData, certData, announceSig)

	scamp.Trace.Printf("cool announceSig: %s", announceSig)
}