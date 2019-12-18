package main

import (
	"github.com/vidsy/affected/pkg/cmd"
)

func main() {
	cmd.CheckErrExit(cmd.RootCmd().Execute())
}
