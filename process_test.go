package appdeploy

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/cheekybits/is"
)

type TestSource struct {
}

var _ ManifestSource = &TestSource{}

type CloseBuffer struct {
	bytes.Buffer
}

func (b *CloseBuffer) Close() error {
	return nil
}

func (t *TestSource) Names() ([]string, error) {
	return []string{"test.yaml"}, nil
}

func (t *TestSource) Get(name string) (io.ReadCloser, error) {
	buf := &CloseBuffer{}

	switch name {
	case "test.yaml":
		buf.WriteString(`
kind: Service
metadata:
  name: test
  hello: {{ .Variables.name }}
{{ if false }}
NOT HERE
{{ end }}`)
	default:
		return nil, fmt.Errorf("Unknown file: %s", name)
	}

	return buf, nil
}

func (t *TestSource) Variables() (*ProcessVariables, error) {
	return &ProcessVariables{
		Variables: map[string]interface{}{
			"name": "world",
		},
	}, nil
}

type TestTarget struct {
	prepareCalled bool
	cleanupCalled bool
	applied       map[string]string
}

var _ Target = &TestTarget{}

func (t *TestTarget) Prepare() error {
	t.prepareCalled = true
	return nil
}

func (t *TestTarget) Apply(m Manifest, data []byte) error {
	if t.applied == nil {
		t.applied = make(map[string]string)
	}

	t.applied[m.Filename("")] = string(data)
	return nil
}

func (t *TestTarget) Cleanup(items []Manifest) error {
	t.cleanupCalled = true
	return nil
}

func TestProcess(t *testing.T) {
	is := is.New(t)

	src := &TestSource{}
	target := &TestTarget{}

	err := Process(src, target)
	is.NoErr(err)

	is.True(target.prepareCalled)
	is.True(target.cleanupCalled)
	is.Equal(len(target.applied), 1)
	is.False(strings.Contains(target.applied["service--test.yaml"], "NOT HERE"))
	is.True(strings.Contains(target.applied["service--test.yaml"], "hello: world"))
}
