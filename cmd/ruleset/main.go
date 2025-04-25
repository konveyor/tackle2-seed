package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/konveyor/tackle2-seed/pkg"
	"github.com/pborman/getopt/v2"
)

const (
	Resources      = "resources"
	RuleSets       = "rulesets"
	RemoteRuleSets = "default/generated"
)

var Deps = []string{
	"rulesets/00-discovery",
	"rulesets/technology-usage",
}

var YesAssumed = true

func main() {
	cmd := Cmd{}
	err := cmd.Main()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

type Cmd struct {
	Path     string
	Remote   Remote
	Manifest struct {
		Current Manifest
		Remote  Manifest
	}
}

func (r *Cmd) Main() (err error) {
	resources := getopt.StringLong(
		"path",
		'p',
		"./"+Resources,
		"The resources path.")
	remote := getopt.StringLong(
		"remote",
		'r',
		"https://github.com/konveyor/rulesets",
		"The remote (ruleset) github repository URL. May be plan file path.")
	ref := getopt.StringLong(
		"branch",
		'b',
		"",
		"The github branch (any ref).")
	yes := getopt.BoolLong(
		"yes",
		'y',
		"Yes assumed.")
	help := getopt.BoolLong(
		"help",
		'h',
		"Show help.")

	getopt.Parse()
	if *help {
		getopt.Usage()
		return
	}

	YesAssumed = *yes
	r.Path = *resources
	r.Remote.URL = *remote
	r.Remote.Ref = *ref
	r.Manifest.Current.root = *resources

	fmt.Printf("\nResources: %s\n", *resources)
	fmt.Printf("Remote: %s\n", *remote)
	fmt.Printf("Ref: %s\n", *ref)

	err = r.Manifest.Current.Load()
	if err != nil {
		return
	}
	err = r.Remote.Load()
	if err != nil {
		return
	}
	err = r.Reconcile()
	if err != nil {
		return
	}

	fmt.Println("\nDone")
	return
}

func (r *Cmd) Reconcile() (err error) {
	tmpDir, err := ioutil.TempDir("", "ruleset-*")
	if err != nil {
		return
	}
	remote := path.Join(r.Remote.Path, RemoteRuleSets)
	dest := path.Join(tmpDir, RuleSets)
	bash := Bash{Silent: true}
	err = r.CopyTree(bash, 2, remote, dest)
	if err != nil {
		return
	}
	r.Manifest.Remote.root = tmpDir
	err = r.Manifest.Remote.Build()
	if err != nil {
		return
	}
	r.Manifest.Current.Reconcile(r.Manifest.Remote)
	r.Manifest.Current.PrintChanged()
	if !r.Manifest.Current.Dirty() {
		return
	}
	bash = Bash{}
	b := bash.Ask("Apply approved changes?")
	if b {
		err = r.Apply()
		if err != nil {
			return
		}
	}
	return
}

func (r *Cmd) Apply() (err error) {
	manifest := r.Manifest.Current
	for _, ruleSet := range manifest.changed.added {
		err = r.ReplaceDir(ruleSet)
		if err != nil {
			return
		}
	}
	for _, ruleSet := range manifest.changed.updated {
		err = r.ReplaceDir(ruleSet)
		if err != nil {
			return
		}
	}
	for _, ruleSet := range manifest.changed.deleted {
		err = r.Delete(ruleSet)
		if err != nil {
			return
		}
	}
	err = r.Manifest.Current.Write()
	if err != nil {
		return
	}
	bash := Bash{}
	bash.Dir = r.Path
	err = bash.Run("git add", RuleSets)
	if err != nil {
		return
	}
	return
}

func (r *Cmd) ReplaceDir(ruleSet *pkg.RuleSet) (err error) {
	bash := Bash{Silent: true}
	remote := path.Join(r.Manifest.Remote.root, ruleSet.Dir())
	current := path.Join(r.Path, ruleSet.Dir())
	err = bash.Run("rm -rf", current)
	if err != nil {
		return
	}
	bash = Bash{}
	err = r.CopyTree(bash, 1, remote, current)
	return
}

func (r *Cmd) CopyTree(bash Bash, depth int, source, dest string) (err error) {
	if depth == 0 {
		return
	}
	entries, _ := os.ReadDir(source)
	if err != nil {
		return
	}
	err = bash.Run("mkdir -p", dest)
	if err != nil {
		return
	}
	for _, ent := range entries {
		if ent.IsDir() {
			err = r.CopyTree(
				bash,
				depth-1,
				path.Join(source, ent.Name()),
				path.Join(dest, ent.Name()))
			if err != nil {
				return
			}
			continue
		}
		err = bash.Run(
			"cp",
			path.Join(source, ent.Name()),
			dest)
		if err != nil {
			return
		}
	}
	return
}

func (r *Cmd) Delete(ruleSet *pkg.RuleSet) (err error) {
	bash := Bash{}
	p := path.Join(r.Path, ruleSet.Dir())
	err = bash.Run("rm -rf", p)
	return
}
