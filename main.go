package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jibon57/secure-file-download-server/internal"
	"gopkg.in/yaml.v3"
)

const Version = "1.3.0"

func main() {
	cnfFile := "config.yaml"
	if len(os.Args[1:]) > 0 {
		cnfFile = os.Args[1:][0]
	}

	err := readYaml(cnfFile)
	if err != nil {
		log.Fatalln(err)
	}

	// create necessary dirs
	err = internal.CreateOrUpdateDirs()
	if err != nil {
		log.Fatalln(err)
	}
	// start scheduler
	go internal.StartScheduler()

	router := internal.Router(Version)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		sig := <-sigChan
		log.Println("exit requested, shutting down", "signal", sig)
		err = router.Shutdown()
		if err != nil {
			log.Fatalln(err)
		}
	}()

	err = router.Listen(fmt.Sprintf(":%d", internal.AppCnf.Port))
	if err != nil {
		log.Panicln(err)
	}
}

func readYaml(filename string) error {
	yamlFile, err := os.ReadFile(filename)

	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yamlFile, &internal.AppCnf)
	if err != nil {
		return err
	}

	return nil
}
