package appdeploy

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"sync"
	"text/template"

	"gopkg.in/yaml.v1"
)

type Manifest struct {
	Kind     string
	Metadata Metadata
}

func (m Manifest) Filename(folder string) string {
	name := fmt.Sprintf("%s--%s.yaml", strings.ToLower(m.Kind), m.Metadata.Name)
	if folder != "" {
		return path.Join(folder, name)
	} else {
		return name
	}
}

type Metadata struct {
	Name string
}

type ProcessVariables struct {
	Namespace        string
	ImagePullSecrets []string
	ManageCronjobs   bool
	Variables        map[string]interface{}
}

func NewProcessVariables() *ProcessVariables {
	return &ProcessVariables{
		Variables: make(map[string]interface{}),
	}
}

func Process(src ManifestSource, target Target) error {
	return ProcessWithFuncs(src, target, nil)
}

func ProcessWithFuncs(src ManifestSource, target Target, funcs template.FuncMap) error {
	names, err := src.Names()
	if err != nil {
		return err
	}

	if len(names) == 0 {
		return errors.New("Cowardly refusing a source with 0 manifests, this would delete everything")
	}

	vars, err := src.Variables()
	if err != nil {
		return err
	}

	if vars == nil {
		vars = NewProcessVariables()
	}

	// Some variables are special, extract them
	if v, ok := vars.Variables["namespace"]; ok {
		if s, ok := v.(string); ok {
			vars.Namespace = s
		}
	}
	if vars.Namespace == "" {
		vars.Namespace = "default"
	}

	// Prepare the target environment
	err = target.Prepare(vars)
	if err != nil {
		return err
	}

	seen := make([]Manifest, 0)
	var seenMutex sync.Mutex
	wg := sync.WaitGroup{}
	wg.Add(len(names))

	// Apply all resources in parallel
	for _, name := range names {
		n := name
		go func() {
			defer wg.Done()
			m, e := process(src, vars, n, target, funcs)
			if e != nil {
				err = e
			}
			if m != nil {
				seenMutex.Lock()
				seen = append(seen, *m)
				seenMutex.Unlock()
			}
		}()
	}

	wg.Wait()
	if err != nil {
		return err
	}

	return target.Cleanup(seen)
}

func process(src ManifestSource, vars *ProcessVariables, name string, target Target, funcs template.FuncMap) (*Manifest, error) {
	m, err := src.Get(name)
	if err != nil {
		return nil, err
	}
	defer m.Close()

	// Read and parse template
	data, err := ioutil.ReadAll(m)
	if err != nil {
		return nil, err
	}

	tpl, err := template.New(name).Funcs(funcs).Parse(string(data))
	if err != nil {
		return nil, err
	}

	// Execute it
	var buf bytes.Buffer
	err = tpl.Execute(&buf, vars)
	if err != nil {
		return nil, err
	}

	data = bytes.TrimSpace(buf.Bytes())
	if string(data) == "" {
		// Nothing here (entire manifest in a false if-block?
		return nil, nil
	}

	// Determine object type
	var manifest Manifest
	err = yaml.Unmarshal(data, &manifest)
	if err != nil {
		return nil, err
	}

	if manifest.Kind == "" || manifest.Metadata.Name == "" {
		return nil, fmt.Errorf("%s: missing type data, not a valid Kubernetes manifest?", name)
	}

	err = target.Apply(manifest, data)
	if err != nil {
		return nil, err
	}
	return &manifest, nil
}
