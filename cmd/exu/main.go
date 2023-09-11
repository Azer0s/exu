package main

import (
	"exu/client"
	"exu/server"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
)

func printUsage() {
	fmt.Println("Usage: exu <command>")
	fmt.Println("Commands:")
	fmt.Println("  server <config-file>")
	fmt.Println("  connect <config-file>")
}

func main() {
	log.SetLevel(log.DebugLevel)

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "server":
		/*if len(os.Args) < 3 {
			printUsage()
			os.Exit(1)
		}*/

		srv := server.New(server.ModeVSwitch)
		srv.Run()
	case "connect":
		/*if len(os.Args) < 3 {
			printUsage()
			os.Exit(1)
		}*/

		cl := client.New()
		cl.Run()
	default:
		printUsage()
		os.Exit(1)
	}
}
