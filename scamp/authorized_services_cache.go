package scamp

// import "fmt"
import "errors"
import "bufio"
import "bytes"

type AuthorizedServiceSpec struct {
	Fingerprint []byte
	Actions []ServiceProxyClass
}

type AuthorizedServicesCache struct {
	services []AuthorizedServiceSpec
}

func NewAuthorizedServicesCache() (cache *AuthorizedServicesCache) {
	cache = new(AuthorizedServicesCache)
	cache.services = make([]AuthorizedServiceSpec, 100)

	return
}

func (cache *AuthorizedServicesCache)LoadAuthorizedServices(s *bufio.Scanner) (err error) {
	var read bool
	count := 1

	for {
		read = s.Scan()

		if !read {
			break
		// Skip empty lines
		} else if len(s.Bytes()) == 0 {
			continue
		}

		_,_ = NewAuthorizedServicesSpec(s.Bytes())

		count = count + 1
	}

	return
}

func NewAuthorizedServicesSpec(line []byte) (spec *AuthorizedServiceSpec, err error) {
	s := bufio.NewScanner(bytes.NewReader(line))
	s.Split(bufio.ScanWords)

	s.Scan()
	if len(s.Bytes()) == 0 || s.Bytes()[0] == '#' {
		err = errors.New("invalid")
		return
	}

	spec = new(AuthorizedServiceSpec)
	spec.Actions = make([]ServiceProxyClass, 100)

	spec.Fingerprint = make([]byte, len(s.Bytes()))
	copy(spec.Fingerprint, s.Bytes())

	var read bool
	for {
		read = s.Scan()
		if !read {
			break
		}
	}

	return
}