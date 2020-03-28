package modules

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/crazyfacka/seedboxsync/domain"
	"golang.org/x/crypto/ssh"
)

func deleteWhatsComplete(conn *ssh.Client, files chan string, wg *sync.WaitGroup) {
	for file := range files {
		wg.Add(1)
		fmt.Printf("Deleting '%s'\n", file)

		session, err := conn.NewSession()
		if err != nil {
			fmt.Printf("Error creating new session: %s\n", err.Error())
			wg.Done()
			continue
		}

		cmd := "rm \"" + file + "\""
		if err := session.Run(cmd); err != nil {
			fmt.Printf("Error executing '%s': %s\n", cmd, err.Error())
		}

		wg.Done()
	}
}

func getRarMedia(conn *ssh.Client, content *domain.Content) error {
	var output bytes.Buffer

	session, err := conn.NewSession()
	if err != nil {
		return err
	}

	session.Stdout = &output
	cmd := "unrar lb \"" + content.FullPath + "\""
	if err := session.Run(cmd); err != nil {
		return err
	}

	re := regexp.MustCompile(`(?i).*\.(mp4|mkv)`)
	items := strings.Split(strings.TrimSpace(output.String()), "\n")
	for _, item := range items {
		if re.MatchString(item) {
			content.MediaContent = append(content.MediaContent, item)
		}
	}

	return nil
}

func extractRar(conn *ssh.Client, content *domain.Content, tempDir string) error {
	err := getRarMedia(conn, content)
	if err != nil {
		return err
	}

	session, err := conn.NewSession()
	if err != nil {
		return err
	}

	fmt.Printf("Extracting '%s'\n", content.ItemName)
	cmd := "unrar e \"" + content.FullPath + "\" \"" + tempDir + "\""
	if err := session.Run(cmd); err != nil {
		return err
	}

	return nil
}

func transferData(conn *ssh.Client, content domain.Content, tempDir string, filesToDelete chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for _, media := range content.MediaContent {
		session, err := conn.NewSession()
		if err != nil {
			fmt.Printf("Error creating new session: %s\n", err.Error())
			return
		}

		fmt.Printf("Copying '%s'...\n", content.ItemName)
		cmd := "mkdir -p \"" + content.DestinationPath + "\" ; scp -rP 2211 crazyfacka@joagonca.com:\"" + tempDir + "/" + media + "\" \"" + content.DestinationPath + "\""
		if err := session.Run(cmd); err != nil {
			fmt.Printf("Error executing '%s': %s\n", cmd, err.Error())
		} else {
			fmt.Printf("Copying %s complete\n", content.ItemName)
			// TODO Store hash in DB
			filesToDelete <- tempDir + "/" + media
		}
	}
}

// ProcessItems will extract what's to be extracted and copy what's to be copied
func ProcessItems(seedbox *ssh.Client, player *ssh.Client, contents []domain.Content, tempDir string) error {
	var wg sync.WaitGroup

	rar := regexp.MustCompile(`(?i).*\.rar`)
	zip := regexp.MustCompile(`(?i).*\.zip`)

	filesToDelete := make(chan string)
	go deleteWhatsComplete(seedbox, filesToDelete, &wg)

	for _, c := range contents {
		if c.IsDirectory {
			files, err := GetContentsFromHost(seedbox, c.FullPath)
			if err != nil {
				fmt.Printf("Error processing %s: %s\n", c.ItemName, err.Error())
			}

			for _, f := range files {
				if rar.MatchString(f.ItemName) {
					err = extractRar(seedbox, &f, tempDir)
					if err != nil {
						fmt.Printf("Error processing %s: %s\n", c.ItemName, err.Error())
						continue
					}

					wg.Add(1)
					c.MediaContent = f.MediaContent
					go transferData(player, c, tempDir, filesToDelete, &wg)
				} else if zip.MatchString(f.ItemName) {
					fmt.Println("zip", f)
					continue
				}
			}
		}

		wg.Wait()
		close(filesToDelete)
	}

	return nil
}
