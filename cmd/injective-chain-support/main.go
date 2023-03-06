// Command router-verifier is main program to verify router swaps.
package main

import (
	"fmt"
	"os"

	"github.com/anyswap/CrossChain-Router/v3/log"
	"github.com/anyswap/CrossChain-Router/v3/params"
	"github.com/anyswap/CrossChain-Router/v3/router/bridge"
	"github.com/anyswap/CrossChain-Router/v3/tokens"
	"github.com/anyswap/RouterSDK-injective/cmd/utils"
	"github.com/anyswap/RouterSDK-injective/config"
	routersdk "github.com/anyswap/RouterSDK-injective/sdk"
	"github.com/anyswap/RouterSDK-injective/server"
	"github.com/urfave/cli/v2"
)

var (
	clientIdentifier = "cosmos chain support"
	// Git SHA1 commit hash of the release (set via linker flags)
	gitCommit = ""
	gitDate   = ""
	// The app that holds all commands and flags.
	app = utils.NewApp(clientIdentifier, gitCommit, gitDate, "cosmos chain support")
)

func initApp() {
	// Initialize the CLI app and start action
	app.Action = run
	app.HideVersion = true // we have a command to print the version
	app.Commands = []*cli.Command{
		utils.VersionCommand,
	}
	app.Flags = []cli.Flag{
		utils.ConfigFileFlag,
		utils.LogFileFlag,
		utils.LogRotationFlag,
		utils.LogMaxAgeFlag,
		utils.VerbosityFlag,
		utils.JSONFormatFlag,
		utils.ColorFormatFlag,
	}
}

func main() {
	initApp()
	if err := app.Run(os.Args); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func run(ctx *cli.Context) error {
	utils.SetLogger(ctx)
	if ctx.NArg() > 0 {
		return fmt.Errorf("invalid command: %q", ctx.Args().Get(0))
	}

	configFile := utils.GetConfigFilePath(ctx)
	config1 := config.LoadConfig(configFile, true)

	initRouterServer := config1.InitRouterServer
	routerConfigFile := config1.RouterConfigFile

	config2 := params.LoadRouterConfig(routerConfigFile, initRouterServer, true)
	tokens.InitRouterSwapType(config2.SwapType)

	routersdk.StartEndpoint()
	server.StartAPIServer()

	bridge.IsWrapperMode = true
	bridge.InitRouterBridges(initRouterServer)
	bridge.StartReloadRouterConfigTask()

	routersdk.InitAfterLoad()

	utils.TopWaitGroup.Wait()
	return nil
}
