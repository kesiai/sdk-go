package main

import (
	"github.com/kesiai/sdk-go/driver"
	"github.com/kesiai/sdk-go/example/driver_lazy/app"
)

func main() {
	// 创建采集主程序
	d := new(app.TestDriver)
	driver.NewApp().Start(d)
}
