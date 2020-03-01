package modules

import (
	"io/ioutil"
	"path/filepath"
	"strconv"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	"github.com/mitchellh/go-homedir"
)

func publicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}

	return ssh.PublicKeys(key)
}

// GetSSHSession returns a valid SSH session
func GetSSHSession(confs map[string]interface{}) (*ssh.Session, error) {
	host := confs["host"].(string)
	port := int(confs["port"].(float64))
	user := confs["user"].(string)
	key := confs["key"].(string)

	homeDirectory, err := homedir.Dir()
	if err != nil {
		return nil, err
	}

	hostKeyCallback, err := knownhosts.New(filepath.Join(homeDirectory, ".ssh", "known_hosts"))
	if err != nil {
		return nil, err
	}

	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			publicKeyFile(key),
		},
		HostKeyCallback: hostKeyCallback,
	}

	conn, err := ssh.Dial("tcp", host+":"+strconv.Itoa(port), sshConfig)
	if err != nil {
		return nil, err
	}

	session, err := conn.NewSession()
	if err != nil {
		return nil, err
	}

	return session, nil
}
