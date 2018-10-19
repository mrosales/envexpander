package envexpander

//go:generate mockery -name=ParameterProvider
//go:generate mockery -dir=./providers -all

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/pkg/errors"
)

var (
	defaultAWSConfig = makeDefaultAWSConfig()
	defaultEnv       = &systemEnvironment{}
	defaultExpander  = NewCustomExpander(map[string]ParameterProvider{
		"ssm":            NewSSMProvider(ssm.New(defaultAWSConfig), true),
		"secretsManager": NewSecretsManagerProvider(secretsmanager.New(defaultAWSConfig)),
	}, defaultEnv)
)

type Expander interface {
	Expand() error
}

type Environment interface {
	SetEnv(k, v string) error
	Environ() []string
}

// Creates a new expander object that replaces environment variables from the provided
// backends. If you have no need of overwriting the schemeMap, use NewExpander instead
// if env is not provided, the system environment will be used instead
func NewCustomExpander(schemeMap map[string]ParameterProvider, env Environment) Expander {
	return &expander{
		map[string][]expanderVar{},
		schemeMap,
		time.Time{},
		map[string]string{},
	}
}

type expander struct {
	varsByScheme      map[string][]expanderVar
	schemeProviderMap map[string]ParameterProvider
	lastUpdate        time.Time
	currentValues     map[string]string
}

type expanderVar struct {
	Scheme    string
	EnvName   string
	RemoteKey string
}

func Expand() error {
	return defaultExpander.Expand()
}

func (e *expander) Expand() (err error) {
	if e.varsByScheme, err = loadEnv(e.schemeProviderMap); err != nil {
		return err
	}

	changeSet, err := e.loadRemoteVariables()
	if err != nil {
		return err
	}

	for k, v := range changeSet {
		if err := os.Setenv(k, v); err != nil {
			return err
		}
	}

	return nil
}

func (e *expander) loadRemoteVariables() (map[string]string, error) {
	changeSet := map[string]string{}
	for scheme, vars := range e.varsByScheme {
		var err error

		keys := make([]string, len(vars))
		reverseLookup := map[string]string{}
		for i, v := range vars {
			keys[i] = v.RemoteKey
			reverseLookup[v.RemoteKey] = v.EnvName
		}

		provider := e.schemeProviderMap[scheme]
		values, err := provider.Get(keys)
		if err != nil {
			return changeSet, errors.Wrapf(err, "fetch %s variables", scheme)
		}

		for remoteKey, v := range values {
			envKey := reverseLookup[remoteKey]
			if e.currentValues[envKey] == v {
				continue
			}
			changeSet[envKey] = v
		}
	}
	return changeSet, nil
}

func loadEnv(schemeMap map[string]ParameterProvider) (map[string][]expanderVar, error) {
	// parse environment
	varsByScheme := map[string][]expanderVar{}
	for _, s := range os.Environ() {
		parts := strings.SplitN(s, "=", 2)
		envKey, envVal := parts[0], parts[1]
		u, err := url.Parse(envVal)
		if err != nil {
			continue
		}

		if schemeMap[u.Scheme] != nil {
			ev := expanderVar{
				EnvName:   envKey,
				RemoteKey: strings.TrimPrefix(envVal, u.Scheme+"://"),
			}

			varsByScheme[u.Scheme] = append(varsByScheme[u.Scheme], ev)
		}
	}
	return varsByScheme, nil
}

type systemEnvironment struct{}

func (e *systemEnvironment) SetEnv(k, v string) error {
	return os.Setenv(k, v)
}

func (e *systemEnvironment) Environ() []string {
	return os.Environ()
}

func makeDefaultAWSConfig() client.ConfigProvider {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Credentials: credentials.NewChainCredentials([]credentials.Provider{
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{},
				&ec2rolecreds.EC2RoleProvider{
					Client: ec2metadata.New(session.Must(session.NewSession(&aws.Config{}))),
				},
			}),
		},
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		fmt.Fprint(os.Stderr, "failed to load aws configuration")
		panic(err)
	}
	return sess
}
