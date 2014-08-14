package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/c4milo/terraform-provider-vix/provider"
)

func main() {
	plugin.Serve(new(provider.ResourceProvider))
}
