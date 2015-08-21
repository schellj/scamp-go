package scamp

import "errors"
import "bufio"
import "bytes"

// Assumptions:
// 1. Services offered on an instance do not change during life of instance
type ServiceCache struct {
	identIndex map[string]*ServiceProxy
}

func NewServiceCache() (cache *ServiceCache) {
	cache = new(ServiceCache)
	cache.identIndex = make(map[string]*ServiceProxy)
	return
}

func (cache *ServiceCache) Store( instance *ServiceProxy ) {
	_,ok := cache.identIndex[instance.ident]
	if !ok {
		cache.identIndex[instance.ident] = instance
	} else {
		// Not sure if this is a hard error yet
		Error.Printf("tried to store instance that was already tracked")
	}

	return
}

func (cache *ServiceCache) Retrieve( ident string ) ( instance *ServiceProxy ) {
	instance,ok := cache.identIndex[ident]
	if !ok {
		instance = nil
		return
	}

	return
}

func (cache *ServiceCache) Size() int {
	return len(cache.identIndex)
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
	// Scan through buf by lines according to this basic ABNF
	// (SLOP* SEP CLASSRECORD NL CERT NL SIG NL NL)*
	var classRecordsRaw, certRaw, sigRaw []byte
	for {
		var didScan bool
		for {
			didScan = s.Scan()
			// Trace.Printf("slop: %s", s.Bytes())
			if bytes.Equal(s.Bytes(), sep) || !didScan {
				break
			}
		}
		if !didScan {
			// Trace.Printf("break out")
			break;
		}
		s.Scan() // consume the separator

		if len(s.Bytes()) == 0 {
			err = errors.New("unexpected newline after separator")
			return
		}
		classRecordsRaw = make([]byte, len(s.Bytes()))
		copy(classRecordsRaw, s.Bytes())
		// Trace.Printf("classRecordsRaw: %s", s.Bytes())
		s.Scan() // consume the classRecords

		if len(s.Bytes()) != 0 {
			err = errors.New("expected newline after class records")
			return
		}

		var certBuffer bytes.Buffer
		for s.Scan() {
			// Trace.Printf("certBuf: %s", s.Bytes())
			if len(s.Bytes()) == 0 {
				break
			}
			certBuffer.Write(s.Bytes())
			certBuffer.Write(newline)
		}
		certRaw = certBuffer.Bytes()[0:len(certBuffer.Bytes())-1]

		var sigBuffer bytes.Buffer
		for s.Scan() {
			// Trace.Printf("sigBuf: %s", s.Bytes())
			if len(s.Bytes()) == 0 {
				break
			}
			sigBuffer.Write(s.Bytes())
			sigBuffer.Write(newline)
		}
		sigRaw = sigBuffer.Bytes()[0:len(sigBuffer.Bytes())-1]

		// Use those extracted value to make an instance
		serviceProxy,err := NewServiceProxy(classRecordsRaw, certRaw, sigRaw)
		if err != nil {
			return err
		}

		cache.Store(serviceProxy)
	}

	return
}
