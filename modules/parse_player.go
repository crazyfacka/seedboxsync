package modules

import (
	"regexp"
	"strings"

	"github.com/crazyfacka/seedboxsync/domain"
	"golang.org/x/crypto/ssh"

	"github.com/texttheater/golang-levenshtein/levenshtein"
)

// FillDestinationDirectories finds the directories in the destination that match the sources
func FillDestinationDirectories(conn *ssh.Client, dir string, source []domain.Content) ([]domain.Content, error) {
	contents, err := GetContentsFromHost(conn, dir)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`(?i)s[0-9]+e[0-9]+`)
	reSeason := regexp.MustCompile(`(?i)s[0-9]+`)

	for i, s := range source {
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

		source[i].DestinationPath = curDir
	}

	return source, nil
}
