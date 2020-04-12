package githubfs

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/google/go-github/v31/github"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"golang.org/x/oauth2"
)

type githubFS struct {
	client *github.Client

	repoOwner string
	repoName  string
	branch    string
}

// New creates an FS to get files from github on-demand
func New(accessToken string, repo string, branch string) (afero.Fs, error) {
	repoParts := strings.Split(repo, "/")
	if len(repoParts) != 2 {
		return nil, errors.New("invalid repo path, expected owner/repo style")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	return &githubFS{
		client:    client,
		repoOwner: repoParts[0],
		repoName:  repoParts[1],
		branch:    branch,
	}, nil
}

func (fs *githubFS) Open(name string) (afero.File, error) {
	fmt.Printf("Opening file: %s\n", name)
	path := strings.TrimPrefix(name, "/")
	if strings.HasPrefix(path, "layouts/layouts") {
		fmt.Println("bla!")
	}

	file, dir, resp, err := fs.client.Repositories.GetContents(
		context.Background(),
		fs.repoOwner,
		fs.repoName,
		path,
		&github.RepositoryContentGetOptions{Ref: fs.branch},
	)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			fmt.Printf("not found: %s\n", name)
			return nil, syscall.ENOENT // afero tries this in the copy on write impl
		}
		return nil, err
	}

	if file != nil {
		content, err := file.GetContent()
		if err != nil {
			return nil, errors.Wrap(err, "failed to read content")
		}
		return &githubFile{
			isDir: false,
			path:  file.GetPath(),
			size:  int64(file.GetSize()),
			body:  bytes.NewReader([]byte(content)),
		}, nil
	}

	if dir != nil {
		return &githubFile{
			isDir: true,
			path:  name,
			files: dir,
		}, nil
	}

	return nil, errors.New("neither file nor dir returned")
}

func (fs *githubFS) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if flag&(os.O_APPEND|os.O_CREATE|os.O_RDWR|os.O_TRUNC|os.O_WRONLY) != 0 {
		return nil, ErrReadOnly
	}

	return fs.Open(name)
}

func (fs *githubFS) Stat(name string) (os.FileInfo, error) {
	file, err := fs.Open(name)
	if err != nil {
		return nil, err
	}

	return file.Stat()
}

func (githubFS) Name() string {
	return "GithubFS"
}

// write operation are not supported, those are below
var (
	ErrReadOnly = errors.New("Operation not supported: FS is read-only")
)

func (fs *githubFS) Create(name string) (afero.File, error) {
	return nil, ErrReadOnly
}

func (fs *githubFS) Mkdir(name string, perm os.FileMode) error {
	return ErrReadOnly
}

func (fs *githubFS) MkdirAll(path string, perm os.FileMode) error {
	return ErrReadOnly
}

func (fs *githubFS) Remove(name string) error {
	return ErrReadOnly
}

func (fs *githubFS) RemoveAll(path string) error {
	return ErrReadOnly
}

func (fs *githubFS) Rename(oldname, newname string) error {
	return ErrReadOnly
}

func (fs *githubFS) Chmod(name string, mode os.FileMode) error {
	return ErrReadOnly
}

func (fs *githubFS) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return ErrReadOnly
}
