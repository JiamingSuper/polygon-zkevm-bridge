package main

import (
	"os"

	zkevmbridgeservice "github.com/JiamingSuper/polygon-zkevm-bridge"
	"github.com/urfave/cli/v2"
)

func versionCmd(*cli.Context) error {
	zkevmbridgeservice.PrintVersion(os.Stdout)
	return nil
}
