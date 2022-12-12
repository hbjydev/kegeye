package main

import (
	"time"

	"github.com/Netflix/go-env"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/hbjydev/kegeye/handlers"
	Z "github.com/rwxrob/bonzai/z"
	"go.uber.org/zap"
)

var cmd = &Z.Cmd{
	Name: "kegeye",
	Call: func(_ *Z.Cmd, _ ...string) (err error) {
		var environment Environment
		if _, err := env.UnmarshalFromEnviron(&environment); err != nil {
			return err
		}

		var l *zap.Logger
		if environment.Env == "production" {
			gin.SetMode(gin.ReleaseMode)

			l, err = zap.NewProduction()
			if err != nil {
				return
			}
		} else {
			gin.SetMode(gin.DebugMode)

			l, err = zap.NewDevelopment()
			if err != nil {
				return
			}
		}

		e := gin.New()

		e.Use(ginzap.Ginzap(l, time.RFC3339, true))
		e.Use(ginzap.RecoveryWithZap(l, false))

    handlers.RegisterDexHandler(e, l)

		//e.GET("/keg/:owner/:repo/:id", func(c *gin.Context) {
		//	owner := c.Param("owner")
		//	repo := c.Param("repo")
		//	idStr := c.Param("id")
		//	id, err := strconv.Atoi(idStr)
		//	if err != nil {
		//		c.AbortWithStatusJSON(400, gin.H{"error": "invalid id format, must be int"})
		//	}

		//	data, err := kegeye.ReadDexEntry(owner, repo, id)
		//	if err != nil {
		//		_ = c.Error(err)
		//		e := c.AbortWithError(500, errors.New("invalid entry"))
		//		_ = e.SetType(gin.ErrorTypePublic)
		//		return
		//	}

		//	c.Header("Content-Type", "text/markdown")
		//	c.String(200, data)
		//})

		return e.Run()
	},
}

type Environment struct {
	Env         string `env:"KE_ENV,default=production,required=true"`
	GithubToken string `env:"KE_GITHUB_TOKEN"`
}

func main() {
	cmd.Run()
}
