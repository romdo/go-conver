package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/Masterminds/semver"
	"github.com/davecgh/go-spew/spew"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/octago/sflags/gen/gflag"
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
		GitTagCreate bool

		// version config
		VersionPrefix string `flag:"version-prefix" desc:"Version prefix"`

		// changelog config
		ChangelogGenerate     bool
		ChangelogTemplatePath string
	}
)

func detectBump(c *config) error {
	if !c.AutodetectBump {
		return nil
	}
	repo, err := git.PlainOpen(".")
	if err != nil {
		return err
	}
	tagrefs, err := repo.Tags()
	err = tagrefs.ForEach(func(t *plumbing.Reference) error {
		log.Println(t.Hash(), t.String())
		return nil
	})
	return nil
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

// func changelogGenerate(c *config) error {
// 	if !c.FileUpdate {
// 		return nil
// 	}
// 	currentVersion := "0.0.0"
// 	versionFileBody, err := ioutil.ReadFile(c.FilePath)
// 	if err != nil {
// 		return err
// 	}
// 	currentVersion = string(versionFileBody)
// 	if currentVersion == "" {
// 		currentVersion = c.VersionPrefix + "0.0.0"
// 	}
// 	newVersion, err := bumpVersion(currentVersion, c)
// 	if err != nil {
// 		return err
// 	}
// 	return ioutil.WriteFile(c.FilePath, []byte(newVersion), 0644)
// }

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
	if cleanCurrentVersion.Equal(&cleanNewVersion) {
		cleanNewVersion = cleanNewVersion.IncPatch()
	}
	return c.VersionPrefix + cleanNewVersion.String(), nil
}

func main() {
	c := &config{
		FilePath:      "VERSION",
		VersionPrefix: "v",
	}
	err := gflag.ParseToDef(c)
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	flag.Parse()
	spew.Dump(c)

	actions := []func(*config) error{
		detectBump,
		fileUpdate,
	}
	for _, action := range actions {
		if err := action(c); err != nil {
			log.Fatalf("err: %v", err)
		}
	}
}
