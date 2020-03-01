package modules

import (
	"bytes"
	"database/sql"
	"strings"

	"github.com/crazyfacka/seedboxsync/domain"
	"golang.org/x/crypto/ssh"
)

// FilterDownloadedContents filters the array and leaves only the contents to download
func FilterDownloadedContents(contents *[]domain.Content, db *sql.DB) {

}

// GetContentsFromSeedbox parses LS output to produce a curated list of contents
func GetContentsFromSeedbox(session *ssh.Session, dir string) ([]domain.Content, error) {
	var output bytes.Buffer
	var contents []domain.Content

	var isDir bool

	if dir[len(dir)-1:] != "/" {
		dir += "/"
	}

	session.Stdout = &output
	if err := session.Run("ls -1F \"" + dir + "\""); err != nil {
		return nil, err
	}

	items := strings.Split(strings.TrimSpace(output.String()), "\n")
	for _, item := range items {
		isDir = false
		if item[len(item)-1:] == "/" {
			isDir = true
		}

		contents = append(contents, domain.Content{
			IsDirectory: isDir,
			FullPath:    dir + item,
		})
	}

	return contents, nil
}
