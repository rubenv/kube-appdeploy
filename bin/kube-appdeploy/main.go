package main

import (
	"log"
	"os"

	"github.com/rubenv/kube-appdeploy"
)

func main() {
	src := appdeploy.NewFolderSource(os.Args[1])

	opts := appdeploy.Options{
		Mode:         appdeploy.WriteToFolder,
		OutputFolder: "/Users/ruben/Desktop/out",
	}

	var target appdeploy.Target

	switch opts.Mode {
	case appdeploy.WriteToFolder:
		if opts.OutputFolder == "" {
			log.Fatal("No output folder specified")
		}

		target = appdeploy.NewFolderTarget(opts.OutputFolder)
	}

	err := appdeploy.Process(src, target)
	if err != nil {
		log.Fatal(err)
	}
}
