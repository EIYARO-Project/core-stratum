package hostprovider

import (
	"errors"

	staticprovider "corepool/common/rpc/hostprovider/static"
	"corepool/common/rpc/hostprovider/types"
)

var providers = map[string]types.Provider{}

func InitStaticProvider(config map[string][]string) error {
	for s, addrs := range config {
		providers[s] = staticprovider.NewProvider(s, addrs)
	}
	return nil
}

// Query uses the mysqlClient to query available services
func Get(service string) ([]string, error) {
	p := providers[service]
	if p == nil {
		return nil, errors.New("service not defined")
	}
	return p.Get()
}
