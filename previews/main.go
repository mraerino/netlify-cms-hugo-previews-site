package main

import (
	"context"
	"errors"
	"syscall/js"

	"github.com/gohugoio/hugo/deps"
	"github.com/gohugoio/hugo/hugofs"
	"github.com/gohugoio/hugo/hugolib"
	"github.com/mraerino/netlify-cms-hugo-previews-site/previews/hugojs"
	"github.com/spf13/afero"
)

const previewTemplatePath = "/layouts/_default/cms_preview.html"
const baseofTemplatePath = "/layouts/_default/baseof.html"

type hugoInterop struct {
	fs   afero.Fs
	hugo *hugolib.HugoSites
}

func (h *hugoInterop) setBackendFs(backend js.Value) error {
	mm := afero.NewMemMapFs()
	js, err := hugojs.NewJSFS(backend)
	if err != nil {
		return err
	}

	h.fs = afero.NewCopyOnWriteFs(js, mm)

	cfg, _, err := hugolib.LoadConfig(hugolib.ConfigSourceDescriptor{
		Fs:         h.fs,
		Filename:   "/config.yaml",
		WorkingDir: "/",
	})
	if err != nil {
		return err
	}

	// required so public files are actually written
	fs := hugofs.NewFrom(afero.NewBasePathFs(h.fs, "/"), cfg)

	deps := deps.DepsCfg{
		Fs:  fs,
		Cfg: cfg,
		//Logger: loggers.NewDebugLogger(),
	}

	site, err := hugolib.NewHugoSites(deps)
	if err != nil {
		return err
	}
	h.hugo = site

	return nil
}

// func (h *hugoInterop) setContent(path string, content js.Value) error {
// 	return nil // todo
// }

func (h *hugoInterop) build(path string) error {
	if h.hugo == nil {
		return errors.New("Not initislized")
	}
	// todo: find out how to limit build to a path, likely via events
	return h.hugo.Build(hugolib.BuildCfg{})
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	interop := new(hugoInterop)

	previewGlobal := js.ValueOf(map[string]interface{}{
		"setBackendFs": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if len(args) < 1 {
				return "Invalid arguments: missing backend"
			}
			err := interop.setBackendFs(args[0])
			if err != nil {
				return err.Error()
			}
			return ""
		}),
		"build": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			err := interop.build("")
			if err != nil {
				return err.Error()
			}
			return ""
		}),
		"stop": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			cancel()
			return nil
		}),
	})

	js.Global().Set("HugoPreview", previewGlobal)

	<-ctx.Done()
}
