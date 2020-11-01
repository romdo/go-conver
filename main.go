package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/octago/sflags/gen/gflag"
	convcom "github.com/wfscheper/convcom"
)

//nolint: lll
const changelogTemplate = `
## {{ .Version }} ({{ simpleDate .LatestTagCommit.Committer.When }})
{{ range $type, $commits := .CommitsGrouped }}
### {{ mapConvComType $type }}
{{ range $commit := $commits }}
- {{ if $commit.Conv.Scope }}**{{ $commit.Conv.Scope }}:** {{ else }}{{ end }}{{ $commit.Conv.Description }} ({{ commitHash $commit.Git.Hash.String }})
{{- end }}
{{ end }}
`

var (
	Version string
	Commit  string
	Date    string
)

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

		PrintVersion bool `flag:"version" desc:"print version"`
	}
	changelogEntry struct {
		Version         string
		LatestTagCommit *object.Commit
		Commits         []*changelogCommit
		CommitsGrouped  map[string][]*changelogCommit
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
	if err != nil {
		return err
	}

	err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		rev := plumbing.Revision(tagRef.Name().String())
		tagCommitHash, err := repo.ResolveRevision(rev) //nolint:govet
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
	if err != nil {
		return err
	}

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
		if strings.Contains(commit.Message, "BREAKING") {
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
	if err != nil {
		return err
	}

	err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		rev := plumbing.Revision(tagRef.Name().String())
		tagCommitHash, err := repo.ResolveRevision(rev) //nolint: govet
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
	_, err = repo.CreateTag(newVersion, head.Hash(), &git.CreateTagOptions{
		Message: "chore(version): bump version to " + newVersion,
	})
	if err != nil {
		return err
	}

	return nil
}

func changelogUpdate(c *config) error { //nolint: funlen,gocyclo
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
	if err != nil {
		return err
	}

	err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		rev := plumbing.Revision(tagRef.Name().String())
		tagCommitHash, err := repo.ResolveRevision(rev) //nolint: govet
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

	convComParser, err := convcom.New(&convcom.Config{})
	if err != nil {
		return err
	}

	// find commits since the latest tag
	commitsSinceTag := []*changelogCommit{}
	commitsSinceTagGrouped := map[string][]*changelogCommit{}
	commitIter, err := repo.Log(&git.LogOptions{})
	if err != nil {
		return err
	}

	err = commitIter.ForEach(func(commit *object.Commit) error {
		// once we reach the commit of the latest tag, we're done
		if commit.Hash == latestTagCommit.Hash {
			return errDone
		}
		convCommit, err := convComParser.Parse(commit.Message) //nolint: govet
		if err != nil {
			return err
		}
		changelogEntry := &changelogCommit{
			Git:  commit,
			Conv: convCommit,
		}
		if convCommit.Type == "chore" {
			return nil
		}
		commitType := convCommit.Type
		if convCommit.IsBreaking {
			commitType = "breaking"
		}
		commitsSinceTag = append(commitsSinceTag, changelogEntry)
		if _, ok := commitsSinceTagGrouped[commitType]; !ok {
			commitsSinceTagGrouped[commitType] = []*changelogCommit{}
		}
		commitsSinceTagGrouped[commitType] = append(
			commitsSinceTagGrouped[commitType],
			changelogEntry,
		)

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
	tmpl, err := template.
		New("changelog").
		Funcs(template.FuncMap{
			"simpleDate": func(t time.Time) string {
				return t.Format("2006-01-02")
			},
			"commitHash": func(h string) string {
				return h[:7]
			},
			"mapConvComType": func(t string) string {
				switch t {
				case "breaking":
					return "Breaking Changes"
				case "fix":
					return "Bug Fixes"
				case "feat":
					return "Features"
				case "build":
					return "Build System"
				case "ci":
					return "Continuous Integration"
				case "docs":
					return "Documentation"
				case "style":
					return "Styling"
				case "refactor":
					return "Code Refactoring"
				case "perf":
					return "Performance Improvements"
				case "test":
					return "Tests"
				default:
					return t
				}
			},
		}).
		Parse(changelogTemplate)
	if err != nil {
		return err
	}

	// render template
	var newBody bytes.Buffer
	err = tmpl.Execute(&newBody, &changelogEntry{
		Version:         newVersion,
		LatestTagCommit: latestTagCommit,
		Commits:         commitsSinceTag,
		CommitsGrouped:  commitsSinceTagGrouped,
	})
	if err != nil {
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
	return ioutil.WriteFile(c.ChangelogPath, mergedBody, 0o644) //nolint: gosec
}

func fileUpdate(c *config) error {
	if !c.FileUpdate {
		return nil
	}

	versionFileBody, err := ioutil.ReadFile(c.FilePath)
	if err != nil {
		return err
	}

	currentVersion := string(versionFileBody)
	if currentVersion == "" {
		currentVersion = c.VersionPrefix + "0.0.0"
	}
	newVersion, err := bumpVersion(currentVersion, c)
	if err != nil {
		return err
	}

	//nolint: gosec
	return ioutil.WriteFile(c.FilePath, []byte(newVersion), 0o644)
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

func printVersion() {
	if Date != "" {
		ts, err := strconv.Atoi(Date)
		if err == nil {
			Date = time.Unix(int64(ts), 0).UTC().String()
		}
	}
	fmt.Printf("conver %s (%s, %s)\n", Version, Commit, Date)
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

	if c.PrintVersion {
		printVersion()
		os.Exit(0)
	}

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
