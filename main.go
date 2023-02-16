package main

import (
	"tvbox_jx/js"
	"tvbox_jx/tools"

	"github.com/robertkrimen/otto"
)

//var
var (
	keyword    string
	result     string
	page       int
	vm         *otto.Otto
	sourcePath string
	startTime  int64
)

func main() {
	SID := "111"
	keyword = "我的"
	sourcePath = "./source.json"
	vm = js.Init(vm)
	tools.Spider(vm, SID, keyword, sourcePath)

}
