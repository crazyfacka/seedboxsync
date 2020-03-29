package domain

import (
	"database/sql"

	"golang.org/x/crypto/ssh"
)

// Bundle represents the state of data for the application to work
type Bundle struct {
	Seedbox *ssh.Client
	Player  *ssh.Client
	DB      *sql.DB

	SeedboxDir string
	PlayerDir  string
	TempDir    string

	Contents []Content

	DryRun bool
}
