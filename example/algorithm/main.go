package main

import (
	"context"
	"github.com/kesiai/sdk-go/algorithm"
	"github.com/kesiai/sdk-go/example/algorithm/app"
)

func main() {
	s := new(app.TestAlgorithm)
	s.Ctx, s.Cancel = context.WithCancel(context.Background())
	algorithm.NewApp().Start(s)
}
