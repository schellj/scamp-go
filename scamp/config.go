package scamp

import "os"
import "bufio"
import "regexp"
import "fmt"
import "strconv"
import "net"

type Config struct {
	// string key for easy equals, byte return for easy nil
	values map[string][]byte
}

// TODO: Will I regret using such a common name as a global variable?
var defaultConfig *Config

var defaultAnnounceInterval = 5
var DefaultConfigPath = "/etc/SCAMP/soa.conf"
var configLine = regexp.MustCompile(`^\s*([\S^=]+)\s*=\s*([\S]+)`)
var globalConfig *Config

var defaultGroupIP = net.IPv4(239, 63, 248, 106)
var defaultGroupPort = 5555

func initConfig(configPath string) (err error) {
	defaultConfig = NewConfig()
	err = defaultConfig.Load(configPath)
	if err != nil {
		err = fmt.Errorf("could not load config: %s", err)
		return
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

func (conf *Config) Load(configPath string) (err error) {
	file,err := os.Open(configPath)
	if err != nil {
		err = fmt.Errorf("no such file %s", DefaultConfigPath)
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

func (conf *Config) DiscoveryMulticastIP() (ip net.IP) {
	rawAddr := conf.values["discovery.multicast_address"]
	if rawAddr != nil {
		return net.IP(rawAddr)
	}

	return defaultGroupIP
}

func (conf *Config) DiscoveryMulticastPort() (port int) {
	port_bytes := conf.values["discovery.port"]
	if port_bytes != nil {
		port64, err := strconv.ParseInt(string(port_bytes), 10, 0)
		if err != nil {
			Error.Printf("could not parse discovery.port `%s`. falling back to default", err)
			port = int(defaultGroupPort)
		} else {
			port = int(port64)
		}

		return
	}

	port = defaultGroupPort
	return
}

func (conf *Config) Get(key string) (value string, ok bool) {
	valueBytes,ok := conf.values[key]
	value = string(valueBytes)
	return
}
