package providers

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// secretsmanager.New(config) conforms to this interface
type SecretsManagerAPI interface {
	GetSecretValue(*secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error)
}

type SecretsManagerProvider struct {
	Client SecretsManagerAPI
}

func (b *SecretsManagerProvider) Get(keys []string) (map[string]string, error) {
	values := map[string]string{} // map env key to ssm value

	for _, v := range keys {
		resp, err := b.Client.GetSecretValue(&secretsmanager.GetSecretValueInput{
			SecretId: aws.String(v),
		})
		if err != nil {
			return values, NewParameterError(v, err)
		}

		values[v] = aws.StringValue(resp.SecretString)
	}

	return values, nil
}
