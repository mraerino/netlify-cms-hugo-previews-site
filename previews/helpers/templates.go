package helpers

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/gohugoio/hugo/hugolib"
	"github.com/gohugoio/hugo/output"
	"github.com/gohugoio/hugo/resources/page"
	"github.com/spf13/afero"
)

var layoutHandler = output.NewLayoutHandler()

func FindTemplate(hugo *hugolib.HugoSites, d output.LayoutDescriptor) (string, error) {
	candidates, err := layoutHandler.For(d, output.HTMLFormat)
	if err != nil {
		return "", err
	}

	var tplName string
	for _, candidate := range candidates {
		if hugo.Tmpl().HasTemplate(candidate) {
			tplName = candidate
			break
		}
	}
	return tplName, nil
}

func templatePath(path string) string {
	return filepath.Join("layouts", path)
}

func ReplaceBaseOf(hugo *hugolib.HugoSites, fs afero.Fs, contentPath, layout, fmType string) error {
	parts := strings.Split(contentPath, "/")
	section := ""
	if len(parts) > 1 {
		section = parts[0]
	}

	kind := page.KindPage
	if len(parts) == 1 && strings.HasPrefix(contentPath, "_index.") {
		kind = page.KindHome
	}

	tplType := "page"
	if section != "" {
		tplType = section
	}
	if fmType != "" {
		tplType = fmType
	}

	baseofDescr := output.LayoutDescriptor{
		Baseof:  true,
		Kind:    kind,
		Type:    tplType,
		Layout:  layout,
		Section: section,
	}
	baseofName, err := FindTemplate(hugo, baseofDescr)
	if err != nil {
		return err
	}
	if baseofName == "" {
		return nil
	}

	previewDescr := output.LayoutDescriptor{
		Kind:    page.KindPage,
		Type:    tplType,
		Layout:  "cms-preview",
		Section: section,
	}
	previewName, err := FindTemplate(hugo, previewDescr)
	if err != nil {
		return err
	}

	if previewName == "" {
		return nil
	}

	previewFile, err := fs.Open(templatePath(previewName))
	if err != nil {
		return err
	}

	baseofFile, err := fs.Open(templatePath(baseofName))
	if err != nil {
		return err
	}

	_, err = io.Copy(baseofFile, previewFile)
	return err
}
