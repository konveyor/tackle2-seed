package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/konveyor/tackle2-seed/pkg"
	"github.com/pborman/getopt/v2"
)

const (
	Resources      = "resources"
	RuleSets       = "rulesets"
	RemoteRuleSets = "stable"
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

	remote := filepath.Join(r.Remote.Path, RemoteRuleSets)
	dest := filepath.Join(tmpDir, RuleSets)
	paths, err := r.copyRuleSets(remote, dest)
	if err != nil {
		return
	}
	r.Manifest.Remote.root = tmpDir
	err = r.Manifest.Remote.Build(paths)
	if err != nil {
		return
	}
	r.Manifest.Current.Reconcile(r.Manifest.Remote)
	r.Manifest.Current.PrintChanged()
	if !r.Manifest.Current.Dirty() {
		return
	}
	bash := Bash{}
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
	remote := filepath.Join(r.Manifest.Remote.root, ruleSet.Dir())
	current := filepath.Join(r.Path, ruleSet.Dir())
	err = bash.Run("rm -rf", current)
	if err != nil {
		return
	}
	bash = Bash{}
	err = bash.Run("mkdir -p", current)
	if err != nil {
		return
	}
	err = bash.Run("cp -r", remote, filepath.Dir(current))
	if err != nil {
		return
	}
	return
}

func (r *Cmd) Delete(ruleSet *pkg.RuleSet) (err error) {
	bash := Bash{}
	p := filepath.Join(r.Path, ruleSet.Dir())
	err = bash.Run("rm -rf", p)
	return
}

func (r *Cmd) copyRuleSets(root, dest string) (paths []string, err error) {
	pathMap := make(map[string][]string)
	var fn func(string)
	fn = func(dir string) {
		if strings.HasSuffix(dir, ".") {
			return
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		var files []string
		for _, ent := range entries {
			if ent.IsDir() {
				next := filepath.Join(dir, ent.Name())
				fn(next)
				if err != nil {
					return
				}
				continue
			}
			file := filepath.Join(dir, ent.Name())
			files = append(files, file)
			if ent.Name() == pkg.RuleSetYaml {
				ruleSetDir := strings.TrimPrefix(dir, root)
				ruleSetDir = filepath.Join(RuleSets, ruleSetDir)
				pathMap[ruleSetDir] = files
			}
		}
	}
	fn(root)
	for ruleSet, files := range pathMap {
		paths = append(paths, ruleSet)
		for _, p := range files {
			var in, out *os.File
			in, err = os.Open(p)
			if err != nil {
				return
			}
			defer func() {
				_ = in.Close()
			}()
			p = strings.TrimPrefix(p, root)
			p = filepath.Join(dest, p)
			err = os.MkdirAll(filepath.Dir(p), 0755)
			if err != nil {
				return
			}
			out, err = os.Create(p)
			if err != nil {
				return
			}
			defer func() {
				_ = out.Close()
			}()
			_, err = io.Copy(out, in)
			if err != nil {
				return
			}
		}
	}
	return
}
