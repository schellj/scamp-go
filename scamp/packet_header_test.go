package scamp

import "testing"
import "bytes"
import "encoding/json"

func TestEncodeEnvelope(t *testing.T) {
	expected := []byte("\"json\"\n")

	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)

	val := ENVELOPE_JSON
	err := encoder.Encode(val)

	if err != nil {
		t.Errorf("got unexpected err `%s`\n", err)
		t.FailNow()
	}
	if !bytes.Equal(expected, buf.Bytes()) {
		t.Errorf("expected `%s` but got `%s`", expected, buf.Bytes())
		t.FailNow()
	}
}

func TestWritePacketHeader(t *testing.T) {
	packetHeader := PacketHeader{
		Action:    "hello.helloworld",
		Envelope:  ENVELOPE_JSON,
		MessageType: MESSAGE_TYPE_REQUEST,
		RequestId: 1,
		Version:   1,
	}
	expected := []byte(`{"action":"hello.helloworld","envelope":"json","request_id":1,"type":"request","version":1}
`)

	buf := new(bytes.Buffer)
	err := packetHeader.Write(buf)
	if err != nil {
		t.Fatalf("unexpected error when serializing packet header: `%s`", err)
	}

	if !bytes.Equal(expected, buf.Bytes()) {
		t.Fatalf("expected\n`%s`\n`%v`\ngot\n`%s`\n`%v`\n", expected, expected, buf.Bytes(), buf.Bytes())
	}
}

func TestDecodePacketHeader(t *testing.T) {
	pkt_hdr_bytes := []byte("{\"envelope\": \"json\"}")
	buf := bytes.NewReader(pkt_hdr_bytes)
	decoder := json.NewDecoder(buf)

	var pktHeader PacketHeader
	err := decoder.Decode(&pktHeader)
	if err != nil {
		t.Errorf("unexpected error while decoding JSON. got `%s`", err)
		t.FailNow()
	}
}

func TestDecodePacketHeaderReply(t *testing.T) {
	pkt_hdr_bytes := []byte("{\"type\":\"reply\",\"request_id\":1}")
	buf := bytes.NewReader(pkt_hdr_bytes)
	decoder := json.NewDecoder(buf)

	var pktHeader PacketHeader
	err := decoder.Decode(&pktHeader)
	if err != nil {
		t.Errorf("unexpected error while decoding JSON. got `%s`", err)
		t.FailNow()
	}
}
