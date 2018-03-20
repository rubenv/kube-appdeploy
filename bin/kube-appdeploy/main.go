package main

import (
	"fmt"
	"log"
	"strings"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	flags "github.com/jessevdk/go-flags"
	"github.com/rubenv/kube-appdeploy"
)

func main() {
	err := do()
	if err != nil {
		log.Fatal(err)
	}
}

type GlobalOptions struct {
	Context     string            `short:"c" long:"context" description:"Kubernetes context to use"`
	Variables   map[string]string `short:"v" long:"variable" description:"Extra variables to set"`
	PullSecrets []string          `short:"i" long:"imagepullsecret" description:"Image pull secrets to configure"`

	Args struct {
		Folder string `positional-arg-name:"folder" description:"Path to the configuration files"`
	} `positional-args:"yes" required:"yes"`
}

var globalOpts = &GlobalOptions{}
var parser = flags.NewParser(globalOpts, flags.Default)

func do() error {
	_, err := parser.Parse()
	if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
		return nil
	}
	if err != nil {
		return err
	}

	var target appdeploy.Target
	src, err := appdeploy.NewFolderSource(globalOpts.Args.Folder)
	if err != nil {
		return err
	}

	vars, err := src.Variables()
	if err != nil {
		return err
	}

	for k, v := range globalOpts.Variables {
		vars.Variables[k] = v
	}
	vars.ImagePullSecrets = globalOpts.PullSecrets

	contextName := globalOpts.Context

	// Prepare Kubernetes client
	po := clientcmd.NewDefaultPathOptions()

	c, err := po.GetStartingConfig()
	if err != nil {
		return err
	}

	if contextName == "" {
		contextName = c.CurrentContext
	}

	context, ok := c.Contexts[contextName]
	if !ok {
		names := make([]string, 0)
		for name, _ := range c.Contexts {
			names = append(names, name)
		}

		return fmt.Errorf("Unknown context: %s, should be one of: %s", contextName, strings.Join(names, ", "))
	}

	authinfo, ok := c.AuthInfos[context.AuthInfo]
	if !ok {
		return fmt.Errorf("Badly configured context, unknown auth: %s", context.AuthInfo)
	}

	cluster, ok := c.Clusters[context.Cluster]
	if !ok {
		return fmt.Errorf("Badly configured context, unknown cluster: %s", context.Cluster)
	}

	config := &rest.Config{
		Host:        cluster.Server,
		BearerToken: authinfo.Token,
		TLSClientConfig: rest.TLSClientConfig{
			CAFile:   cluster.CertificateAuthority,
			CertFile: authinfo.ClientCertificate,
			KeyFile:  authinfo.ClientKey,
		},
	}

	target = appdeploy.NewKubernetesTarget(config)

	return appdeploy.Process(src, target)
}
