package handlers

import (
	"context"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/google/go-github/v48/github"
	"github.com/hbjydev/kegeye/keg"
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

  // Clone the branch set by the middleware
  s := memory.NewStorage()
  r, err := git.Clone(s, nil, &git.CloneOptions{
    URL: cloneUrl,
    ReferenceName: plumbing.NewBranchReferenceName(branch),
    Depth: 1,
    SingleBranch: true,
  })

  if err != nil {
    _ = c.AbortWithError(500, err)
    return
  }

  // Get the current state of <default branch>
  head, err := r.Head()
  if err != nil {
    _ = c.AbortWithError(500, err)
    return
  }

  // Get the commit object for the latest commit to <default branch>
  commit, err := r.CommitObject(head.Hash())
  if err != nil {
    _ = c.AbortWithError(500, err)
    return
  }

  file, err := commit.File("keg")
  if err != nil {
    // If the keg is not at /, try /docs/
    if err.Error() == "file not found" {
      file, err = commit.File("docs/keg")
      if err != nil {
        _ = c.AbortWithError(500, err)
        return
      }
    } else {
      _ = c.AbortWithError(500, err)
      return
    }
  }

  contents, err := file.Contents()
  if err != nil {
    _ = c.AbortWithError(500, err)
    return
  }

  var keg keg.KegInfo
  if err := yaml.Unmarshal([]byte(contents), &keg); err != nil {
    _ = c.AbortWithError(500, err)
    return
  }

	c.JSON(200, gin.H{"keg": keg})
}
