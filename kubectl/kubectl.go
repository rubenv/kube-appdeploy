package kubectl

import (
	"bytes"
	"fmt"
	"os/exec"

	"k8s.io/client-go/rest"
)

type KubeCtl struct {
	config    *rest.Config
	namespace string
}

func NewKubeCtl(config *rest.Config, namespace string) *KubeCtl {
	return &KubeCtl{
		config:    config,
		namespace: namespace,
	}
}

func (t *KubeCtl) Run(stdin []byte, args ...string) (string, error) {
	args = append(t.configArgs(), args...)

	cmd := exec.Command("kubectl", args...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	out, err := cmd.Output()
	if err != nil {
		errmsg := err.Error()
		exiterr, ok := err.(*exec.ExitError)
		if ok {
			errmsg = string(exiterr.Stderr)
		}
		return "", fmt.Errorf("Kubectl %v failed: %s, %s", args, errmsg, out)
	}
	return string(out), nil
}

func (t *KubeCtl) configArgs() []string {
	args := []string{
		"--namespace", t.namespace,
	}

	cfg := t.config
	if cfg.Host != "" {
		args = append(args, "--server", cfg.Host)
	}
	if cfg.CAFile != "" {
		args = append(args, "--certificate-authority", cfg.CAFile)
	}
	if cfg.CertFile != "" {
		args = append(args, "--client-certificate", cfg.CertFile)
	}
	if cfg.CertFile != "" {
		args = append(args, "--client-key", cfg.KeyFile)
	}
	if cfg.BearerToken != "" {
		args = append(args, "--token", cfg.BearerToken)
	}

	return args
}
