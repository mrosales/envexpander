package envexpander

import "github.com/mrosales/envexpander/providers"

type ParameterProvider interface {
	Get(keys []string) (map[string]string, error)
}

// Creates a ParameterProvider using AWS SSM Parameter Store as a backend.
// Supports both KMS-encrypted secrets as well as unencrypted strings.
func NewSSMProvider(client providers.SSMAPI, withDecryption bool) ParameterProvider {
	return &providers.SSMProvider{
		Client:         client,
		WithDecryption: withDecryption,
		BatchSize:      10, // SSM supports batching up to 10 parameter requests at once
	}
}

// Creates a ParameterProvider using AWS SSM Parameter Store as a backend.
// Supports both KMS-encrypted secrets as well as unencrypted strings.
func NewSecretsManagerProvider(client providers.SecretsManagerAPI) ParameterProvider {
	return &providers.SecretsManagerProvider{
		Client: client,
	}
}
