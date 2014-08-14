package provider

type Config struct {
	Product   string `mapstructure:"product"`
	VerifySSL bool   `mapstructure:"verify_ssl"`
}
