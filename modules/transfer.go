package modules

import (
	"github.com/crazyfacka/seedboxsync/domain"
	"golang.org/x/crypto/ssh"
)

// TransferData transfers data from the source to the destination
func TransferData(seedbox *ssh.Session, player *ssh.Session, contents []domain.Content) error {

	return nil
}
