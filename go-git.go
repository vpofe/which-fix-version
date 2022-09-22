package main

import (
	"log"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/memory"
)

func (m model) findFixVersion() tea.Msg {

	rootCandidates, releases := getRemoteBranches()

	// fetch commit list from ma(in/ster)
	root := selectRoot(rootCandidates)
	// check latest release

	sortedReleases := getSortedReleases(releases)

	c := getRootCommit(m.commitHash, root)

	var message string

	if c == nil {
		message = "No such hash in the root of this repo"
		return fixVersionMsg(message)
	} else {
		message = "No fixed version found"

		fixedVersions := make([]string, 0)

		for _, version := range sortedReleases {
			if isCommitPresentOnBranch(c, releases[version]) {
				fixedVersions = append(fixedVersions, version)
			}

			// FIXME: cancel looking further if previous doesn't have a fixed version any longer
		}

		if len(fixedVersions) > 0 {
			return fixVersionMsg(fixedVersions[len(fixedVersions)-1])
		} else {
			return fixVersionMsg("No fixed version found")
		}
	}
}

func isCommitPresentOnBranch(rootCommit *object.Commit, branch string) bool {
	result := false

	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:           "https://github.com/vpofe/just-in-time",
		ReferenceName: plumbing.ReferenceName(branch),
		RemoteName:    "origin",
	})

	CheckIfError(err)

	// Gets the HEAD history from HEAD, just like this command:

	// ... retrieves the branch pointed by HEAD
	ref, err := r.Head()
	CheckIfError(err)

	// ... retrieves the commit history
	since := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2099, 7, 30, 0, 0, 0, 0, time.UTC)
	cIter, err := r.Log(&git.LogOptions{From: ref.Hash(), Since: &since, Until: &until})
	CheckIfError(err)

	// ... just iterates over the commits, printing it
	err = cIter.ForEach(func(c *object.Commit) error {
		// FIXME: get to the bottom of isAncestor logic
		isAncestor, parseErr := rootCommit.IsAncestor(c)

		CheckIfError(parseErr)
		if isAncestor {
			result = true
			return nil
		}

		return nil
	})

	CheckIfError(err)

	return result
}

func getSortedReleases(releases map[string]string) []string {
	versions := make([]string, 0)
	for k := range releases {
		versions = append(versions, k)
	}

	// FIXME: do real semver number sort instead of string alphabetical sort
	sort.Strings(versions)

	for i := len(versions)/2 - 1; i >= 0; i-- {
		opp := len(versions) - 1 - i
		versions[i], versions[opp] = versions[opp], versions[i]
	}

	return versions
}

func selectRoot(rootCandidates []string) string {
	// TODO: this should come as default from a flag, lets have, main, master, development fallback
	return "main" // rootCandidates[0]
}

func getRootCommit(hash string, rootBranch string) *object.Commit {
	// Clones the given repository, creating the remote, the local branches
	// and fetching the objects, everything in memory:
	// FIXME: repo should be stored centrally
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: "https://github.com/vpofe/just-in-time",
	})

	CheckIfError(err)

	// Gets the HEAD history from HEAD, just like this command:
	// ... retrieves the branch pointed by HEAD
	ref, err := r.Head()
	CheckIfError(err)

	// ... retrieves the commit history
	since := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2099, 7, 30, 0, 0, 0, 0, time.UTC)
	cIter, err := r.Log(&git.LogOptions{From: ref.Hash(), Since: &since, Until: &until})
	CheckIfError(err)

	var commit *object.Commit
	// ... just iterates over the commits, printing it
	err = cIter.ForEach(func(c *object.Commit) error {
		if c.Hash.String() == hash {
			commit = c
			return nil
		}
		return nil
	})

	CheckIfError(err)

	return commit
}

func getRemoteBranches() ([]string, map[string]string) {
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{"https://github.com/vpofe/just-in-time"},
	})

	refs, err := remote.List(&git.ListOptions{})

	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	releases := make(map[string]string)
	rootCandidates := make([]string, 0)

	for _, ref := range refs {
		s := ref.String()
		if strings.Contains(s, "refs/heads/") {
			branchName := strings.SplitAfter(s, " ")[1]

			var branchVersion string

			if strings.Contains(branchName, "release/") {
				branchVersion = strings.SplitAfter(branchName, "release/")[1]
				releases[branchVersion] = branchName
			} else if strings.Contains(branchName, "releases/") {
				branchVersion = strings.SplitAfter(branchName, "releases/")[1]
				releases[branchVersion] = branchName
			} else if strings.Contains(branchName, "release-") {
				branchVersion = strings.SplitAfter(branchName, "release-")[1]
				releases[branchVersion] = branchName
			} else if branchName == "main" || branchName == "master" || branchName == "development" {
				// FIXME: hardcoded main
				rootCandidates = append(rootCandidates, "main")
			}
		}
	}

	return rootCandidates, releases
}
