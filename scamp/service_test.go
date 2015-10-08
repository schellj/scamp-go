package scamp

import "testing"
import "time"
import "bytes"
import "encoding/json"
import "net"
import "crypto/tls"
import "io/ioutil"

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

// TODO: I'm cutting some corners in this test, it tests two complicated things at once:
// 1. Copying `Service` properties to new `ServiceProxy`
// 2. Marshaling `ServiceProxy` to announce format
func TestServiceToProxyMarshal(t *testing.T) {
	s := Service {
		serviceSpec: "123",
		humanName: "a-cool-name",
		name: "a-cool-name-1234",
		listenerIP: net.ParseIP("174.10.10.10"),
		listenerPort: 30100,
		actions: make(map[string]*ServiceAction),
	}
	s.Register("Logging.info", func(_ Request, _ *Session) {
	})

	serviceProxy := ServiceAsServiceProxy(&s)
	serviceProxy.timestamp = 10
	b,err := json.Marshal(&serviceProxy)
	if err != nil {
		t.Fatalf("could not serialize service proxy")
	}
	expected := []byte(`[3,"a-cool-name-1234","sector",1,2500,"beepish+tls://174.10.10.10:30100",["json"],[["Logging",["info","",1]]],10.000000]`)
	if !bytes.Equal(b, expected) {
		t.Fatalf("expected: `%s`,\n             got:      `%s`", expected, b)
	}

}

func TestFullServiceMarshal(t *testing.T) {
	// TODO big assumption that you environment is set up like mine:
	//   root repo `scamp-go` has a sibling folder called `scamp-go-workspace` where `scamp-go`
	//   is symlinked in as such: ../scamp-go-workspace/src/github.com/gudtech/scamp-go
	// it's crazy, I know. thanks GOPATH.
	cert, err := tls.LoadX509KeyPair( "./../../scamp-go/fixtures/sample.crt", "./../../scamp-go/fixtures/sample.key" )
	if err != nil {
		t.Fatalf("could not load fixture keypair: `%s`", err)
	}

	encodedCert,err := ioutil.ReadFile("/Users/xavierlange/code/gudtech/scamp-go/fixtures/sample.crt")
	if err != nil {
		t.Fatalf("could not load fixture certificate")
	}
	encodedCert = bytes.TrimSpace(encodedCert)

	s := Service {
		serviceSpec: "123",
		humanName: "a-cool-name",
		name: "a-cool-name-1234",
		listenerIP: net.ParseIP("174.10.10.10"),
		listenerPort: 30100,
		actions: make(map[string]*ServiceAction),
		pemCert: encodedCert,
		cert: cert,
	}
	s.Register("Logging.info", func(_ Request, _ *Session) {
	})

	// TODO: confirm output of marshalling the payload.
	_,err = s.MarshalText()
	if err != nil {
		t.Fatalf("unexpected error serializing service: `%s`", err)
	}
	// t.Fatalf("b: `%s`", b)

}
