package handlers

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/google/go-github/v48/github"
	"github.com/hbjydev/kegeye/keg"
	_keg "github.com/rwxrob/keg"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func RegisterDexHandler(e *gin.Engine, l *zap.Logger) {
	if os.Getenv("GITHUB_TOKEN") == "" {
		return
	}

	r := e.Group("/keg/github")
	{
		r.GET("/:owner/:repo", githubRepoExists(l), fetchKeg)
		r.GET("/:owner/:repo/dex", githubRepoExists(l), fetchDexEntries)
    r.GET("/:owner/:repo/nodes/:id", githubRepoExists(l), fetchDexEntry)
	}
}

func githubRepoExists(l *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		owner := c.Param("owner")
		name := c.Param("repo")

		client := github.NewClient(nil)
		repo, _, err := client.Repositories.Get(context.TODO(), owner, name)
		if err != nil {
			if strings.HasSuffix(err.Error(), ": 404 Not Found []") {
				c.AbortWithStatusJSON(404, gin.H{"error": "repository does not exist"})
				return
			}

			_ = c.AbortWithError(500, err)
			return
		}

		c.Set("cloneUrl", repo.GetCloneURL())
		c.Set("branch", repo.GetDefaultBranch())
	}
}

func fetchKeg(c *gin.Context) {
	cloneUrl := c.GetString("cloneUrl")
	branch := c.GetString("branch")

	keg, _, err := fetchKegInfo(cloneUrl, branch)
	if err != nil {
		_ = c.AbortWithError(500, err)
		return
	}

	c.JSON(200, gin.H{"keg": keg})
}

func fetchDexEntries(c *gin.Context) {
	cloneUrl := c.GetString("cloneUrl")
	branch := c.GetString("branch")

	keg, commit, err := fetchKegInfo(cloneUrl, branch)
	if err != nil {
		_ = c.AbortWithError(500, err)
		return
	}

  dex, err := fetchDex(keg, commit)
  if err != nil {
		_ = c.AbortWithError(500, err)
		return
  }

	c.JSON(200, dex.ByID())
}

func fetchDexEntry(c *gin.Context) {
	cloneUrl := c.GetString("cloneUrl")
	branch := c.GetString("branch")
  idStr := c.Param("id")

  id, err := strconv.Atoi(idStr)
  if err != nil {
		_ = c.AbortWithError(500, err)
		return
  }

	keg, commit, err := fetchKegInfo(cloneUrl, branch)
	if err != nil {
		_ = c.AbortWithError(500, err)
		return
	}

  dex, err := fetchDex(keg, commit)
  if err != nil {
		_ = c.AbortWithError(500, err)
		return
  }

  en := dex.Lookup(id)
  if en == nil {
    c.AbortWithStatusJSON(200, gin.H{"error": "no such entry found"})
		return
  }

  file, err := getKegFile(fmt.Sprintf("%v/README.md", en.ID()), commit)
  if err != nil {
    _ = c.AbortWithError(500, err)
    return
  }

  contents, err := file.Contents()
  if err != nil {
    _ = c.AbortWithError(500, err)
    return
  }

  c.String(200, contents)
}

func fetchDex(keg *keg.KegInfo, commit *object.Commit) (*_keg.Dex, error) {
  file, err := getKegFile("dex/changes.md", commit)
  if err != nil {
    return nil, err
  }

  contents, err := file.Contents()
  if err != nil {
    return nil, err
  }

  dex, err := _keg.ParseDex(contents) 
  if err != nil {
    return nil, err
  }

  return dex, nil
}

func fetchKegInfo(cloneUrl string, branch string) (*keg.KegInfo, *object.Commit, error) {
	// Clone the branch set by the middleware
	s := memory.NewStorage()
	r, err := git.Clone(s, nil, &git.CloneOptions{
		URL:           cloneUrl,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		Depth:         1,
		SingleBranch:  true,
	})

	if err != nil {
		return nil, nil, err
	}

	// Get the current state of <default branch>
	head, err := r.Head()
	if err != nil {
		return nil, nil, err
	}

	// Get the commit object for the latest commit to <default branch>
	commit, err := r.CommitObject(head.Hash())
	if err != nil {
		return nil, nil, err
	}

  file, err := getKegFile("keg", commit)
  if err != nil {
    return nil, nil, err
  }

	contents, err := file.Contents()
	if err != nil {
		return nil, nil, err
	}

	var keg keg.KegInfo
	if err := yaml.Unmarshal([]byte(contents), &keg); err != nil {
		return nil, nil, err
	}

	return &keg, commit, nil
}

func getKegFile(path string, commit *object.Commit) (*object.File, error) {
	file, err := commit.File(path)
	if err != nil {
		// If the keg is not at /, try /docs/
		if err.Error() == "file not found" {
			file, err = commit.File(fmt.Sprintf("docs/%v", path))
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

  return file, nil
}
