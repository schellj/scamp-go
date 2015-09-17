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

func main() {
	scamp.Initialize()

	var announceData string
	flag.StringVar(&announceData, "announce", "", "payload to be signed")

	// var certPath string
	// flag.StringVar(&certPath, "signer", "", "path to cert used for signing")
	var keyPath string
	flag.StringVar(&keyPath, "keypath", "", "path to service private key")

	flag.Parse()

	// if len(certPath) == 0 || len(announceData) == 0 {
	if len(keyPath) == 0 || len(announceData) == 0 {
		fmt.Println("must provide keypath and announce")
		return
	}

	// certData,err := ioutil.ReadFile(certPath)
	// if err != nil {
	// 	fmt.Println("could not read certPath")
	// 	return
	// }

	// decodedCertData,_ := pem.Decode(certData)
	// if decodedCertData == nil {
	// 	fmt.Println("could not decode cert data")
	// 	return
	// }

	// cert,err := x509.ParseCertificate(decodedCertData.Bytes)
	// if err != nil {
	// 	fmt.Printf("could not parse certificate: `%s`", err)
	// 	return
	// }

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

	announceSig,err := scamp.SignSHA256( []byte(announceData), privKey)

	scamp.Trace.Printf("cool announceSig: %s", announceSig)
}