package main

import (
	"github.com/Warashi/gid"

	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { unitchecker.Main(gid.Analyzer) }
