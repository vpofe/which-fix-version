package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vpofe/which-fix-version/app"
	"github.com/vpofe/which-fix-version/git"
)

func NewFindCommand() *FindCommand {
	fc := &FindCommand{
		fs: flag.NewFlagSet("find", flag.ContinueOnError),
	}

	fc.fs.StringVar(&fc.commitHash, "commitHash", "", "the main/master/development/custom branch commit hash to find the minimal fix version")
	fc.fs.StringVar(&fc.url, "url", "", "git repository url")
	fc.fs.StringVar(&fc.remoteName, "remoteName", "origin", "remote name to fetch branches from")
	fc.fs.StringVar(&fc.developmentBranchName, "developmentBranchName", "main", "name of the central development branch")
	fc.fs.StringVar(&fc.releaseBranchPrependIdentifiers, "releaseBranchPrependIdentifiers", "release- releases/ release/", "all string characters before the release version")

	return fc
}

type FindCommand struct {
	fs *flag.FlagSet

	url                             string
	remoteName                      string
	developmentBranchName           string
	releaseBranchPrependIdentifiers string
	commitHash                      string
}

func (g *FindCommand) Name() string {
	return g.fs.Name()
}

func (g *FindCommand) Init(args []string) error {
	return g.fs.Parse(args)
}

func (g *FindCommand) Run() error {
	app := app.NewApp(&git.GitConfig{
		CommitHash:                      g.commitHash,
		URL:                             g.url,
		RemoteName:                      g.remoteName,
		DevelopBranchName:               g.developmentBranchName,
		ReleaseBranchPrependIdentifiers: strings.Split(g.releaseBranchPrependIdentifiers, " "),
	})

	if err := tea.NewProgram(app.Model).Start(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}

	return nil
}