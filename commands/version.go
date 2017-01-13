package commands

import (
	"fmt"
	"os"

	"github.com/pivotalservices/goblob"
)

func init() {
	Goblob.Version = func() {
		fmt.Println(goblob.Version)
		os.Exit(0)
	}
}
