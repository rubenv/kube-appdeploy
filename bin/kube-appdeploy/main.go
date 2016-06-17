package main

import (
	"log"
	"os"

	"github.com/rubenv/kube-appdeploy"
)

func main() {
	src := appdeploy.NewFolderSource(os.Args[1])

	err := appdeploy.Process(src, appdeploy.Options{
		Mode:         appdeploy.WriteToFolder,
		OutputFolder: "/Users/ruben/Desktop/out",
	})
	if err != nil {
		log.Fatal(err)
	}
}
