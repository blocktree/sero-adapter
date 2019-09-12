package main

import (
	"fmt"
	"github.com/blocktree/sero-adapter/serocli/commands"
	"gopkg.in/urfave/cli.v1"
	"os"
	"path/filepath"
	"sort"
)

const (
	Identifier   = "openw-sero" // Client identifier to advertise over the network
)

var (
	CommandHelpTemplate = `{{.cmd.Name}}{{if .cmd.Subcommands}} command{{end}}{{if .cmd.Flags}} [command options]{{end}} [arguments...]
{{if .cmd.Description}}{{.cmd.Description}}
{{end}}{{if .cmd.Subcommands}}
SUBCOMMANDS:
	{{range .cmd.Subcommands}}{{.cmd.Name}}{{with .cmd.ShortName}}, {{.cmd}}{{end}}{{ "\t" }}{{.cmd.Usage}}
	{{end}}{{end}}{{if .categorizedFlags}}
{{range $idx, $categorized := .categorizedFlags}}{{$categorized.Name}} OPTIONS:
{{range $categorized.Flags}}{{"\t"}}{{.}}
{{end}}
{{end}}{{end}}`

	// Git SHA1 commit hash of the release (set via linker flags)
	// The app that holds all commands and flags.
	app = NewApp(commands.GitRev, "the Wallet Manager Driver command line interface")
)

func init() {
	cli.AppHelpTemplate = `{{.Name}} {{if .Flags}}[global options] {{end}}command{{if .Flags}} [command options]{{end}} [arguments...]

VERSION:
   {{.Version}}

COMMANDS:
   {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
   {{end}}{{if .Flags}}
GLOBAL OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{end}}
`

	cli.CommandHelpTemplate = CommandHelpTemplate
}

// NewApp creates an app with sane defaults.
func NewApp(gitCommit, usage string) *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Author = ""
	//app.Authors = nil
	app.Email = ""
	app.Version = commands.Version
	if len(gitCommit) >= 0 {
		app.Version += "-" + gitCommit
	}
	app.Usage = usage
	return app
}

func init() {
	// Initialize the CLI app and start openw-sero
	app.Name = Identifier
	app.Action = openwsero
	app.HideVersion = true // we have a command to print the version
	app.Copyright = "Copyright 2018 The OpenWallet Authors"
	app.Version = commands.Version
	app.Commands = commands.Commands
	app.Flags = []cli.Flag{
		commands.AppNameFlag,
		commands.LogDirFlag,
		commands.DebugFlag,
		commands.ConfFlag,
	}

	sort.Sort(cli.CommandsByName(app.Commands))
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

//openwsero
func openwsero(ctx *cli.Context) error {
	return nil
}
