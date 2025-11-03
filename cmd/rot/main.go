package main

import (
	"github.com/ribice/smgt/rot"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(rot.NewAnalyzer())
}
