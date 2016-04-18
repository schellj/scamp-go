package scamp

import (
  "errors"
  "bufio"
  "bytes"
  "fmt"
  "sync"
  "os"
)

type ServiceCache struct {
	path string

	cacheM sync.Mutex
	identIndex map[string]*ServiceProxy
	verifyRecords bool
}

func NewServiceCache(path string) (cache *ServiceCache, err error) {
	cache = new(ServiceCache)
	cache.path = path

	cache.identIndex = make(map[string]*ServiceProxy)
	cache.verifyRecords = true
	return
}

func (cache *ServiceCache) DisableRecordVerification() (){
	cache.verifyRecords = true
}

func (cache *ServiceCache) EnableRecordVerification() (){
	cache.verifyRecords = false
}

func (cache *ServiceCache) Store( instance *ServiceProxy ) {
	cache.cacheM.Lock()
	defer cache.cacheM.Unlock()

	cache.storeNoLock( instance )

	return
}

func (cache *ServiceCache) storeNoLock( instance *ServiceProxy ) {
	_,ok := cache.identIndex[instance.ident]
	if !ok {
		cache.identIndex[instance.ident] = instance
	} else {
		// Not sure if this is a hard error yet
		// Error.Printf("tried to store instance that was already tracked")
		// Override existing version. Correct logic?
		cache.identIndex[instance.ident] = instance
	}

	return
}

func (cache *ServiceCache) removeNoLock( instance *ServiceProxy ) (err error) {
	_,ok := cache.identIndex[instance.ident]
	if !ok {
		err = fmt.Errorf("tried removing an ident which was not being tracked: %s", instance.ident)
		return
	}

	delete(cache.identIndex,instance.ident)

	return
}

func (cache *ServiceCache) Retrieve( ident string ) ( instance *ServiceProxy ) {
	cache.cacheM.Lock()
	defer cache.cacheM.Unlock()

	instance,ok := cache.identIndex[ident]
	if !ok {
		instance = nil
		return
	}

	return
}

func (cache *ServiceCache) Size() int {
	cache.cacheM.Lock()
	defer cache.cacheM.Unlock()

	return len(cache.identIndex)
}

func (cache *ServiceCache) All() (proxies []*ServiceProxy) {
	cache.cacheM.Lock()
	defer cache.cacheM.Unlock()

	size := len(cache.identIndex)
	proxies = make([]*ServiceProxy, size)

	index := 0
	for _,proxy := range cache.identIndex {
		proxies[index] = proxy
		index += 1
	}

	return
}


var sep = []byte(`%%%`)
var newline = []byte("\n")

func (cache *ServiceCache) Scan() (err error) {
	cache.cacheM.Lock()
	defer cache.cacheM.Unlock()

	stat,err := os.Stat(cache.path)
	if err != nil {
		return
	}
	Error.Printf("mtime: %s\n",stat.ModTime())


	cacheHandle,err := os.Open(cache.path)
	if err != nil {
		return
	}

  s := bufio.NewScanner(cacheHandle)

	// Scan through buf by lines according to this basic ABNF
	// (SLOP* SEP CLASSRECORD NL CERT NL SIG NL NL)*
	var classRecordsRaw, certRaw, sigRaw []byte
	for {
		var didScan bool
		for {
			didScan = s.Scan()
			if bytes.Equal(s.Bytes(), sep) || !didScan {
				break
			}
		}
		if !didScan {
			break;
		}
		s.Scan() // consume the separator

		if len(s.Bytes()) == 0 {
			err = errors.New("unexpected newline after separator")
			return
		}
		classRecordsRaw = make([]byte, len(s.Bytes()))
		copy(classRecordsRaw, s.Bytes())
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
		certRaw = certBuffer.Bytes()[0:len(certBuffer.Bytes())-1]

		var sigBuffer bytes.Buffer
		for s.Scan() {
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
			return fmt.Errorf("NewServiceProxy: %s",err)
		}

		// Validating is a very expensive operation in the benchmarks
		if cache.verifyRecords {
			err = serviceProxy.Validate()
			if err != nil {
				Error.Printf("could not validate service proxy `%s`. Removing from cache.", err)
				err = cache.removeNoLock(serviceProxy)
				if err != nil {
					Error.Printf("could not remove service proxy (benign on first pass, otherwise it means the service has gone to a bad state): `%s`", err)
				}
				continue
			}
		}

		cache.storeNoLock(serviceProxy)
	}

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
