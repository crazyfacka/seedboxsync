package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/crazyfacka/seedboxsync/modules"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName(".seedboxsync")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file: %s\n", err.Error())
		os.Exit(-1)
	}

	fmt.Println("== Configuration ==")
	b, _ := json.MarshalIndent(viper.AllSettings(), "", "  ")
	fmt.Println(string(b))

	seedboxConfs := viper.Get("seedbox").(map[string]interface{})
	playerConfs := viper.Get("player").(map[string]interface{})

	seedBoxSession, err := modules.GetSSHSession(seedboxConfs)
	if err != nil {
		fmt.Printf("Unable to setup seedbox session: %v", err)
	}

	playerSession, err := modules.GetSSHSession(playerConfs)
	if err != nil {
		fmt.Printf("Unable to setup player session: %v", err)
	}

	var output bytes.Buffer
	seedBoxSession.Stdout = &output
	playerSession.Stdout = &output

	if err := seedBoxSession.Run("ls"); err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}
	fmt.Println(output.String())

	if err := playerSession.Run("ls"); err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}
	fmt.Println(output.String())
}
