package main

import (
	"fmt"

	"monitor/base"
)

var b = base.NewBase()

func main() {
	resutlt := b.TranslateToK("1.85G")

	fmt.Println(resutlt)
}
