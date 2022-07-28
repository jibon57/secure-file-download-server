package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var AppCnf AppConfig

func main() {
	cnfFile := "config.yaml"
	if len(os.Args[1:]) > 0 {
		cnfFile = os.Args[1:][0]
	}

	err := readYaml(cnfFile)
	if err != nil {
		log.Panicln(err)
	}

	router := Router()

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

	err = router.Listen(fmt.Sprintf(":%d", AppCnf.Port))
	if err != nil {
		log.Panicln(err)
	}
}

func readYaml(filename string) error {
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		return err
	}

	err = yaml.Unmarshal(yamlFile, &AppCnf)
	if err != nil {
		return err
	}

	return nil
}
