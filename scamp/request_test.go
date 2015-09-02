package scamp

import "testing"
import "fmt"

func TestGenerateMessageId(t *testing.T) {
	req := Request{}
	req.GenerateMessageId()
	if req.MessageId == "" || len(req.MessageId) != 18 {
		t.Fatalf("MesasageId should have been 18-byte string but got `%s`", req.MessageId)
		t.FailNow()
	}
}

func TestHeaderRequestToPackets(t *testing.T) {
	req := Request{
		Action:         "hello.helloworld",
		EnvelopeFormat: ENVELOPE_JSON,
		Version:        1,
	}

	if req.MessageId != "" {
		t.Fatalf("MessageId should not be set")
	}

	pkts := req.toPackets(0)
	if len(pkts) != 2 {
		t.Fatalf("expected 2 packet")
	}
	if req.MessageId == "" {
		panic( fmt.Sprintf("req.MID: %s", req.MessageId) )
		t.Fatalf("MessageId was not set by toPackets")
	}

	hdrPkt := pkts[0]
	if hdrPkt.packetType != HEADER {
		t.Fatalf("expected HEADER type")
	}
	if hdrPkt.msgNo != 0 {
		t.Fatalf("header msgNo was %d but expected %d", hdrPkt.msgNo, 0)
	}
	expectedHeader := PacketHeader{
		Action:    "hello.helloworld",
		Version:   1,
		MessageId: req.MessageId,
	}
	if hdrPkt.packetHeader != expectedHeader {
		t.Fatalf("packetHeader was `%v` but expected `%v`", hdrPkt.packetHeader, expectedHeader)
	}

	eofPkt := pkts[1]
	if eofPkt.packetType != EOF {
		t.Fatalf("expected EOF type")
	}
	if eofPkt.msgNo != 0 {
		t.Fatalf("eof msgNo was %d but expected %d", eofPkt.msgNo, 0)
	}
}
