package config

// AzureConfig holds global test configuration translated from environment variables
// for azure provider
type AzureConfig struct {
	AzureClientID string `envconfig:"AZURE_CLIENT_ID"`
	ClientSecret  string `envconfig:"AZURE_CLIENT_SECRET"`
	TenantID      string `envconfig:"TENANT_ID"`
	KeyvaultName  string `envconfig:"KEYVAULT_NAME" default:"csi-secrets-store-e2e"`
	SecretName    string `envconfig:"SECRET_NAME" default:"secret1"`
	SecretVersion string `envconfig:"SECRET_VERSION"`
	SecretValue   string `envconfig:"SECRET_VALUE" default:"test"`
}

func ()