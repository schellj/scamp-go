package scamp

import "encoding/json"
import "encoding/pem"
import "crypto/x509"
import "crypto/rsa"

import "fmt"
import "errors"

type ServiceProxy struct {
	version int
	ident string
	sector string
	weight int
	announceInterval int
	connspec string
	protocols []string
	actions []ServiceProxyClass

	rawClassRecords []byte
	rawCert []byte
	rawSig []byte

	conn *Connection
}

type ServiceProxyClass struct {
	className string
	actions []actionDescription
}

type actionDescription struct {
	actionName string
	crudTags string
	version string
}

func NewServiceProxy(classRecordsRaw []byte, certRaw []byte, sigRaw []byte) (proxy *ServiceProxy, err error) {
	proxy = new(ServiceProxy)
	proxy.rawClassRecords = classRecordsRaw
	proxy.rawCert = certRaw
	proxy.rawSig = sigRaw

	var classRecords []json.RawMessage
	err = json.Unmarshal(classRecordsRaw, &classRecords)
	if err != nil {
		return
	}
	if len(classRecords) != 9 {
		err = errors.New( fmt.Sprintf("expected 9 entries in class record. got %d", len(classRecords)) )
	}

	// OMG, position-based, heterogenously typed values in an array suck to deal with.
	err = json.Unmarshal(classRecords[0], &proxy.version)
	if err != nil {
		return
	}

	err = json.Unmarshal(classRecords[1], &proxy.ident)
	if err != nil {
		return
	}

	err = json.Unmarshal(classRecords[2], &proxy.sector)
	if err != nil {
		return
	}

	err = json.Unmarshal(classRecords[3], &proxy.weight)
	if err != nil {
		return
	}

	err = json.Unmarshal(classRecords[4], &proxy.announceInterval)
	if err != nil {
		return
	}

	err = json.Unmarshal(classRecords[5], &proxy.connspec)
	if err != nil {
		return
	}

	err = json.Unmarshal(classRecords[6], &proxy.protocols)
	if err != nil {
		return
	}

	var rawClasses [][]json.RawMessage
	err = json.Unmarshal(classRecords[7], &rawClasses)
	if err != nil {
		return
	}
	classes := make([]ServiceProxyClass, len(rawClasses), len(rawClasses))
	proxy.actions = classes

	for i,rawClass := range rawClasses {
		if len(rawClass) < 2 {
			err = errors.New( fmt.Sprintf("expected rawClass to have at least 2 entries. was: `%s`", rawClass) )
			return nil, err
		}

		err = json.Unmarshal(rawClass[0], &classes[i].className)
		if err != nil {
			return nil, err
		}

		rawActionsSlice := rawClass[1:]
		classes[i].actions = make([]actionDescription, len(rawActionsSlice), len(rawActionsSlice))

		for j,rawActionSpec := range rawActionsSlice {
			var actionsRawMessages []json.RawMessage
			err = json.Unmarshal(rawActionSpec, &actionsRawMessages)
			if err != nil {
				Trace.Printf("could not parse: %s", rawActionSpec)
				return nil, err
			} else if len(actionsRawMessages) != 2 && len(actionsRawMessages) != 3 {
				err = errors.New( fmt.Sprintf("expected action spec to have 2 or 3 entries. got `%s` (%d)", actionsRawMessages, len(actionsRawMessages) ) )
			}

			err = json.Unmarshal(actionsRawMessages[0], &classes[i].actions[j].actionName)
			if err != nil {
				return nil, err
			}

			err = json.Unmarshal(actionsRawMessages[1], &classes[i].actions[j].crudTags)
			if err != nil {
				return nil, err
			}
		}	
	}

	proxy.conn = nil // we connect on demand
	return
}

func (proxy *ServiceProxy)Validate() (err error) {
	decoded,_ := pem.Decode(proxy.rawCert)
	if decoded == nil {
		err = errors.New( fmt.Sprintf("could not find valid cert in `%s`", proxy.rawCert) )
		return
	}

	// Put pem in form useful for fingerprinting
	cert,err := x509.ParseCertificate(decoded.Bytes)
	if err != nil {
		return err
	}

	pkixInterface := cert.PublicKey
	rsaPubKey, ok := pkixInterface.(*rsa.PublicKey)
	if !ok {
		err = errors.New("could not cast parsed value to rsa.PublicKey")
		return
	}

	valid,err := VerifySHA256(proxy.rawClassRecords, rsaPubKey, proxy.rawSig, false)
	if !valid {
		return
	}

	_ = sha1FingerPrint(cert)

	return nil
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
