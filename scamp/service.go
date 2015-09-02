package scamp

import "errors"
import "net"
import "crypto/tls"

type ServiceAction func(Request,*Session)

type Service struct {
	listener      net.Listener
	actions       map[string]ServiceAction
	sessChan      (chan *Session)
	isRunning     bool
	openConns     []*Connection
}

func NewService(serviceSpec string) (serv *Service, err error){
	serv = new(Service)
	serv.actions = make(map[string]ServiceAction)
	serv.sessChan = make(chan *Session, 100)

	err = serv.listen(serviceSpec)
	if err != nil {
		return
	}

	return
}

func (serv *Service)listen(serviceSpec string) (err error) {
	cert, err := tls.LoadX509KeyPair( "/etc/SCAMP/services/helloworld.crt","/etc/SCAMP/services/helloworld.key" )
	if err != nil {
		return
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{ cert },
	}

	Trace.Printf("starting service on %s", serviceSpec)

	serv.listener,err = tls.Listen("tcp", serviceSpec, config)
	if err != nil {
		return err
	}

	return
}

func (serv *Service)Register(name string, action ServiceAction) (err error) {
	if serv.isRunning {
		return errors.New("cannot register handlers while server is running")
	}

	serv.actions[name] = action

	return
}

func (serv *Service)Run() {
	go serv.RouteSessions()

	for {
		netConn,err := serv.listener.Accept()
		if err != nil {
			Info.Printf("accept returned error. exiting service Run()")
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
			var action ServiceAction

			Trace.Printf("waiting for request to be received")
			request,err := newSess.RecvRequest()
			if err != nil {
				return
			}
			Trace.Printf("request came in for action `%s`", request.Action)

			action = serv.actions[request.Action]
			if action != nil {
				action(request, newSess)
				newSess.Free()
			} else {
				Error.Printf("unknown action `%s`", request.Action)
				// TODO: need to respond with 'unknown action'
			}
		}()
	}

	return
}

func (serv *Service)Stop(){
	serv.listener.Close()
	for _,conn := range serv.openConns {
		conn.Close()
	}
}
