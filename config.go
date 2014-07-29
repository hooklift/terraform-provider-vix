package terraform_vix

import (
	"log"
	"strings"

	"github.com/c4milo/govix"
)

type Config struct {
	Product   string `mapstructure:"product"`
	VerifySSL bool   `mapstructure:"verify_ssl"`
}

// Client() returns a new connection for accesing VMware Fusion, Workstation or
// Server
func (c *Config) Client() (*vix.Host, error) {
	var p vix.Provider
	switch strings.ToLower(c.Product) {
	case "fusion", "workstation":
		p = vix.VMWARE_WORKSTATION
	case "serverv1":
		p = vix.VMWARE_SERVER
	case "serverv2":
		p = vix.VMWARE_VI_SERVER
	case "player":
		p = vix.VMWARE_PLAYER
	case "workstation_shared":
		p = vix.VMWARE_WORKSTATION_SHARED
	default:
		p = vix.VMWARE_WORKSTATION
	}

	var options vix.HostOption
	if c.VerifySSL {
		options = vix.VERIFY_SSL_CERT
	}

	host, err := vix.Connect(vix.ConnectConfig{
		Provider: p,
		Options:  options,
	})

	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] VIX Client configured for provider: VMware %s. SSL: %t", c.Product, c.VerifySSL)

	return host, nil
}
