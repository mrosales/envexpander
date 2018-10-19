package providers

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
)

const (
	maxSSMBatchSize = 10
)

// An minimal interface that defines support for the SSM API.
// This is used over ssmiface.SSMAPI in order to facilitate testing.
// N.B.: ssm.New(config) conforms to this interface
type SSMAPI interface {
	GetParameters(input *ssm.GetParametersInput) (*ssm.GetParametersOutput, error)
}

type SSMProvider struct {
	Client         SSMAPI
	WithDecryption bool
	BatchSize      int
}

func (b *SSMProvider) Get(keys []string) (map[string]string, error) {
	if b.BatchSize > maxSSMBatchSize {
		b.BatchSize = maxSSMBatchSize
	}

	values := map[string]string{} // map env key to ssm value

	names := aws.StringSlice(keys)

	for i := 0; i < len(names); i += b.BatchSize {
		j := i + b.BatchSize
		if j > len(names) {
			j = len(names)
		}
		resp, err := b.Client.GetParameters(&ssm.GetParametersInput{
			Names:          names[i:j],
			WithDecryption: aws.Bool(b.WithDecryption),
		})
		if err != nil {
			return values, err
		} else if len(resp.InvalidParameters) > 0 {
			errs := make([]error, len(resp.InvalidParameters))
			for i, name := range resp.InvalidParameters {
				if name == nil {
					errs[i] = NewParameterError(
						"<nil>",
						fmt.Errorf("nil parameter name is invalid"))
					continue
				}
				errs[i] = NewParameterError(aws.StringValue(name), fmt.Errorf("parameter is invalid"))
			}
			return values, &ProviderError{errs}
		}

		for _, param := range resp.Parameters {
			values[*param.Name] = *param.Value
		}
	}
	return values, nil
}
