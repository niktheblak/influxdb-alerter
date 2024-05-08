package main

import (
	"fmt"
	"os"

	"github.com/niktheblak/influxdb-alerter/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		if _, err := fmt.Fprintf(os.Stderr, "%s\n", err); err != nil {
			panic(err)
		}
		os.Exit(1)
	}
}
