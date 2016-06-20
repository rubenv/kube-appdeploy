package appdeploy

import (
	"io"
	"testing"

	"github.com/cheekybits/is"
)

type TestSource struct {
}

var _ ManifestSource = &TestSource{}

func (t *TestSource) Names() ([]string, error) {
	return []string{}, nil
}

func (t *TestSource) Get(name string) (io.ReadCloser, error) {
	panic("not implemented")
}

type TestTarget struct {
	prepareCalled bool
	cleanupCalled bool
}

var _ Target = &TestTarget{}

func (t *TestTarget) Prepare() error {
	t.prepareCalled = true
	return nil
}

func (t *TestTarget) Apply(m Manifest, data []byte) error {
	panic("not implemented")
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
}
