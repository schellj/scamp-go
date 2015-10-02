package scamp

import "testing"
import "time"
import "bytes"

// TODO: fix Session API (aka, simplify design by dropping it)
// func TestServiceHandlesRequest(t *testing.T) {
// 	Initialize()

// 	hasStopped := make(chan bool, 1)
// 	service := spawnTestService(hasStopped)
// 	connectToTestService(t)
// 	service.Stop()
// 	<-hasStopped

// }

func spawnTestService(hasStopped (chan bool)) (service *Service) {
	service,err := NewService(":30100", "helloworld")
	if err != nil {
		Error.Fatalf("error creating new service: `%s`", err)
	}
	service.Register("helloworld.hello", func(req Request, sess *Session){
		if len(req.Blob) > 0 {
			Info.Printf("helloworld had data: %s", req.Blob)
		} else {
			Trace.Printf("helloworld was called without data")
		}

		err = sess.Send(Reply{
			Blob: []byte("sup"),
		})
		if err != nil {
			Error.Printf("error while sending reply: `%s`. continuing.", err)
			return
		}
		Trace.Printf("successfully responded to hello world")
	})

	go func(){
		service.Run()
		hasStopped <- true
	}()
	return
}

func connectToTestService(t *testing.T) {
	conn, err := Connect("127.0.0.1:30100")
	defer conn.Close()

	if err != nil {
		Error.Fatalf("could not connect! `%s`\n", err)
	}

	err = conn.Send(&Request{
		Action:         "helloworld.hello",
		EnvelopeFormat: ENVELOPE_JSON,
		Version:        1,
	})
	if err != nil {
		Error.Fatalf("error initiating session: `%s`", err)
		t.FailNow()
	}

	sess := conn.Recv()

	select {
		case msg := <-sess.RecvChan():
			reply,ok := msg.(Reply)
			if !ok {
				t.Errorf("expected reply")
			}
			
			if !bytes.Equal(reply.Blob, []byte("sup")) {
				t.Fatalf("did not get expected response `sup`")
			}
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("timed out waiting for response")
	}

	return
}
