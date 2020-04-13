package main

import (
	"fmt"
	"os"

	"github.com/gohugoio/hugo/common/loggers"
	"github.com/gohugoio/hugo/deps"
	"github.com/gohugoio/hugo/hugofs"
	"github.com/gohugoio/hugo/hugolib"
	"github.com/spf13/afero"
)

func main() {
	// ghFS, err := githubfs.New(
	// 	os.Getenv("HUGO_PREVIEW_GITHUB_TOKEN"),
	// 	os.Getenv("HUGO_PREVIEW_GITHUB_REPO"),
	// 	"",
	// )
	// if err != nil {
	// 	panic(err)
	// }
	osFs := afero.NewOsFs()

	mm := afero.NewMemMapFs()
	cachedFs := afero.NewCacheOnReadFs(osFs, mm, 0)
	fs := afero.NewCopyOnWriteFs(cachedFs, mm)

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	cfg, _, err := hugolib.LoadConfig(hugolib.ConfigSourceDescriptor{
		Fs:         fs,
		Filename:   "config.yaml",
		WorkingDir: cwd,
	})
	if err != nil {
		panic(err)
	}

	cfg.Set("buildDrafts", true)
	cfg.Set("buildFuture", true)
	cfg.Set("buildExpired", true)
	cfg.Set("environment", "preview")

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

	if err := site.Build(hugolib.BuildCfg{}); err != nil {
		panic(err)
	}

	content, err := afero.ReadFile(mm, "public/index.html")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(content))
}
