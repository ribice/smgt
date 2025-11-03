package main

import (
	"github.com/ribice/smgt/loopnow"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(loopnow.NewAnalyzer())
}
