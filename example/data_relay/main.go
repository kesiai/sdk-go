package main

import (
	"context"
	"github.com/kesiai/sdk-go/data_relay"
	"github.com/kesiai/sdk-go/example/data_relay/app"
)

func main() {
	d := new(app.TestRelay)
	d.Ctx, d.Cancel = context.WithCancel(context.Background())
	data_relay.NewApp().Start(d)
}
