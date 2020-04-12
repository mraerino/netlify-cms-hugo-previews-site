package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/fsnotify/fsnotify"
	"github.com/gohugoio/hugo/common/loggers"
	"github.com/gohugoio/hugo/deps"
	"github.com/gohugoio/hugo/hugofs"
	"github.com/gohugoio/hugo/hugolib"
	nutil "github.com/netlify/netlify-commons/util"
	"github.com/spf13/afero"
)

type previewAPI struct {
	hugo  *hugolib.HugoSites
	memfs afero.Fs

	initialBuildDone nutil.AtomicBool
}

func writeFiles(fs afero.Fs, files map[string]string) error {
	for name, content := range files {
		err := afero.WriteFile(fs, name, []byte(content), os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func newPreviewAPI() (*previewAPI, error) {
	mm := afero.NewMemMapFs()
	// js, err := hugojs.NewJSFS(backend)
	// if err != nil {
	// 	return err
	// }

	// h.fs = afero.NewCopyOnWriteFs(js, mm)

	files := map[string]string{
		"/config.yml": `
languageCode = "en-us"
title = "Hugo Previews Demo Site"
		`,
		"/content/_index.md": `
---
title: "Homepage"
---
# Hello World!
		`,
		"/layouts/index.html": `
<div class="homepage-content">
	{{.Content}}
</div>
		`,
	}
	if err := writeFiles(mm, files); err != nil {
		return nil, err
	}

	cfg, _, err := hugolib.LoadConfig(hugolib.ConfigSourceDescriptor{
		Fs:         mm,
		Filename:   "/config.yaml",
		WorkingDir: "/",
	})
	if err != nil {
		return nil, err
	}

	// required so public files are actually written
	fs := hugofs.NewFrom(afero.NewBasePathFs(mm, "/"), cfg)

	deps := deps.DepsCfg{
		Fs:     fs,
		Cfg:    cfg,
		Logger: loggers.NewDebugLogger(),
	}

	site, err := hugolib.NewHugoSites(deps)
	if err != nil {
		return nil, err
	}

	return &previewAPI{
		hugo:  site,
		memfs: mm,

		initialBuildDone: nutil.NewAtomicBool(false),
	}, nil
}

func (a *previewAPI) build(path string) error {
	partialBuild := a.initialBuildDone.Get()
	var events []fsnotify.Event
	if partialBuild {
		events = append(events, fsnotify.Event{
			Name: path,
			Op:   fsnotify.Write,
		})
	}

	err := a.hugo.Build(hugolib.BuildCfg{}, events...)
	if err != nil {
		return err
	}

	if !partialBuild {
		a.initialBuildDone.Set(true)
	}
	return nil
}

type payload struct {
	Path string `json:"path"`
}

func errResp(code int, msg string, err error) (*events.APIGatewayProxyResponse, error) {
	if err != nil {
		fmt.Printf("%+v\n", err)
	}
	return &events.APIGatewayProxyResponse{
		StatusCode: code,
		Body:       msg,
	}, nil
}

func (a *previewAPI) handler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	if request.HTTPMethod != http.MethodPost {
		return errResp(http.StatusBadRequest, "Only POST is allowed", nil)
	}

	pl := new(payload)
	if err := json.Unmarshal([]byte(request.Body), pl); err != nil {
		return errResp(http.StatusBadRequest, "Failed to read request body", err)
	}

	err := a.build(pl.Path)
	if err != nil {
		return errResp(http.StatusInternalServerError, "Failed to render site", err)
	}

	var publicPath string
	for _, page := range a.hugo.Pages() {
		if !page.File().IsZero() && page.Filename() == pl.Path {
			publicPath = page.RelPermalink()
			if strings.HasSuffix(publicPath, "/") {
				publicPath += "index.html"
			}
			break
		}
	}

	if publicPath == "" {
		return errResp(http.StatusNotFound, "Failed to find public path", nil)
	}

	content, err := afero.ReadFile(a.memfs, filepath.Join("/public", publicPath))
	if err != nil {
		return errResp(http.StatusInternalServerError, "Failed to read content", err)
	}

	return &events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "text/html",
		},
		Body: string(content),
	}, nil
}

func main() {
	api, err := newPreviewAPI()
	if err != nil {
		panic(err)
	}
	lambda.Start(api.handler)
}
