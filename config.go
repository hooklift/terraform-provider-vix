package terraform_vix

type Config struct {
	Product   string `mapstructure:"product"`
	VerifySSL bool   `mapstructure:"verify_ssl"`
}
