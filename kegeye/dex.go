package kegeye

import (
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
)

type Dex struct {
  Commit string `json:"commit"`
  Nodes NodeMap `json:"nodes"`
}

func ReadDexEntry(owner string, repo string, id int) (string, error) {
  commit, err := getRepoHead(owner, repo)
  if err != nil {
    return "", err
  }

  entryFile, err := commit.File(fmt.Sprintf("docs/%v/README.md", id))
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
    return Dex{}, err
  }

  dexFile, err := commit.File("docs/dex/nodes.tsv")
  if err != nil {
    return Dex{}, err
  }

  dex, err := dexFile.Contents()
  if err != nil {
    return Dex{}, err
  }

  sr := strings.NewReader(dex)
  r := csv.NewReader(sr)
  r.Comma = '\t'

  data, err := r.ReadAll()
  if err != nil {
    return Dex{}, err
  }

  nodes, err := NodeSliceFromTsv(data)
  if err != nil {
    return Dex{}, err
  }

  m := nodes.ToNodeMap()

  d.Commit = commit.Hash.String()
  d.Nodes = m

  return d, nil
}

func getRepoHead(owner string, repo string) (*object.Commit, error) {
  s := memory.NewStorage()

  e, err := git.Clone(s, nil, &git.CloneOptions{
    URL: fmt.Sprintf("https://github.com/%v/%v.git", owner, repo),
    Tags: git.NoTags,
    Depth: 1,
  })
  if err != nil {
    return nil, err
  }

  head, err := e.Head()
  if err != nil {
    return nil, err
  }

  obj, err := e.CommitObject(head.Hash())
  if err != nil {
    return nil, err
  }

  return obj, nil
}
