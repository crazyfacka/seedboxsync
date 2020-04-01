package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/crazyfacka/seedboxsync/domain"

	"github.com/texttheater/golang-levenshtein/levenshtein"
)

// RefreshLibrary refreshes the player library
func RefreshLibrary(host string) error {
	reqBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "VideoLibrary.Scan",
		"id":      1,
		"params": map[string]bool{
			"showdialogs": true,
		},
	})

	if err != nil {
		return err
	}

	fmt.Println("Refreshing player's library...")
	resp, err := http.Post("http://"+host+":8080/jsonrpc", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}

// FillDestinationDirectories finds the directories in the destination that match the sources
func FillDestinationDirectories(b *domain.Bundle) error {
	contents, err := GetContentsFromHost(b.Player, b.PlayerDir)
	if err != nil {
		return err
	}

	re := regexp.MustCompile(`(?i)s[0-9]+e[0-9]+`)
	reSeason := regexp.MustCompile(`(?i)s[0-9]+`)

	for i, s := range b.Contents {
		trimLocation := re.FindIndex([]byte(s.ItemName))
		toMatch := strings.TrimSpace(strings.ReplaceAll(s.ItemName[:trimLocation[0]], ".", " "))
		season := reSeason.FindString(s.ItemName)[1:]

		curDistance := 100
		curDir := ""

		for _, c := range contents {
			distance := levenshtein.DistanceForStrings([]rune(toMatch), []rune(c.ItemName), levenshtein.DefaultOptions)
			if distance < curDistance {
				curDistance = distance
				curDir = c.FullPath + "Season " + season + "/"
			}
		}

		b.Contents[i].DestinationPath = curDir
	}

	return nil
}
