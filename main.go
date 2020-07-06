package main

import (
	"github.com/ontio/ontology/cmd"
	cmd2 "github.com/qiluge/globalparam/cmd"
	"github.com/urfave/cli"
	"os"
	"runtime"
)

func setupAPP() *cli.App {
	app := cli.NewApp()
	app.Usage = "Ontology GasPrice CLI"
	app.Copyright = "Copyright in 2018 The Ontology Authors"
	app.Commands = []cli.Command{
		cmd2.GenUpdateGlobalParamTxCmd,
		cmd2.GenCreateSnapshotTxCmd,
		cmd2.MultiSignTxCmd,
		cmd2.SendTxCmd,
		cmd2.UpdateGlobalParamByCfgCmd,
		cmd2.CreateSnapshotByCfgCmd,
	}
	app.Flags = []cli.Flag{
	}
	app.Before = func(context *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		return nil
	}
	return app
}

func main() {
	if err := setupAPP().Run(os.Args); err != nil {
		cmd.PrintErrorMsg(err.Error())
		os.Exit(1)
	}
}
