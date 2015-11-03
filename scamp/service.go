package scamp

import "errors"
import "net"
import "crypto/tls"
import "crypto/rand"
import "crypto/rsa"
import "encoding/base64"
import "encoding/json"
import "fmt"
import "bytes"
import "io/ioutil"

type ServiceActionFunc func(Request,*Session)
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
	sessChan      (chan *Session)
	isRunning     bool
	openConns     []*Connection

	cert          tls.Certificate
	pemCert       []byte // just a copy of what was read off disk at tls cert load time
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
	serv.sessChan = make(chan *Session, 100)

	crtPath := defaultConfig.ServiceCertPath(serv.humanName)
	keyPath := defaultConfig.ServiceKeyPath(serv.humanName)

	if crtPath == nil || keyPath == nil {
		err = fmt.Errorf( "could not find valid crt/key pair for service %s (`%s`,`%s`)", serv.name, crtPath, keyPath )
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
  serv.listenerIP = serv.listener.Addr().(*net.TCPAddr).IP
  Trace.Printf("serv.listenerIP: `%s`", serv.listenerIP)
  // serv.listenerIP, err = IPForAnnouncePacket()
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
	go serv.RouteSessions()

	for {
		netConn,err := serv.listener.Accept()
		if err != nil {
			Info.Printf("exiting service service Run(): `%s`", err)
			break
		}

		var tlsConn (*tls.Conn) = (netConn).(*tls.Conn)
		if tlsConn == nil {
			Error.Fatalf("could not create tlsConn")
			break
		}

		conn,err := newConnection(tlsConn, serv.sessChan)
		if err != nil {
			Error.Fatalf("error with new connection: `%s`", err)
			break
		}
		serv.openConns = append(serv.openConns, conn)

		go conn.packetRouter(false, true)
	}

	close(serv.sessChan)
}

// Spawn a router for each new session received over sessChan
func (serv *Service)RouteSessions() (err error){

	for newSess := range serv.sessChan {
		
		// if !stillOpen {
		// 	Trace.Printf("sessChan was closed. server is probably shutting down.")
		// 	break
		// }

		go func(){
			var action *ServiceAction

			request,err := newSess.RecvRequest()
			if err != nil {
				return
			}

			action = serv.actions[request.Action]
			if action != nil {
				action.callback(request, newSess)
			} else {
				Error.Printf("unknown action `%s`", request.Action)
				// TODO: need to respond with 'unknown action'
			}
		}()
	}

	return
}

func (serv *Service)Stop(){
	// Sometimes we Stop() before service after service has been init but before it is started
	// The usual case is a bad config in another plugin
	if serv.listener != nil {
		serv.listener.Close()
	}
	for _,conn := range serv.openConns {
		conn.Close()
	}
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