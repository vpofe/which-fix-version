package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vpofe/which-fix-version/app"
	"github.com/vpofe/which-fix-version/git"
	"github.com/vpofe/which-fix-version/tui"
)

func NewFindLocalCommand() *FindLocalCommand {
	fc := &FindLocalCommand{
		fs: flag.NewFlagSet("local", flag.ContinueOnError),
	}

	fc.fs.StringVar(&fc.commitHash, "commitHash", "", "the main/master/development/custom branch commit hash to find the minimal fix version")
	fc.fs.StringVar(&fc.developmentBranchName, "developmentBranchName", "main", "name of the central development branch")
	fc.fs.StringVar(&fc.releaseBranchPrependIdentifiers, "releaseBranchPrependIdentifiers", "release- releases/ release/", "all string characters before the release version")
	fc.fs.StringVar(&fc.path, "path", "", "the absolute path of the local git repository")

	return fc
}

type FindLocalCommand struct {
	fs *flag.FlagSet

	developmentBranchName           string
	releaseBranchPrependIdentifiers string
	commitHash                      string
	path                            string
}

func (g *FindLocalCommand) Name() string {
	return g.fs.Name()
}

func (g *FindLocalCommand) Init(args []string) error {
	return g.fs.Parse(args)
}

func (g *FindLocalCommand) Run() error {
	app := app.NewApp(&git.GitConfig{
		CommitHash:                      g.commitHash,
		Path:                            g.path,
		DevelopBranchName:               g.developmentBranchName,
		ReleaseBranchPrependIdentifiers: strings.Split(g.releaseBranchPrependIdentifiers, " "),
	}, tui.Local)

	if err := tea.NewProgram(app.Model).Start(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}

	return nil
}