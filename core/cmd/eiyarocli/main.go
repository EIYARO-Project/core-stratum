package main

import (
	"runtime"

	cmd "corepool/core/cmd/eiyarocli/commands"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	cmd.Execute()
}
