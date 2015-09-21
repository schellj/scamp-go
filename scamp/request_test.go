package scamp

import "testing"
import "fmt"

func TestHeaderRequestToPackets(t *testing.T) {
	req := Request{
		Action:         "hello.helloworld",
		EnvelopeFormat: ENVELOPE_JSON,
		Version:        1,
		MessageId:      0,
	}

	pkts := req.toPackets(1)
	if len(pkts) != 2 {
		t.Fatalf("expected 2 packet")
	}
	if req.MessageId == 1 {
		panic( fmt.Sprintf("req.MID: %s", req.MessageId) )
		t.Fatalf("MessageId was not set by toPackets")
	}

	hdrPkt := pkts[0]
	if hdrPkt.packetType != HEADER {
		t.Fatalf("expected HEADER type")
	}
	if hdrPkt.msgNo != 1 {
		t.Fatalf("header msgNo was %d but expected %d", hdrPkt.msgNo, 1)
	}
	expectedHeader := PacketHeader{
		Action:    "hello.helloworld",
		Envelope:  ENVELOPE_JSON,
		Version:   1,
		MessageId: 1,
	}
	if hdrPkt.packetHeader != expectedHeader {
		t.Fatalf("packetHeader was `%v` but expected `%v`", hdrPkt.packetHeader, expectedHeader)
	}

	eofPkt := pkts[1]
	if eofPkt.packetType != EOF {
		t.Fatalf("expected EOF type")
	}
	if eofPkt.msgNo != 1 {
		t.Fatalf("eof msgNo was %d but expected %d", eofPkt.msgNo, 1)
	}
}
