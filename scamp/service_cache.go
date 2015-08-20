package scamp

import "errors"
import "fmt"
import "bufio"
import "bytes"


type actionName string
type services []*ServiceProxy

type ServiceCache struct {
	nameIndex map[actionName]services
}

func NewServiceCache() (cache *ServiceCache) {
	cache = new(ServiceCache)
	cache.nameIndex = make(map[actionName]services)
	return
}

func (cache *ServiceCache) Store( action string, instance *ServiceProxy ) {
	actType := actionName(action)
	matchingServices,ok := cache.nameIndex[actType]
	if !ok {
		matchingServices = make([]*ServiceProxy, 0, 10)
		matchingServices = append(matchingServices, instance)
		cache.nameIndex[actType] = matchingServices
	}

	return
}

func (cache *ServiceCache) Retrieve( action string ) ( instance *ServiceProxy ) {
	actType := actionName(action)
	matchingServices,ok := cache.nameIndex[actType]
	if !ok {
		instance = nil
		return
	} else if len(matchingServices) == 0 {
		instance = nil
		return
	}

	instance = matchingServices[0]
	return
}

var startCert = []byte(`-----BEGIN CERTIFICATE-----`)
var endCert = []byte(`-----END CERTIFICATE-----`)
func scanCertficates(data []byte, atEOF bool) (advance int, token []byte, err error) {
	var i int

	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// assert cert start line
	if i = bytes.Index(data, startCert); i == -1{
		return 0, nil, nil
	}

	// assert end line, consume if present
	if i = bytes.Index(data, endCert); i >= 0 {
		return i+len(endCert), data[0:i+len(endCert)], nil
	} else {
		return 0, nil ,nil
	}
}

var sep = []byte(`%%%`)
var newline = []byte("\n")

func (cache *ServiceCache) LoadAnnounceCache(s *bufio.Scanner) (err error) {
	for {
		for s.Scan() {
			fmt.Printf("slop: %s\n", s.Bytes())
			if bytes.Equal(s.Bytes(), sep) {
				break
			}
		}
		s.Scan() // consume the separator

		if len(s.Bytes()) == 0 {
			err = errors.New("expected class records after separator, found empty line")
			return
		}
		classRecords := make([]byte, len(s.Bytes()))
		copy(classRecords, s.Bytes())
		fmt.Printf("classRecords: %s\n", classRecords)
		s.Scan() // consume the classRecords

		if len(s.Bytes()) != 0 {
			err = errors.New("expected newline after class records")
			return
		}

		var certBuffer bytes.Buffer
		for s.Scan() {
			if len(s.Bytes()) == 0 {
				break
			}
			certBuffer.Write(s.Bytes())
			certBuffer.Write(newline)
		}
		cert := certBuffer.Bytes()
		fmt.Printf("cert: `%s`\n", cert)

		// if len(s.Bytes()) != 0 {
		// 	err = errors.New("expected new line after cert")
		// 	return
		// }
		// s.Scan()

		var sigBuffer bytes.Buffer
		for s.Scan() {
			if len(s.Bytes()) == 0 {
				break
			}
			sigBuffer.Write(s.Bytes())
			sigBuffer.Write(newline)
		}
		sig := sigBuffer.Bytes()[0:len(sigBuffer.Bytes())-1]
		fmt.Printf("sig: `%s`\n", sig)
		s.Scan()

		if len(s.Bytes()) != 0 {
			err = errors.New("expected new line after signature")
			return
		}
	}



	return
}

// *** SERVICE PROXY ***

type ServiceProxy struct {
	name string
	connspec string
	conn *Connection
}

func NewServiceProxy(connspec string) (proxy *ServiceProxy) {
	proxy = new(ServiceProxy)
	proxy.connspec = connspec
	proxy.conn = nil // we connect on demand
	return
}

func (proxy *ServiceProxy)GetConnection() (conn *Connection, err error) {
	if proxy.conn != nil {
		conn = proxy.conn
		return
	}

	proxy.conn, err = Connect(proxy.connspec)
	if err != nil {
		return
	}

	return
}
