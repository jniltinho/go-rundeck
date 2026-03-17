package main

import (
	"go-rundeck/cmd"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	cmd.SetEmbeds(TemplatesFS, StaticFS)
	cmd.Execute(Version, BuildTime, GitCommit)
}
