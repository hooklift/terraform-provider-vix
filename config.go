package terraform_vix

import (
	"log"
	"strings"

	"github.com/c4milo/govix"
)

type Config struct {
	Product string `mapstructure:"product"`
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

	host, err := vix.Connect(vix.ConnectConfig{
		Provider: p,
	})

	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] VIX Client configured for provider: VMware %s", c.Product)

	return host, nil
}
