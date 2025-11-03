package main

import (
	"github.com/ribice/smgc/loopnow"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(loopnow.NewAnalyzer())
}
