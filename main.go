package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/crazyfacka/seedboxsync/domain"
	"github.com/crazyfacka/seedboxsync/handler"
	"github.com/crazyfacka/seedboxsync/modules"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName(".seedboxsync")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	dryrun := flag.Bool("dry", false, "doesn't transfer data from seedbox to player")
	flag.Parse()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file: %s\n", err.Error())
		os.Exit(-1)
	}

	fmt.Println("== Configuration ==")
	b, _ := json.MarshalIndent(viper.AllSettings(), "", "  ")
	fmt.Println(string(b))

	seedboxConn, err := modules.GetSSHConn(viper.GetStringMap("seedbox"))
	if err != nil {
		fmt.Printf("Unable to setup seedbox session: %s\n", err.Error())
		return
	}

	playerConn, err := modules.GetSSHConn(viper.GetStringMap("player"))
	if err != nil {
		fmt.Printf("Unable to setup player session: %v", err)
		return
	}

	db, err := modules.GetDB()
	if err != nil {
		fmt.Printf("Unable to open DB: %s\n", err.Error())
		return
	}

	bundle := &domain.Bundle{
		Seedbox:    seedboxConn,
		Player:     playerConn,
		DB:         db,
		SeedboxDir: viper.GetStringMap("seedbox")["dir"].(string),
		PlayerDir:  viper.GetStringMap("player")["dir"].(string),
		TempDir:    viper.GetStringMap("seedbox")["temp_dir"].(string),
		DryRun:     *dryrun,
	}

	contents, err := handler.GetContentsFromHost(bundle.Seedbox, bundle.SeedboxDir)
	if err != nil {
		fmt.Printf("Unable to get seedbox contents: %s\n", err.Error())
	}

	bundle.Contents = contents

	fmt.Println("== Contents ==")
	err = handler.FilterDownloadedContents(bundle)
	if err != nil {
		fmt.Printf("Error filtering contents: %s\n", err.Error())
	}

	err = handler.FillDestinationDirectories(bundle)
	if err != nil {
		fmt.Printf("Error finding destionation contents: %s\n", err.Error())
	}

	err = handler.ProcessItems(bundle)
	if err != nil {
		fmt.Printf("Error processing contents: %s\n", err.Error())
	}

	modules.CloseDB()

	err = handler.RefreshLibrary(viper.GetStringMap("player")["host"].(string))
	if err != nil {
		fmt.Printf("Error refreshing library: %s\n", err.Error())
	}

	fmt.Println("Done")
}
