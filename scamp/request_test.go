package scamp

import "testing"

func TestHeaderRequestToPackets(t *testing.T) {
	req := Request{
		Action:         "hello.helloworld",
		EnvelopeFormat: ENVELOPE_JSON,
		Version:        1,
		RequestId:      0,
	}

	pkts := req.toPackets()
	if len(pkts) != 2 {
		t.Fatalf("expected 2 packet")
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
		Envelope:  ENVELOPE_JSON,
		Version:   1,
		RequestId: 0,
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
