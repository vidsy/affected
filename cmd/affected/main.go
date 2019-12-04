package main

import (
	"github.com/vidsy/affected/pkg/cli"
)

func main() {
	cli.CheckErrExit(cli.RootCmd().Execute())
}
