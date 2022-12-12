package kegeye

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
)

var (
	kegSearchPaths = []string{"", "docs"}

  ErrUnsupportedSourceType = errors.New("unsupported dex source type")
	ErrNoKegFound  = errors.New("no keg found in repo")
)

type DexSourceType string

const (
	DexSourceGithub DexSourceType = "github"
)

type Dex struct {
	Source     string        `json:"source"`
	SourceType DexSourceType `json:"sourceType"`
	Commit     string        `json:"commit"`
	Nodes      NodeMap       `json:"nodes"`

  gitCommit *object.Commit `json:"-"`
}

func NewDex(sourceType DexSourceType, source string) Dex {
  return Dex{
    Source: source,
    SourceType: sourceType,
  }
}

func (d *Dex) Hydrate() error {
  switch d.SourceType {
  case DexSourceGithub:
    return d.hydrateFromGithub()

  default:
    return ErrUnsupportedSourceType
  }
}

func (d *Dex) hydrateFromGithub() error {
  url := fmt.Sprintf("https://github.com/%v.git", d.Source)
  commit, err := getRepoHead(url, nil)
  if err != nil {
    return err
  }

  d.gitCommit = commit
  return nil
}

func (d *Dex) kegFromGithub() {
  
}

func ReadDexEntry(owner string, repo string, id int) (string, error) {
	commit, err := getRepoHead(owner, repo)
	if err != nil {
		return "", err
	}

	bd, err := getBasedir(commit)
	if err != nil {
		return "", err
	}

	var path string
	if bd == "" {
		path = fmt.Sprintf("%v/README.md", id)
	} else {
		path = fmt.Sprintf("%v/%v/README.md", bd, id)
	}

	entryFile, err := commit.File(path)
	if err != nil {
		return "", err
	}

	entry, err := entryFile.Contents()
	if err != nil {
		return "", err
	}

	return entry, nil
}

func GetDexFromRepo(owner string, repo string) (Dex, error) {
	d := Dex{}

	commit, err := getRepoHead(owner, repo)
	if err != nil {
		return Dex{}, fmt.Errorf("error getting repo head: %v", err)
	}

	bd, err := getBasedir(commit)
	if err != nil {
		return Dex{}, fmt.Errorf("error getting base dir: %v", err)
	}

	var path string
	if bd == "" {
		path = "dex/nodes.tsv"
	} else {
		path = fmt.Sprintf("%v/dex/nodes.tsv", bd)
	}

	dexFile, err := commit.File(path)
	if err != nil {
		return Dex{}, fmt.Errorf("error reading dex file: %v", err)
	}

	dex, err := dexFile.Contents()
	if err != nil {
		return Dex{}, fmt.Errorf("error reading tsv file: %v", err)
	}

	sr := strings.NewReader(dex)
	r := csv.NewReader(sr)
	r.Comma = '\t'
	r.LazyQuotes = true

	data, err := r.ReadAll()
	if err != nil {
		return Dex{}, fmt.Errorf("error reading tsv entries: %v", err)
	}

	nodes, err := NodeSliceFromTsv(data)
	if err != nil {
		return Dex{}, fmt.Errorf("error getting node slice from tsv: %v", err)
	}

	m := nodes.ToNodeMap()

	d.Commit = commit.Hash.String()
	d.Nodes = m

	return d, nil
}

func getRepoHead(url string, branch *string) (*object.Commit, error) {
	s := memory.NewStorage()

  opts := git.CloneOptions{
		URL:   url,
		Tags:  git.NoTags,
		Depth: 1,
	}

  if branch != nil {
    opts.ReferenceName = plumbing.NewBranchReferenceName(*branch)
    opts.SingleBranch = true
  }

	e, err := git.Clone(s, nil, &opts)
	if err != nil {
		return nil, fmt.Errorf("error cloning keg repo in getRepoHead: %v", err)
	}

	head, err := e.Head()
	if err != nil {
		return nil, fmt.Errorf("error getting HEAD hash in getRepoHead: %v", err)
	}

	obj, err := e.CommitObject(head.Hash())
	if err != nil {
		return nil, fmt.Errorf("error getting HEAD commit in getRepoHead: %v", err)
	}

	return obj, nil
}

func getBasedir(commit *object.Commit) (string, error) {
	dir := "UNSET"

	for _, p := range kegSearchPaths {
		var err error

		if p == "" {
			log.Println("searching root")
			_, err = commit.File("keg")
		} else {
			log.Printf("searching [%v]", p)
			_, err = commit.File(fmt.Sprintf("%v/keg", p))
		}

		if err != nil {
			continue
		} else {
			return p, nil
		}
	}

	if dir == "UNSET" {
		return "", ErrNoKegFound
	}

	return dir, nil
}
