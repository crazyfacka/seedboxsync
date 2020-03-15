package modules

import (
	"fmt"
	"regexp"

	"github.com/crazyfacka/seedboxsync/domain"
	"golang.org/x/crypto/ssh"
)

func extractRar(conn *ssh.Client, content domain.Content, tempDir string) error {
	session, err := conn.NewSession()
	if err != nil {
		return err
	}

	cmd := "unrar e \"" + content.FullPath + "\" \"" + tempDir + "\""
	if err := session.Run(cmd); err != nil {
		return err
	}

	return nil
}

func transferData(seedbox *ssh.Client, player *ssh.Client, content domain.Content) error {

	return nil
}

// ProcessItems will extract what's to be extracted and copy what's to be copied
func ProcessItems(seedbox *ssh.Client, contents []domain.Content, tempDir string) error {
	rar := regexp.MustCompile(`(?i).*\.rar`)
	zip := regexp.MustCompile(`(?i).*\.zip`)

	for _, c := range contents {
		if c.IsDirectory {
			files, err := GetContentsFromHost(seedbox, c.FullPath)
			if err != nil {
				fmt.Printf("Error processing %s: %s\n", c.ItemName, err.Error())
			}

			for _, f := range files {
				if rar.MatchString(f.ItemName) {
					err = extractRar(seedbox, f, tempDir)

					if err != nil {
						fmt.Printf("Error processing %s: %s\n", c.ItemName, err.Error())
					}
					continue
				} else if zip.MatchString(f.ItemName) {
					fmt.Println("zip", f)
					continue
				}
			}
		}
	}

	return nil
}
