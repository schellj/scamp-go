package scamp

import (
	"errors"
	"net"
	"crypto/tls"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"bytes"
	"io/ioutil"
	"time"

	"sync/atomic"
)

// Two minute timeout on clients
var msgTimeout = time.Second * 120

type ServiceActionFunc func(*Message, *Client)
type ServiceAction struct {
	callback ServiceActionFunc
	crudTags string
	version  int
}

type Service struct {
	serviceSpec   string
	name          string
	humanName     string

	listener      net.Listener
	listenerIP    net.IP
	listenerPort  int

	actions       map[string]*ServiceAction
	isRunning     bool
	clients       []*Client

	// requests      ClientChan

	cert          tls.Certificate
	pemCert       []byte // just a copy of what was read off disk at tls cert load time

	// stats
	statsCloseChan chan bool
	connectionsAccepted uint64
}

func NewService(serviceSpec string, humanName string) (serv *Service, err error){
	if len(humanName) > 18 {
		err = fmt.Errorf("name `%s` is too long, must be less than 18 bytes", humanName)
		return
	}

	serv = new(Service)
	serv.serviceSpec = serviceSpec
	serv.humanName = humanName
	serv.generateRandomName()

	serv.actions = make(map[string]*ServiceAction)

	crtPath := defaultConfig.ServiceCertPath(serv.humanName)
	keyPath := defaultConfig.ServiceKeyPath(serv.humanName)

	if crtPath == nil || keyPath == nil {
		err = fmt.Errorf( "could not find valid crt/key pair for service %s (`%s`,`%s`)", serv.humanName, crtPath, keyPath )
		return
	}

	// Load keypair for tls socket library to use
	serv.cert, err = tls.LoadX509KeyPair( string(crtPath), string(keyPath) )
	if err != nil {
		return
	}

	// Load cert in to memory for announce packet writing
	serv.pemCert, err = ioutil.ReadFile(string(crtPath))
	if err != nil {
		return
	}
	serv.pemCert = bytes.TrimSpace(serv.pemCert)

	// Finally, get ready for incoming requests
	err = serv.listen()
	if err != nil {
		return
	}

	serv.statsCloseChan = make(chan bool)
	go PrintStatsLoop(serv, time.Duration(15)*time.Second, serv.statsCloseChan)

	Trace.Printf("done initializing service")

	return
}

func (serv *Service)generateRandomName() {
	randBytes := make([]byte, 18, 18)
	read,err := rand.Read(randBytes)
	if err != nil {
		err = fmt.Errorf("could not generate all rand bytes needed. only read %d of 18", read)
		return
	}
	base64RandBytes := base64.StdEncoding.EncodeToString(randBytes)

	var buffer bytes.Buffer
	buffer.WriteString(serv.humanName)
	buffer.WriteString("-")
	buffer.WriteString(base64RandBytes[0:])
	serv.name = string(buffer.Bytes())
}

// TODO: port discovery and interface/IP discovery should happen here
// important to set values so announce packets are correct
func (serv *Service)listen() (err error) {
	config := &tls.Config{
		Certificates: []tls.Certificate{ serv.cert },
	}

	Info.Printf("starting service on %s", serv.serviceSpec)
	serv.listener,err = tls.Listen("tcp", serv.serviceSpec, config)
	if err != nil {
		return err
	}

  // TODO: get listenerIP to return 127.0.0.1 or something other than '::'/nil
  // serv.listenerIP = serv.listener.Addr().(*net.TCPAddr).IP
  serv.listenerIP, err = IPForAnnouncePacket()
  Trace.Printf("serv.listenerIP: `%s`", serv.listenerIP)
  if err != nil {
  	return
  }
  
	serv.listenerPort = serv.listener.Addr().(*net.TCPAddr).Port

	return
}

func (serv *Service)Register(name string, callback ServiceActionFunc) (err error) {
	if serv.isRunning {
		return errors.New("cannot register handlers while server is running")
	}

	serv.actions[name] = &ServiceAction {
		callback: callback,
		version: 1,
	}

	return
}

func (serv *Service)Run() {

	for {
		netConn,err := serv.listener.Accept()
		Trace.Printf("accepted new connection...")
		if err != nil {
			Info.Printf("exiting service service Run(): `%s`", err)
			break
		}

		var tlsConn (*tls.Conn) = (netConn).(*tls.Conn)
		if tlsConn == nil {
			Error.Fatalf("could not create tlsConn")
			break
		}

		conn := NewConnection(tlsConn)
		client := NewClient(conn)

		serv.clients = append(serv.clients, client)
		go serv.Handle(client)

		atomic.AddUint64(&serv.connectionsAccepted, 1)
	}
}

func (serv *Service)Handle(client *Client) {
	var action *ServiceAction

	HandlerLoop:
	for {
		select {
		case msg,ok := <-client.Incoming():
			if !ok {
				break HandlerLoop
			}
			// Trace.Printf("msg!!!! `%s`", msg)
			// Trace.Printf("action: `%s`", msg.Action)
			action = serv.actions[msg.Action]

			if action != nil{
				// yay
				action.callback(msg, client)
			} else {
				// TODO: gotta tell them I don't know how to do that
			}
		case <- time.After(msgTimeout):
			Error.Printf("timeout... dying!")
			client.Close()
			serv.RemoveClient(client)
			break HandlerLoop
		}
	}

	Trace.Printf("done handling client")
	
}

func (serv *Service)RemoveClient(client *Client) (err error){
	index := -1
	for i,entry := range serv.clients {
		if client == entry {
			index = i
			break
		}
	}

	if index == -1 {
		Error.Printf("tried removing client that wasn't being tracked")
		return fmt.Errorf("unknown client") // TODO can I get the client's IP?
	}

	serv.clients = append(serv.clients[:index], serv.clients[index+1:]...)
	return nil
}

func (serv *Service)Stop(){
	// Sometimes we Stop() before service after service has been init but before it is started
	// The usual case is a bad config in another plugin
	if serv.listener != nil {
		serv.listener.Close()
	}
	for _,conn := range serv.clients {
		conn.Close()
	}

	serv.statsCloseChan <- true
}

func (serv *Service)MarshalText() (b []byte, err error){
	var buf bytes.Buffer

	serviceProxy := ServiceAsServiceProxy(serv)
	classRecord,err := json.Marshal(&serviceProxy)
	if err != nil {
		return
	}

	sig, err := SignSHA256(classRecord, serv.cert.PrivateKey.(*rsa.PrivateKey))
	if err != nil {
		return
	}

	buf.Write(classRecord)
	buf.WriteString("\n\n")
	buf.Write(serv.pemCert)
	buf.WriteString("\n\n")
	buf.WriteString(sig)

	b = buf.Bytes()
	return
}
