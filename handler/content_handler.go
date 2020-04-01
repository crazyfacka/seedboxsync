package handler

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/crazyfacka/seedboxsync/domain"
	"golang.org/x/crypto/ssh"
)

// FilterDownloadedContents filters the array and leaves only the contents to download
func FilterDownloadedContents(b *domain.Bundle) error {
	var res int
	var filtered []domain.Content

	re := regexp.MustCompile(`(?i)s[0-9]+e[0-9]+`)

	stmt, err := b.DB.Prepare("SELECT COUNT(1) FROM downloaded WHERE hash = ?")
	if err != nil {
		return err
	}

	defer stmt.Close()

	for _, item := range b.Contents {
		if re.MatchString(item.ItemName) {
			hash := md5.Sum([]byte(item.FullPath))
			err := stmt.QueryRow(hex.EncodeToString(hash[:])).Scan(&res)
			if err != nil {
				return err
			}

			if res == 0 {
				fmt.Printf("Adding '%s' to queue\n", item.ItemName)
				filtered = append(filtered, item)
			}
		}
	}

	b.Contents = filtered
	return nil
}

// GetContentsFromHost parses LS output to produce a curated list of contents
func GetContentsFromHost(conn *ssh.Client, dir string) ([]domain.Content, error) {
	var output bytes.Buffer
	var contents []domain.Content

	var isDir bool
	var itemName string

	if dir[len(dir)-1:] != "/" {
		dir += "/"
	}

	session, err := conn.NewSession()
	if err != nil {
		return nil, err
	}

	session.Stdout = &output
	if err := session.Run("ls -1F \"" + dir + "\""); err != nil {
		return nil, err
	}

	items := strings.Split(strings.TrimSpace(output.String()), "\n")
	for _, item := range items {
		if item[len(item)-1:] == "/" {
			isDir = true
			itemName = item[:len(item)-1]
		} else {
			isDir = false
			itemName = item
		}

		contents = append(contents, domain.Content{
			IsDirectory: isDir,
			ItemName:    itemName,
			FullPath:    dir + item,
		})
	}

	return contents, nil
}
