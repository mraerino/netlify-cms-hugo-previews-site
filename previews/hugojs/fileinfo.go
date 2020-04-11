package hugojs

import (
	"os"
	"syscall/js"
	"time"

	"github.com/pkg/errors"
)

type jsFileInfo struct {
	name  string
	size  int64
	isDir bool
}

func fileInfoFromValue(val js.Value) (*jsFileInfo, error) {
	if val.Type() != js.TypeObject {
		return nil, errors.New("invalid type, expected object")
	}

	obj := new(jsFileInfo)

	jsName := val.Get("name")
	if jsName.Type() != js.TypeString {
		return nil, errors.New("invalid type, expected string")
	}
	obj.name = jsName.String()

	jsSize := val.Get("size")
	if jsSize.Type() != js.TypeNumber {
		return nil, errors.New("invalid type, expected number")
	}
	obj.size = int64(jsSize.Int())

	obj.isDir = val.Get("isDir").Truthy()

	return obj, nil
}

func (fi *jsFileInfo) Name() string {
	return fi.name
}

func (fi *jsFileInfo) Size() int64 {
	return fi.size
}

func (fi *jsFileInfo) Mode() os.FileMode {
	return os.ModePerm
}

func (fi *jsFileInfo) ModTime() time.Time {
	return time.Now()
}

func (fi *jsFileInfo) IsDir() bool {
	return fi.isDir
}

func (fi *jsFileInfo) Sys() interface{} {
	return nil
}
