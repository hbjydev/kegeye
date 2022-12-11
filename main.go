package main

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hbjydev/kegeye/kegeye"
	Z "github.com/rwxrob/bonzai/z"
)

var cmd = &Z.Cmd{
  Name: "kegeye",
  Call: func(_ *Z.Cmd, _ ...string) error {
    e := gin.Default()

    e.GET("/", func (c *gin.Context) {
      c.JSON(200, gin.H{ "hello": "world" })
    })

    e.GET("/keg/:owner/:repo", func (c *gin.Context) {
      owner := c.Param("owner")
      repo := c.Param("repo")

      data, err := kegeye.GetDexFromRepo(owner, repo)
      if err != nil {
        _ = c.Error(err)
        _ = c.AbortWithError(500, errors.New("invalid dex format"))
        return
      }

      c.JSON(200, gin.H{"dex": data})
    })

    e.GET("/keg/:owner/:repo/:id", func (c *gin.Context) {
      owner := c.Param("owner")
      repo := c.Param("repo")
      idStr := c.Param("id")
      id, err := strconv.Atoi(idStr)
      if err != nil {
        c.AbortWithStatusJSON(400, gin.H{"error": "invalid id format, must be int"})
      }

      data, err := kegeye.ReadDexEntry(owner, repo, id)
      if err != nil {
        _ = c.Error(err)
        _ = c.AbortWithError(500, errors.New("invalid entry"))
        return
      }

      c.Header("Content-Type", "text/markdown")
      c.String(200, data)
    })


    return e.Run()
  },
}

func main() {
  cmd.Run()
}
