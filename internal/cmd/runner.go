package cmd

import (
	cmdRunner "github.com/berquerant/git-iter-go/internal/cmd/runner"
)

// Aliases for runner functions moved to internal/cmd/runner.
var (
	RunDo      = cmdRunner.RunDo
	RunRead    = cmdRunner.RunRead
	RunList    = cmdRunner.RunList
	RunGrep    = cmdRunner.RunGrep
	RunPull    = cmdRunner.RunPull
	RunFetch   = cmdRunner.RunFetch
	RunLsFiles = cmdRunner.RunLsFiles
	RunStatus  = cmdRunner.RunStatus
	RunDiff    = cmdRunner.RunDiff
)
