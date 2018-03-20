package appdeploy

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"gopkg.in/yaml.v1"
)

type ManifestSource interface {
	Names() ([]string, error)
	Get(name string) (io.ReadCloser, error)
	Variables() (*ProcessVariables, error)
}

var _ ManifestSource = &FolderSource{}

type FolderSource struct {
	Path      string
	variables *ProcessVariables
	names     []string
}

func NewFolderSource(p string) (*FolderSource, error) {
	src := &FolderSource{
		Path:      p,
		variables: NewProcessVariables(),
	}

	files, err := ioutil.ReadDir(src.Path)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0)
	for _, file := range files {
		name := file.Name()
		if name == "variables.yaml" {
			vars := NewProcessVariables()

			data, err := ioutil.ReadFile(path.Join(src.Path, name))
			if err != nil {
				return nil, err
			}

			err = yaml.Unmarshal(data, vars.Variables)
			if err != nil {
				return nil, err
			}

			src.variables = vars
			continue
		}

		if !file.IsDir() && strings.HasSuffix(name, ".yaml") {
			names = append(names, name)
		}
	}
	src.names = names

	return src, nil
}

func (s *FolderSource) Names() ([]string, error) {
	return s.names, nil
}

func (s *FolderSource) Get(name string) (io.ReadCloser, error) {
	path := path.Join(s.Path, name)
	return os.Open(path)
}

func (s *FolderSource) Variables() (*ProcessVariables, error) {
	return s.variables, nil
}

func (s *FolderSource) SetVariables(variables *ProcessVariables) {
	s.variables = variables
}
