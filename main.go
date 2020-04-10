package main

import (
	"flag"
	"os"
	"time"

	"github.com/crazyfacka/seedboxsync/domain"
	"github.com/crazyfacka/seedboxsync/handler"
	"github.com/crazyfacka/seedboxsync/modules"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName(".seedboxsync")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	dryrun := flag.Bool("dry", false, "doesn't transfer data from seedbox to player")
	debug := flag.Bool("debug", false, "sets log level to debug")
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	if err := viper.ReadInConfig(); err != nil {
		log.Error().Err(err).Msg("Error reading config file")
		os.Exit(-1)
	}

	log.Debug().Interface(".seedboxsync", viper.AllSettings()).Msg("Loaded configuration")

	seedboxConn, err := modules.GetSSHConn(viper.GetStringMap("seedbox"))
	if err != nil {
		log.Error().Err(err).Msg("Unable to setup seedbox session")
		return
	}

	playerConn, err := modules.GetSSHConn(viper.GetStringMap("player"))
	if err != nil {
		log.Error().Err(err).Msg("Unable to setup player session")
		return
	}

	db, err := modules.GetDB()
	if err != nil {
		log.Error().Err(err).Msg("Unable to open DB")
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
		log.Error().Err(err).Msg("Unable to get seedbox contents")
	}

	bundle.Contents = contents

	err = handler.FilterDownloadedContents(bundle)
	if err != nil {
		log.Error().Err(err).Msg("Error filtering contents")
	}

	err = handler.FillDestinationDirectories(bundle)
	if err != nil {
		log.Error().Err(err).Msg("Error finding destination contents")
	}

	err = handler.ProcessItems(bundle)
	if err != nil {
		log.Error().Err(err).Msg("Error processing contents")
	}

	modules.CloseDB()

	err = handler.RefreshLibrary(viper.GetStringMap("player")["host"].(string))
	if err != nil {
		log.Error().Err(err).Msg("Error refreshing library")
	}

	log.Info().Msg("Done")
}
