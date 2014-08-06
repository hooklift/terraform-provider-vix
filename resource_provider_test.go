package terraform_vix

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/terraform"
)

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *ResourceProvider

func init() {
	testAccProvider = new(ResourceProvider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"vix": testAccProvider,
	}
}

func TestResourceProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = new(ResourceProvider)
}

func TestResourceProvider_Configure(t *testing.T) {
	rp := new(ResourceProvider)

	raw := map[string]interface{}{
		"product":    "fusion",
		"verify_ssl": "true",
	}

	rawConfig, err := config.NewRawConfig(raw)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = rp.Configure(terraform.NewResourceConfig(rawConfig))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := Config{
		Product:   "fusion",
		VerifySSL: true,
	}

	if !reflect.DeepEqual(rp.Config, expected) {
		t.Fatalf("bad: %#v", rp.Config)
	}
}

func TestResourceProvider_Defaults(t *testing.T) {
	rp := new(ResourceProvider)

	rawConfig, err := config.NewRawConfig(nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = rp.Configure(terraform.NewResourceConfig(rawConfig))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := Config{
		Product:   "workstation",
		VerifySSL: false,
	}

	if !reflect.DeepEqual(rp.Config, expected) {
		t.Fatalf("bad: %#v", rp.Config)
	}
}
