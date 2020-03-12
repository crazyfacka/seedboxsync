package modules

import (
	"fmt"
	"regexp"

	"github.com/crazyfacka/seedboxsync/domain"
	"golang.org/x/crypto/ssh"
)

// ProcessItems will extract what's to be extracted and copy what's to be copied
func ProcessItems(conn *ssh.Client, contents []domain.Content) error {
	rar := regexp.MustCompile(`(?i).*\.rar`)
	zip := regexp.MustCompile(`(?i).*\.zip`)

	for _, c := range contents {
		if c.IsDirectory {
			files, err := GetContentsFromHost(conn, c.FullPath)
			if err != nil {
				fmt.Printf("Error processing %s: %s\n", c.ItemName, err.Error())
			}

			for _, f := range files {
				if rar.MatchString(f.ItemName) {
					fmt.Println("rar", f)
				} else if zip.MatchString(f.ItemName) {
					fmt.Println("zip", f)
				}
			}
		}
	}

	return nil
}

// TransferData transfers data from the source to the destination
func TransferData(seedbox *ssh.Session, player *ssh.Session, contents []domain.Content) error {

	return nil
}
