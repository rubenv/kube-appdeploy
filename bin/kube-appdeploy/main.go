package main

import (
	"log"
	"os"

	"github.com/rubenv/kube-appdeploy"
)

func main() {
	src := appdeploy.NewFolderSource(os.Args[1])
	folder := "/Users/ruben/Desktop/out"

	var target appdeploy.Target

	if folder == "" {
		log.Fatal("No output folder specified")
	}

	target = appdeploy.NewFolderTarget(folder)

	err := appdeploy.Process(src, target)
	if err != nil {
		log.Fatal(err)
	}
}
