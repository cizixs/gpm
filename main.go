/*
gpm stands for go Projects manager.

It provides the basic functionality to visually manage go Projects
under `$GOPATH/src`. You can list, find, remove, and goto Projects easily
without tediously `cd` to the target path and
type long commands which can be error prune.

	gpm source list
	gpm owner list --source=github.com
	gpm repo list --source=github.com --owner=cizixs

	gpm goto repo

	gpm --help

*/

package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type sourceType struct {
	source string
}

func newSource(source string) *sourceType {
	return &sourceType{source: source}
}

type sourcesList struct {
	sources   map[sourceType]bool
	maxLength int
}

func newSourcesList() *sourcesList {
	s := &sourcesList{}
	s.sources = make(map[sourceType]bool)
	s.maxLength = 0
	return s
}

func (s *sourcesList) addSource(source *sourceType) {
	if !s.sources[*source] {
		s.sources[*source] = true
	}
	if len(source.source) > s.maxLength {
		s.maxLength = len(source.source)
	}
}

type ownerType struct {
	owner  string
	source string
}

func newOwner(owner, source string) *ownerType {
	return &ownerType{
		owner:  owner,
		source: source,
	}
}

type ownersList struct {
	owners    map[ownerType]bool
	maxLength int
}

func newOwnersList() *ownersList {
	o := &ownersList{}
	o.owners = make(map[ownerType]bool)
	o.maxLength = 0
	return o
}

func (o *ownersList) addOwner(owner *ownerType) {
	if !o.owners[*owner] {
		o.owners[*owner] = true
	}

	if len(owner.owner) > o.maxLength {
		o.maxLength = len(owner.owner)
	}
}

// goRepo structure represents a go package
type goRepo struct {
	source string
	owner  string
	repo   string
}

// Projects stores structed data for all go packages
type Projects struct {
	sources *sourcesList
	owners  *ownersList
	repos   []goRepo // all repos found in GOPATH
}

// NewProjects returns a new Projects structure
func NewProjects() *Projects {
	p := &Projects{}
	p.sources = newSourcesList()
	p.owners = newOwnersList()
	p.repos = make([]goRepo, 0, 10)
	return p
}

// AddRepo adds a repo to projects
func (p *Projects) AddRepo(repo goRepo) {
	p.repos = append(p.repos, repo)

	p.sources.addSource(newSource(repo.source))
	p.owners.addOwner(newOwner(repo.owner, repo.source))
}

// Sources return all sources in a slice
func (p *Projects) Sources() []string {
	sources := make([]string, 0, len(p.sources.sources))
	for source := range p.sources.sources {
		sources = append(sources, source.source)
	}
	return sources
}

// Owners return all owners in a slice
func (p *Projects) Owners() []string {
	owners := make([]string, 0, len(p.owners.owners))
	for owner := range p.owners.owners {
		owners = append(owners, owner.owner)
	}
	return owners
}

var allProjects = NewProjects()

// isDir checks whether a given path is a valid directory
func isDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return fileInfo.IsDir()
}

func isGitRoot(dir string) bool {
	return isDir(path.Join(dir, ".git"))
}

func parsegoRepo(path string) {
	// path comes with `GOPATH/src` prefix, we only need the last items.
	parts := strings.Split(path, string(filepath.Separator))
	n := len(parts)

	// FIXME(cizixs): there might be cases repo does not have exactly 3 parts
	allProjects.AddRepo(
		goRepo{
			source: parts[n-1],
			owner:  parts[n-2],
			repo:   parts[n-3],
		},
	)
}

func isGogoRepo(path string) bool {
	// TODO(cizixs): right now this only checks that this is a git repo,
	// non-git repo should be supported too.
	return isDir(path) && isGitRoot(path)
}

func visit(path string, f os.FileInfo, err error) error {
	// If find a valid project path, process it,
	// and skip walking the subdirectories or files in the path
	if isGogoRepo(path) {
		parsegoRepo(path)
		return filepath.SkipDir
	}
	return nil
}

func printResult(allProjects *Projects) {
	for i, project := range allProjects.repos {
		fmt.Printf("%d:  %s\t%s\t%s\n", i, project.repo, project.owner, project.source)
	}
}

func main() {
	// TODO(cizixs): GOPATH might contains multiple directories
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return
	}
	filepath.Walk(path.Join(gopath, "src"), visit)

	printResult(allProjects)
}
