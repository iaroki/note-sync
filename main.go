package main

import (
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ProtonMail/gopenpgp/v2/helper"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/mitchellh/go-homedir"
)

func encryptData(data []byte, publicKey []byte) (string, error) {
	armor, err := helper.EncryptMessageArmored(string(publicKey), string(data))

	if err != nil {
		log.Fatalln("Encrypt error", err)
		return "", err
	}

	return armor, err

}

func decryptData(data []byte, passphrase []byte, privateKey []byte) (string, error) {
	if len(passphrase) == 0 {
		passphrase = nil
	}
	descryptedData, err := helper.DecryptMessageArmored(string(privateKey), passphrase, string(data))

	if err != nil {
		log.Fatalln("Decrypt error", err)
		return "", err
	}

	return descryptedData, err

}

func getNotes(rootPath string) []string {

	notesList := []string{}

	err := filepath.Walk(rootPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println(err)
			return err
		}
		if !info.IsDir() && strings.Contains(filePath, ".md") {
			notesList = append(notesList, filePath)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	return notesList
}

func writeNote(path string, data string) {
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		log.Fatalln("Can't create directories", err)
	}
	err = ioutil.WriteFile(path, []byte(data), fs.FileMode(0644))
	if err != nil {
		log.Fatalln("Can't write file", err)
	}
	log.Println("Written:", path)
}

func pushNotes(notesPath string, encPath string, publicKey []byte) {
	notesList := getNotes(notesPath)
	for _, noteFile := range notesList {
		log.Println("Processing:", noteFile)
		data, err := ioutil.ReadFile(noteFile)

		if err != nil {
			log.Fatalln("Can't open note file", err)
		}

		encFile := strings.Replace(noteFile, notesPath, encPath, 1) + ".gpg"
		encData, err := encryptData(data, publicKey)

		if err != nil {
			log.Fatalln("Can't encrypt data", err)
		}

		writeNote(encFile, encData)
	}
}

func pullNotes(notesPath string, encPath string, passphrase string, privateKey []byte) {
	notesList := getNotes(encPath)
	for _, noteFile := range notesList {
		log.Println("Processing:", noteFile)
		data, err := ioutil.ReadFile(noteFile)

		if err != nil {
			log.Fatalln("Can't open note file", err)
		}

		decFile := strings.TrimRight(strings.Replace(noteFile, encPath, notesPath, 1), ".gpg")
		decData, err := decryptData(data, []byte(passphrase), privateKey)

		if err != nil {
			log.Fatalln("Can't decrypt data", err)
		}

		writeNote(decFile, decData)
	}

}

func pullGit(repoPath string, privateKeyPath string) {

	publicKeys, err := ssh.NewPublicKeysFromFile("git", privateKeyPath, "")

	if err != nil {
		log.Fatalln("Can't init ssh key", err)
	}

	log.Println("=> Pulling the repo: ", repoPath)
	r, err := git.PlainOpen(repoPath)

	if err != nil {
		log.Fatalln("Can't open repo", err)
	}

	w, err := r.Worktree()

	if err != nil {
		log.Fatalln("Can't init worktree", err)
	}

	err = w.Pull(&git.PullOptions{
		RemoteName: "origin",
		Auth:       publicKeys,
	})

	if err != nil {
		log.Println("==> ", err)
	}
}

func pushGit(repoPath string, encDir string, privateKeyPath string) {

	publicKeys, err := ssh.NewPublicKeysFromFile("git", privateKeyPath, "")

	if err != nil {
		log.Fatalln("Can't init ssh key", err)
	}

	log.Println("=> Pushing the repo: ", repoPath)
	r, err := git.PlainOpen(repoPath)

	if err != nil {
		log.Fatalln("Can't open repo", err)
	}

	w, err := r.Worktree()

	if err != nil {
		log.Fatalln("Can't init worktree", err)
	}

	_, err = w.Add(encDir)

	if err != nil {
		log.Fatalln("Can't add files", err)
	}

	_, err = w.Commit("Auto update", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "iaroki",
			Email: "iaroki@protonmail.com",
			When:  time.Now(),
		},
	})

	if err != nil {
		log.Fatalln("Error committing files", err)
	}

	err = r.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth:       publicKeys,
	})

	if err != nil {
		log.Println("==> ", err)
	}
}

func gracefulShutdown() {
	appVersion := "v0.2"
	appCopy := "(c)iaroki, 2023"
	log.Printf("note-sync: %s, %s", appVersion, appCopy)
	log.Println("Please use [pull] or [push] action.")
	os.Exit(0)
}

func main() {

	var action string
	if len(os.Args) == 2 {
		action = os.Args[1]
	} else {
		gracefulShutdown()
	}

	configPath := findConfig()
	appConfig := getConfig(configPath)

	homeDir, _ := homedir.Dir()
	notesDir := appConfig.NotesDir
	gitDir := appConfig.GitDir
	encDir := appConfig.EncDir
	notesPath := filepath.Join(homeDir, notesDir)
	encPath := filepath.Join(gitDir, encDir)

	if action == "pull" {

		pullGit(gitDir, appConfig.SSHPrivateKey)

		privateKeyPath := appConfig.GPGPrivateKey
		privateKey, err := ioutil.ReadFile(privateKeyPath)

		if err != nil {
			log.Fatalln("Can't open private key file", err)
		}

		pullNotes(notesPath, encPath, "", privateKey)

	} else if action == "push" {

		publicKeyPath := appConfig.GPGPublicKey
		publicKey, err := ioutil.ReadFile(publicKeyPath)

		if err != nil {
			log.Fatalln("Can't open public key file", err)
		}

		pushNotes(notesPath, encPath, publicKey)
		pushGit(gitDir, encDir, appConfig.SSHPrivateKey)

	} else {

		gracefulShutdown()
	}

}
