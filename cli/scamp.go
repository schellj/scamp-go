package main

// import "scamp"
import "fmt"
import "flag"
import "io/ioutil"

import (
	"crypto/x509"
	"encoding/pem"
)

func main() {
	var announceData string
	flag.StringVar(&announceData, "announce", "", "payload to be signed")

	var certPath string
	flag.StringVar(&certPath, "keypath", "", "path to key used for signing")

	flag.Parse()

	if len(certPath) == 0 || len(announceData) == 0 {
		fmt.Println("must provide certPath and announceData")
		return
	}

	certData,err := ioutil.ReadFile(certPath)
	if err != nil {
		fmt.Println("could not read certPath")
		return
	}

	decodedCertData,_ := pem.Decode(certData)
	if decodedCertData == nil {
		fmt.Println("could not decode cert data")
		return
	}

	cert,err = x509.ParseCertificate(decodedCertData.Bytes)
	if err != nil {
		fmt.Printf("could not parse certificate: `%s`", err)
		return
	}

	scamp.

	fmt.Printf("%s %s\n", announceData, certPath)
}