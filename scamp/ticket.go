package scamp

import "bytes"
import "errors"
import "encoding/base64"
import "strconv"
import "fmt"

import "encoding/pem"
import "crypto"
import "crypto/rsa"
import "crypto/sha256"
import "crypto/x509"

type Ticket struct {
	Version int64
	UserId  int64
	ClientId int64
	ValidityStart int64
	ValidityEnd   int64

	Ttl int
	Expired bool
}

var separator = []byte(",")
var supportedVersion = []byte("1")
var padding = []byte("=")

func ReadTicket(incoming []byte, signingPubKey []byte) (ticket Ticket, err error) {
	rsaPubKey, err := parseRsaPubKey(signingPubKey)
	if err != nil {
		return
	}

	ticketBytes, signature := splitTicketPayload(incoming)

	decodedSignature,err := decodeUnpaddedBase64(signature)
	if err != nil {
		return
	}

	h := sha256.New()
	h.Write(ticketBytes)
	digest := h.Sum(nil)

	err = rsa.VerifyPKCS1v15(rsaPubKey, crypto.SHA256, digest, decodedSignature)
	if err != nil {
		return
	}

	ticket,err = parseTicketBytes(ticketBytes)
	if err != nil {
		return
	}

	return
}

func ReadTicketNoVerify(incoming []byte) (ticket Ticket, err error) {
	ticketBytes,_ := splitTicketPayload(incoming)
	return parseTicketBytes(ticketBytes)
}

func splitTicketPayload(incoming []byte) (ticketBytes []byte, ticketSig []byte) {
	lastIndex := bytes.LastIndex(incoming, separator)
	ticketBytes = incoming[:lastIndex]
	ticketSig = incoming[lastIndex+1:]
	return
}

func parseRsaPubKey(signingPubKey []byte) (rsaPubKey *rsa.PublicKey, err error) {
	block, _ := pem.Decode(signingPubKey)
	if block == nil {
		err = errors.New("expected valid block")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return
	}

	rsaPubKey, ok := key.(*rsa.PublicKey)
	if !ok {
		err = errors.New("could not cast parsed value to rsa.PublicKey")
		return
	}

	return
}

func parseTicketBytes(ticketBytes []byte) (ticket Ticket, err error) {
	chunks := bytes.Split(ticketBytes, separator)

	if !bytes.Equal(chunks[0], supportedVersion) {
		err = errors.New("ticket must be version 1")
		return
	}

	ticket.Version,err = strconv.ParseInt(string(chunks[0]), 10, 0)
	if(err != nil){
		return
	}

	ticket.UserId,err = strconv.ParseInt(string(chunks[1]), 10, 0)
	if(err != nil){
		return
	}

	ticket.ClientId,err = strconv.ParseInt(string(chunks[2]), 10, 0)
	if(err != nil){
		return
	}

	ticket.ValidityStart,err = strconv.ParseInt(string(chunks[3]), 10, 0)
	if(err != nil){
		return
	}

	validityDuration,err := strconv.ParseInt(string(chunks[4]), 10, 0)
	if(err != nil){
		return
	}
	ticket.ValidityEnd = ticket.ValidityStart + validityDuration

	return
}

func decodeUnpaddedBase64(incoming []byte) (decoded []byte, err error) {
	if m := len(incoming) % 4; m != 0 {
		paddingBytes := bytes.Repeat(padding, 4-m)
		incoming = append(incoming, paddingBytes[:]...)
	}

	decoded,err = base64.URLEncoding.DecodeString(string(incoming))
	if(err != nil){
		err = errors.New( fmt.Sprintf("err: `%s` could not decode `%s`", err, incoming) )
		return
	}

	return
}