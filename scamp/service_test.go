package scamp

import "testing"
import "time"
import "bytes"

func TestServiceHandlesRequest(t *testing.T) {
	Initialize()

	hasStopped := make(chan bool, 1)
	service := spawnTestService(hasStopped)
	connectToTestService(t)
	service.Stop()
	<-hasStopped

}

func spawnTestService(hasStopped (chan bool)) (service *Service) {
	service,err := NewService(":30100")
	if err != nil {
		Error.Fatalf("error creating new service: `%s`", err)
	}
	service.Register("helloworld.hello", func(req Request, sess *Session){
		if len(req.Blob) > 0 {
			Info.Printf("helloworld had data: %s", req.Blob)
		} else {
			Trace.Printf("helloworld was called without data")
		}

		err = sess.SendReply(Reply{
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

	var sess *Session

	sess, err = conn.Send(&Request{
		Action:         "helloworld.hello",
		EnvelopeFormat: ENVELOPE_JSON,
		Version:        1,
	})
	if err != nil {
		Error.Fatalf("error initiating session: `%s`", err)
		t.FailNow()
	}

	select {
		case msg := <-sess.RecvChan():
			reply,ok := msg.(Reply)
			if !ok {
				t.Errorf("expected reply")
			}
			
			if !bytes.Equal(reply.Blob, []byte("sup")) {
				t.Errorf("did not get expected response `sup`")
				t.FailNow()
			}
		case <-time.After(500 * time.Millisecond):
			t.Errorf("timed out waiting for response")
			t.FailNow()
	}

	return
}
