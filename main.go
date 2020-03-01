package main

import (
	"encoding/json"
	"fmt"
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

	seedBoxSession, err := modules.GetSSHSession(viper.GetStringMap("seedbox"))
	if err != nil {
		fmt.Printf("Unable to setup seedbox session: %s\n", err.Error())
	}

	playerSession, err := modules.GetSSHSession(viper.GetStringMap("player"))
	if err != nil {
		fmt.Printf("Unable to setup player session: %v", err)
	}

	db, err := modules.GetDB()
	if err != nil {
		fmt.Printf("Unable to open DB: %s\n", err.Error())
	}

	contents, err := modules.GetContentsFromHost(seedBoxSession, viper.GetStringMap("seedbox")["dir"].(string))
	if err != nil {
		fmt.Printf("Unable to get seedbox contents: %s\n", err.Error())
	}

	filtered, err := modules.FilterDownloadedContents(contents, db)
	if err != nil {
		fmt.Printf("Error filtering contents: %s\n", err.Error())
	}

	filtered, err = modules.FillDestinationDirectories(playerSession, viper.GetStringMap("player")["dir"].(string), filtered)
	if err != nil {
		fmt.Printf("Error finding destionation contents: %s\n", err.Error())
	}

	fmt.Println(filtered)

	modules.CloseDB()
}
