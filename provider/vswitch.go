package provider

type VSwitch struct {
	// Name for this switch
	Name string
	// Whether to allow machines in this switch to access outside of the nework
	// using Network Address Translations
	NAT bool
	// Whether to enable DHCP on this switch
	DHCP bool
	// CIDR block for the DHCP server
	Range string
	// Whether to attach the host machine to this switch
	HostAccess bool
}

func (v *VSwitch) Create()  {}
func (v *VSwitch) Destroy() {}
func (v *VSwitch) Refresh() {}
func (v *VSwitch) Update()  {}
