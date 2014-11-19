// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"github.com/hashicorp/terraform/plugin"
	vix "github.com/hooklift/terraform-provider-vix/provider"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: vix.Provider,
	})
}
