/*
Copyright © 2023 syu.m.5151@gmail.com
*/
package main

import (
	"fmt"
	"os"

	"github.com/nwiizo/aicommand/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
