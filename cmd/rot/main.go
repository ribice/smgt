package main

import (
	"github.com/ribice/smgc/rot"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(rot.NewAnalyzer())
}
