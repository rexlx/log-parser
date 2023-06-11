package main

import (
	"fmt"
	"log"
	"os"
	"os/user"

	"golang.org/x/crypto/ssh"
)

type SshConfig struct {
	Address string
	User    string
	Port    int
}

func NewSshClient(addr string, uname string, port int) *ssh.Client {
	// var hostKey ssh.PublicKey
	if uname == "" {
		me, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		uname = me.Name
	}

	keyDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(fmt.Sprintf("%v/.ssh", keyDir)); os.IsNotExist(err) {
		log.Fatal("no key directory found at ~/.ssh/ !")
	}
	fmt.Println(addr, port, uname, keyDir)
	client, err := ssh.Dial("tcp", fmt.Sprintf("%v:%v", addr, port), &ssh.ClientConfig{
		User: uname,
		Auth: []ssh.AuthMethod{
			loadKey(getKey(fmt.Sprintf("%v/.ssh", keyDir))),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func getKey(path string) string {
	stuff, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	for _, i := range stuff {
		if i.Name() == "id_ecdsa" {
			return fmt.Sprintf("%v/id_ecdsa", path)
		} else if i.Name() == "id_rsa" {
			return fmt.Sprintf("%v/id_rsa", path)
		}
	}
	return ""
}

func loadKey(path string) ssh.AuthMethod {
	// contents is a bytes array
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	key, err := ssh.ParsePrivateKey(contents)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}
