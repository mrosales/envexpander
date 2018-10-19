package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/pkg/errors"

	"github.com/mrosales/envexpander"
)

var (
	withDecryption = false
)

func main() {
	flag.BoolVar(&withDecryption, "with-decryption", false, "Enable KMS decryption for SSM backend")
	flag.Parse()
	args := flag.Args()

	if len(args) <= 0 {
		flag.Usage()
		os.Exit(1)
	}

	path, err := exec.LookPath(args[0])
	if err != nil {
		panic(errors.Wrapf(err, "command '%s' does not exist", args[0]))
	}

	if err := envexpander.Expand(); err != nil {
		fail(err)
	}

	if err := syscall.Exec(path, args[0:], os.Environ()); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "envexpander: %v\n", err)
	os.Exit(1)
}
