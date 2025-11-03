package main

import (
	"github.com/ribice/smgc/set"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(set.NewAnalyzer())
}
