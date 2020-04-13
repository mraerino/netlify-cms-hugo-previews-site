package main

import (
	"fmt"
	"os"

	"github.com/gohugoio/hugo/common/loggers"
	"github.com/gohugoio/hugo/deps"
	"github.com/gohugoio/hugo/hugofs"
	"github.com/gohugoio/hugo/hugolib"
	"github.com/mraerino/netlify-cms-hugo-previews-site/previews/githubfs"
	"github.com/mraerino/netlify-cms-hugo-previews-site/previews/helpers"
	"github.com/spf13/afero"
)

func main() {
	ghFS, err := githubfs.New(
		os.Getenv("HUGO_PREVIEW_GITHUB_TOKEN"),
		os.Getenv("HUGO_PREVIEW_GITHUB_REPO"),
		"",
	)
	if err != nil {
		panic(err)
	}

	mm := afero.NewMemMapFs()
	cachedFs := afero.NewCacheOnReadFs(ghFS, mm, 0)
	fs := afero.NewCopyOnWriteFs(cachedFs, mm)

	cfg, _, err := hugolib.LoadConfig(hugolib.ConfigSourceDescriptor{
		Fs:         fs,
		Filename:   "config.yaml",
		WorkingDir: "/",
	})
	if err != nil {
		panic(err)
	}

	// BasePathFs is required so public files are actually written
	//hugoFs := hugofs.NewFrom(afero.NewBasePathFs(fs, "/"), cfg)
	hugoFs := hugofs.NewFrom(fs, cfg)
	deps := deps.DepsCfg{
		Fs:     hugoFs,
		Cfg:    cfg,
		Logger: loggers.NewDebugLogger(),
	}

	site, err := hugolib.NewHugoSites(deps)
	if err != nil {
		panic(err)
	}

	if err := helpers.ReplaceBaseOf(site, mm, "_index.md", "", ""); err != nil {
		panic(err)
	}

	if err := site.Build(hugolib.BuildCfg{}); err != nil {
		panic(err)
	}

	content, err := afero.ReadFile(mm, "public/index.html")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(content))
}
