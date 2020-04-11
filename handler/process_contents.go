package handler

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/crazyfacka/seedboxsync/domain"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
)

func deleteWhatsComplete(conn *ssh.Client, files chan string, wg *sync.WaitGroup) {
	for file := range files {
		log.Info().Str("file", file).Msg("Deleting")

		session, err := conn.NewSession()
		if err != nil {
			log.Error().Err(err).Msg("Error creating new session")
			wg.Done()
			continue
		}

		cmd := "rm \"" + file + "\""
		if err := session.Run(cmd); err != nil {
			log.Error().Str("cmd", cmd).Err(err).Msg("Error executing")
		}

		wg.Done()
	}
}

func storeHashInDB(db *sql.DB, data chan string, wg *sync.WaitGroup) {
	stmt, err := db.Prepare("INSERT INTO downloaded(hash) VALUES(?)")
	if err != nil {
		log.Error().Err(err).Msg("Error creating statement")
		return
	}

	defer stmt.Close()

	for d := range data {
		log.Info().Str("data", d).Msg("Hashing")

		hash := md5.Sum([]byte(d))
		_, err = stmt.Exec(hex.EncodeToString(hash[:]))
		if err != nil {
			log.Error().Err(err).Msg("Error executing query")
			wg.Done()
			continue
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

	log.Info().Str("item", content.ItemName).Msg("Extracting")
	cmd := "unrar e -y \"" + content.FullPath + "\" \"" + tempDir + "\""
	if err := session.Run(cmd); err != nil {
		return err
	}

	return nil
}

func transferData(conn *ssh.Client, content domain.Content, tempDir string, filesToDelete chan string, toHash chan string, wg *sync.WaitGroup, dryrun bool) {
	defer wg.Done()

	for _, media := range content.MediaContent {
		session, err := conn.NewSession()
		if err != nil {
			log.Error().Err(err).Msg("Error creating new session")
			return
		}

		cmd := "mkdir -p \"" + content.DestinationPath + "\" ; scp -rP 2211 crazyfacka@joagonca.com:\"" + tempDir + "/" + media + "\" \"" + content.DestinationPath + "\""

		if dryrun {
			log.Info().Str("item", content.ItemName).Msg("[DRY] Copying")
			log.Info().Str("cmd", cmd).Msg("[DRY]")
			time.Sleep(5 * time.Second)
			log.Info().Str("item", content.ItemName).Msg("[DRY] Copying complete")

			wg.Add(1)
			toHash <- content.FullPath
			if filesToDelete != nil {
				wg.Add(1)
				filesToDelete <- tempDir + "/" + media
			}
		} else {
			if err := session.Run(cmd); err != nil {
				log.Error().Str("cmd", cmd).Err(err).Msg("Error executing")
			} else {
				log.Info().Str("item", content.ItemName).Msg("Copying complete")
				wg.Add(1)
				toHash <- content.FullPath

				if filesToDelete != nil {
					wg.Add(1)
					filesToDelete <- tempDir + "/" + media
				}
			}
		}
	}
}

// ProcessItems will extract what's to be extracted and copy what's to be copied
func ProcessItems(b *domain.Bundle) error {
	var wg sync.WaitGroup

	rar := regexp.MustCompile(`(?i).*\.rar`)
	mkv := regexp.MustCompile(`(?i).*\.mkv`)
	zip := regexp.MustCompile(`(?i).*\.zip`)

	filesToDelete := make(chan string, 2)
	toHash := make(chan string, 2)

	if len(b.Contents) > 0 {
		go deleteWhatsComplete(b.Seedbox, filesToDelete, &wg)
		go storeHashInDB(b.DB, toHash, &wg)
	}

	for _, c := range b.Contents {
		if c.IsDirectory {
			files, err := GetContentsFromHost(b.Seedbox, c.FullPath)
			if err != nil {
				log.Error().Str("item", c.ItemName).Err(err).Msg("Error processing")
			}

			for _, f := range files {
				if rar.MatchString(f.ItemName) {
					err = extractRar(b.Seedbox, &f, b.TempDir)
					if err != nil {
						log.Error().Str("item", c.ItemName).Err(err).Msg("Error processing")
						continue
					}

					wg.Add(1)
					c.MediaContent = f.MediaContent
					go transferData(b.Player, c, b.TempDir, filesToDelete, toHash, &wg, b.DryRun)
				} else if mkv.MatchString(f.ItemName) {
					wg.Add(1)
					c.MediaContent = []string{f.ItemName}
					go transferData(b.Player, c, c.FullPath[:len(c.FullPath)-1], nil, toHash, &wg, b.DryRun)
				} else if zip.MatchString(f.ItemName) {
					log.Warn().Str("item", f.ItemName).Msg("ZIP handling not implemented")
					continue
				}
			}
		}
	}

	wg.Wait()
	close(filesToDelete)
	close(toHash)

	return nil
}
