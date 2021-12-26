package myerror

import (
	"fmt"
	"os"
)

func CheckError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "<\033[0;32;31m error: %s \033[0m>", err.Error())
	}
}

func CheckErrorExit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "<\033[0;32;31m error: %s \033[0m>", err.Error())
		os.Exit(1)
	}
}
