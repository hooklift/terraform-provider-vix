package main

import (
	"github.com/cloudescape/terraform-provider-vix/provider"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(new(provider.ResourceProvider))
}
