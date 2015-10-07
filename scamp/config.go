package scamp

import "os"
import "bufio"
import "regexp"
import "fmt"

type Config struct {
	// string key for easy equals, byte return for easy nil
	values map[string][]byte
}

// TODO: Will I regret using such a common name as a global variable?
var defaultConfig *Config

var defaultConfigPath = "/etc/SCAMP/soa.conf"
var configLine = regexp.MustCompile(`^\s*([\S^=]+)\s*=\s*([\S]+)`)
var globalConfig *Config
var defaultMulticastPort = 9999

func initConfig() (err error) {
	defaultConfig = NewConfig()
	err = defaultConfig.Load()
	if err != nil {
		panic( fmt.Sprintf("could not load config: %s", err) )
	}

	return
}

func NewConfig() (conf *Config) {
	conf = new(Config)
	conf.values = make(map[string][]byte)

	return
}

func DefaultConfig() (conf *Config) {
	return defaultConfig
}

func (conf *Config) Load() (err error) {
	file,err := os.Open(defaultConfigPath)
	if err != nil {
		err = fmt.Errorf("no such file %s", defaultConfigPath)
		return
	}
	scanner := bufio.NewScanner(file)
	conf.doLoad(scanner)

	return
}

func (conf *Config) doLoad(scanner *bufio.Scanner) (err error) {
	var read bool
	for {
		read = scanner.Scan()
		if !read {
			break
		}

		re := configLine.FindSubmatch(scanner.Bytes())
		if re != nil {
			conf.values[string(re[1])] = re[2]
		}
	}

	return
}

func (conf *Config) ServiceKeyPath(serviceName string) (keyPath []byte) {
	return conf.values[serviceName+".soa_key"]
}

func (conf *Config) ServiceCertPath(serviceName string) (certPath []byte) {
	return conf.values[serviceName+".soa_cert"]
}

func (conf *Config) BusPort() (port int) {
	return defaultMulticastPort
}

// Actively probes environment so if no default provided it could error
// TODO: hard-coded to 
func (conf *Config) BusAddress() (address string, err error) {
	defaultBusAddress := conf.values["bus_address"]
	if defaultBusAddress != nil {
		address = string(defaultBusAddress)
		return
	}

  bestAddr,err := MulticastAddrForInterface("lo0")
  if err != nil {
    Error.Printf("could not find best addr: `%s`", err)
    return
  }

  address = bestAddr.String()
	return 
}

func (conf *Config) BusSpec() (spec string, err error) {
	address,err := conf.BusAddress()
	if err != nil {
		return
	}

	spec = fmt.Sprintf("%s:%d", address, conf.BusPort())

	return
}