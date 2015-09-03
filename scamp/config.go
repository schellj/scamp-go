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
var config *Config

var defaultConfigPath = "/etc/SCAMP/soa.conf"
var configLine = regexp.MustCompile(`^\s*([\S^=]+)\s*=\s*([\S]+)`)
var globalConfig *Config

func initConfig() (err error) {
	config = NewConfig()
	err = config.Load()
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

func (conf *Config) Load() (err error) {
	Trace.Printf("reading config")
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