# envexpander

`envexpander` is a basic Golang package that can expand environment variables 
from from remote stores. Out of the box, it supports:

* [AWS Simple Systems Manager (SSM) Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-paramstore.html)
* [AWS SecretsManager](https://aws.amazon.com/secrets-manager/)

It also allows you to provide an interface that allows you to implement 
custom parameter-loading logic.

## Command-line Usage

Installation: 
```console
go get -u github.com/mrosales/envexpander/cmd/envexpander
```

Usage:
```console
envexpander [-with-decryption] COMMAND
```

## Using as a package

Instructions and Godocs coming soon
