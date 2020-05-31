package main

import (
	"context"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	promdiscovery "github.com/sgoroshko/promdiscovery/cmd"
)

// VSN represent app version, set by LDFLAGS
var VSN = ""

func main() {
	app := cli.NewApp()
	app.Usage = ""
	app.Version = VSN
	app.Commands = []*cli.Command{
		promdiscovery.NewCommandCompose(),
	}

	err := app.Run(os.Args)
	if err != nil {
		if errors.Cause(err) == context.Canceled {
			return // cancelOnSignal(...
		}

		logrus.Errorf("%+v", err)
	}
}
