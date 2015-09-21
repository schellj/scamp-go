package main

// import "scamp"
import "fmt"
import "flag"
import "io/ioutil"
import "bytes"

import (
  "github.com/gudtech/scamp-go/scamp"
	"crypto/x509"
	"encoding/pem"
)

var announcePath string
var certPath string
var keyPath string
var fingerprintPath string

func main() {
	scamp.Initialize()

	flag.StringVar(&announcePath, "announcepath", "", "payload to be signed")
	flag.StringVar(&certPath, "certpath", "", "path to cert used for signing")
	flag.StringVar(&keyPath, "keypath", "", "path to service private key")
	flag.StringVar(&fingerprintPath, "fingerprintpath", "", "path to cert to fingerprint")
	flag.Parse()

	if (len(keyPath) == 0 || len(announcePath) == 0 || len(certPath) == 0) && (len(fingerprintPath) == 0) {
		fmt.Printf("fingerprintpath: %s", fingerprintPath)
		fmt.Println("not enough options specified\nmust provide\n\tcertpath, keypath, and announcepath\nOR\n\tfingerprintpath")
		return
	}

	if len(keyPath) != 0 {
		doFakeDiscoveryCache()
	} else {
		doCertFingerprint()
	}

}

func doFakeDiscoveryCache(){
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

	fmt.Printf("\n%%%%%%\n%s\n\n%s\n\n%s\n", announceData, bytes.TrimSpace(certData), announceSig)
}

func doCertFingerprint(){
	certData,err := ioutil.ReadFile(fingerprintPath)
	if err != nil {
		scamp.Error.Fatalf("could not read cert from %s", fingerprintPath)
	}

	decoded,_ := pem.Decode(certData)
	if decoded == nil {
		scamp.Error.Fatalf("could not decode cert. is it PEM encoded?")
	}

	// Put pem in form useful for fingerprinting
	cert,err := x509.ParseCertificate(decoded.Bytes)
	if err != nil {
		scamp.Error.Fatalf("could not parse certificate. is it valid x509?")
	}

	fingerprint := scamp.SHA1FingerPrint(cert)
	if len(fingerprint) > 0 {
		fmt.Printf("fingerprint: %s\n", fingerprint)
	} else {
		scamp.Error.Fatalf("could not fingerprint certificate")
	}
}