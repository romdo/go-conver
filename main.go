package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"strings"

	"github.com/Masterminds/semver"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/octago/sflags/gen/gflag"
	convcom "github.com/wfscheper/convcom"
)

const changelogTemplate = `
## {{ .Version }}

{{- range $commit := .Commits }}

### {{ $commit.Git.Hash }}

{{ $commit.Git.Message }}
* Type: {{ $commit.Conv.Type }}
* Scope: {{ $commit.Conv.Scope }}
* Description: {{ $commit.Conv.Description }}
* Header: {{ $commit.Conv.Header }}
* MergeHeader: {{ $commit.Conv.MergeHeader }}
* Body: {{ $commit.Conv.Body }}
* Footers: {{ $commit.Conv.Footers }}
* Mentions: {{ $commit.Conv.Mentions }}
* References: {{ $commit.Conv.References }}
* Notes: {{ $commit.Conv.Notes }}
* Reverts: {{ $commit.Conv.Reverts }}
* IsBreaking: {{ $commit.Conv.IsBreaking }}
{{ end }}
`

type (
	config struct {
		// action
		AutodetectBump bool `flag:"bump-auto" desc:"Bump version based on semantic commits"`
		BumpMajor      bool `flag:"bump-major" desc:"Bump major"`
		BumpMinor      bool `flag:"bump-minor" desc:"Bump minor"`
		BumpPatch      bool `flag:"bump-patch" desc:"Bump patch"`

		// version file config
		FileUpdate bool   `flag:"file-update" desc:"Use version file"`
		FilePath   string `flag:"file-path" desc:"Version file path"`

		// git tag config
		GitTagUpdate bool `flag:"git-tag-update" desc:"Use git tags"`

		// version config
		VersionPrefix string `flag:"version-prefix" desc:"Version prefix"`

		// changelog config
		ChangelogUpdate bool   `flag:"changelog-update" desc:"Update changelog"`
		ChangelogPath   string `flag:"changelog-path" desc:"Changelog file path"`
	}
	changelogEntry struct {
		Version string
		Commits []*changelogCommit
	}
	changelogCommit struct {
		Git  *object.Commit
		Conv *convcom.Commit
	}
)

// errDone is used to exit the git iterators early
var errDone = errors.New("done")

func autodetectBump(c *config) error {
	// check if we should be autodetecting the bump
	if !c.AutodetectBump {
		return nil
	}

	// open repository
	repo, err := git.PlainOpen(".")
	if err != nil {
		return err
	}

	// get head
	head, err := repo.Head()
	if err != nil {
		return err
	}

	// find the latest tag
	var latestTagCommit *object.Commit
	tagRefs, err := repo.Tags()
	err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		rev := plumbing.Revision(tagRef.Name().String())
		tagCommitHash, err := repo.ResolveRevision(rev)
		if err != nil {
			return err
		}
		commit, err := repo.CommitObject(*tagCommitHash)
		if err != nil {
			return err
		}
		if latestTagCommit == nil {
			latestTagCommit = commit
		}
		if commit.Committer.When.After(latestTagCommit.Committer.When) {
			latestTagCommit = commit
		}
		return nil
	})
	if err != nil && err != errDone {
		return err
	}

	// find commits since the latest tag
	commitsSinceTag := []*object.Commit{}
	commitIter, err := repo.Log(&git.LogOptions{})
	err = commitIter.ForEach(func(commit *object.Commit) error {
		// once we reach the commit of the latest tag, we're done
		if commit.Hash == latestTagCommit.Hash {
			return errDone
		}
		commitsSinceTag = append(commitsSinceTag, commit)
		return nil
	})
	if err != nil && err != errDone {
		return err
	}

	// check current head is the latest tag
	if latestTagCommit.Hash == head.Hash() {
		return errors.New("head is already tagged")
	}

	// go through the commits and figure out what we need to bump
	c.BumpMajor = false
	c.BumpMinor = false
	c.BumpPatch = false
	for _, commit := range commitsSinceTag {
		switch {
		case strings.Contains(commit.Message, "BREAKING"):
			c.BumpMajor = true
		}
	}

	return nil
}

func gitTagUpdate(c *config) error {
	// check if we are updating git tags
	if !c.GitTagUpdate {
		return nil
	}

	// open repository
	repo, err := git.PlainOpen(".")
	if err != nil {
		return err
	}

	// get head
	head, err := repo.Head()
	if err != nil {
		return err
	}

	// find the latest tag
	var latestTagCommit *object.Commit
	var latestTagName string
	tagRefs, err := repo.Tags()
	err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		rev := plumbing.Revision(tagRef.Name().String())
		tagCommitHash, err := repo.ResolveRevision(rev)
		if err != nil {
			return err
		}
		commit, err := repo.CommitObject(*tagCommitHash)
		if err != nil {
			return err
		}
		if latestTagCommit == nil {
			latestTagCommit = commit
			latestTagName = tagRef.Name().Short()
		}
		if commit.Committer.When.After(latestTagCommit.Committer.When) {
			latestTagCommit = commit
			latestTagName = tagRef.Name().Short()
		}
		return nil
	})
	if err != nil && err != errDone {
		return err
	}

	// check current head is the latest tag
	if latestTagCommit.Hash == head.Hash() {
		return errors.New("head is already tagged")
	}

	// bump version
	newVersion, err := bumpVersion(latestTagName, c)
	if err != nil {
		return err
	}

	// create new tag
	repo.CreateTag(newVersion, head.Hash(), &git.CreateTagOptions{
		Message: "chore(version): bump version to " + newVersion,
	})
	return nil
}

func changelogUpdate(c *config) error {
	// check if we are updating changelog
	if !c.ChangelogUpdate {
		return nil
	}

	// open repository
	repo, err := git.PlainOpen(".")
	if err != nil {
		return err
	}

	// get head
	head, err := repo.Head()
	if err != nil {
		return err
	}

	// find the latest tag
	var latestTagCommit *object.Commit
	var latestTagName string
	tagRefs, err := repo.Tags()
	err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		rev := plumbing.Revision(tagRef.Name().String())
		tagCommitHash, err := repo.ResolveRevision(rev)
		if err != nil {
			return err
		}
		commit, err := repo.CommitObject(*tagCommitHash)
		if err != nil {
			return err
		}
		if latestTagCommit == nil {
			latestTagCommit = commit
			latestTagName = tagRef.Name().Short()
		}
		if commit.Committer.When.After(latestTagCommit.Committer.When) {
			latestTagCommit = commit
			latestTagName = tagRef.Name().Short()
		}
		return nil
	})
	if err != nil && err != errDone {
		return err
	}

	// check current head is the latest tag
	if latestTagCommit.Hash == head.Hash() {
		return errors.New("head is already tagged")
	}

	// bump version
	newVersion, err := bumpVersion(latestTagName, c)
	if err != nil {
		return err
	}

	fmt.Println(latestTagName, newVersion)

	convComParser, err := convcom.New(&convcom.Config{})
	if err != nil {
		return err
	}

	// find commits since the latest tag
	commitsSinceTag := []*changelogCommit{}
	commitIter, err := repo.Log(&git.LogOptions{})
	err = commitIter.ForEach(func(commit *object.Commit) error {
		// once we reach the commit of the latest tag, we're done
		if commit.Hash == latestTagCommit.Hash {
			return errDone
		}
		convCommit, err := convComParser.Parse(commit.Message)
		if err != nil {
			return err
		}
		commitsSinceTag = append(commitsSinceTag, &changelogCommit{
			Git:  commit,
			Conv: convCommit,
		})
		return nil
	})
	if err != nil && err != errDone {
		return err
	}

	// check current head is the latest tag
	if latestTagCommit.Hash == head.Hash() {
		return errors.New("head is already tagged")
	}

	// go through the commits and figure out what we need to bump
	tmpl, err := template.New("changelog").Parse(changelogTemplate)
	if err != nil {
		return err
	}

	// render template
	var newBody bytes.Buffer
	if err := tmpl.Execute(&newBody, &changelogEntry{
		Version: newVersion,
		Commits: commitsSinceTag,
	}); err != nil {
		return err
	}

	// get existing file contents
	changelogBody, err := ioutil.ReadFile(c.ChangelogPath)
	if err != nil {
		return err
	}

	// merge existing changelog with new
	mergedBody := append(newBody.Bytes(), changelogBody...)

	// and update file
	return ioutil.WriteFile(c.ChangelogPath, mergedBody, 0644)
}

func fileUpdate(c *config) error {
	if !c.FileUpdate {
		return nil
	}
	currentVersion := "0.0.0"
	versionFileBody, err := ioutil.ReadFile(c.FilePath)
	if err != nil {
		return err
	}
	currentVersion = string(versionFileBody)
	if currentVersion == "" {
		currentVersion = c.VersionPrefix + "0.0.0"
	}
	newVersion, err := bumpVersion(currentVersion, c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(c.FilePath, []byte(newVersion), 0644)
}

func bumpAtLeastMinor(c *config) error {
	if !c.BumpMajor && !c.BumpMinor && !c.BumpPatch {
		c.BumpPatch = true
	}
	return nil
}

func bumpVersion(currentVersion string, c *config) (string, error) {
	cleanCurrentVersionString := currentVersion[len(c.VersionPrefix):]
	cleanCurrentVersion, err := semver.NewVersion(cleanCurrentVersionString)
	if err != nil {
		return "", err
	}
	cleanNewVersion := *cleanCurrentVersion
	if c.BumpMajor {
		cleanNewVersion = cleanNewVersion.IncMajor()
	}
	if c.BumpMinor {
		cleanNewVersion = cleanNewVersion.IncMinor()
	}
	if c.BumpPatch {
		cleanNewVersion = cleanNewVersion.IncPatch()
	}
	return c.VersionPrefix + cleanNewVersion.String(), nil
}

func main() {
	c := &config{
		FilePath:      "VERSION",
		VersionPrefix: "v",
		ChangelogPath: "CHANGELOG",
	}
	err := gflag.ParseToDef(c)
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	flag.Parse()

	actions := []func(*config) error{
		autodetectBump,
		bumpAtLeastMinor,
		changelogUpdate,
		fileUpdate,
		gitTagUpdate,
	}
	for _, action := range actions {
		if err := action(c); err != nil {
			log.Fatalf("err: %v", err)
		}
	}
}
