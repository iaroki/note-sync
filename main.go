package main

import (
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ProtonMail/gopenpgp/v2/helper"
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

func getNotes(path string) []string {

  notesList := []string{}

  err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
    if err != nil {
            log.Println(err)
            return err
        }
        if !info.IsDir() {
          notesList = append(notesList, filepath.Base(path))
        }
        return nil
    })
    if err != nil {
        log.Fatal(err)
    }

  return notesList
}

func writeNote(path string, data string) {
  err := ioutil.WriteFile(path, []byte(data), fs.FileMode(0644))
  if err != nil {
    log.Fatalln("Cant write file", err)
  }
}

func pushNotes(notesPath string, encPath string, publicKey []byte) {
  notesList := getNotes(notesPath)
    for _, file := range(notesList) {
      log.Println("Processing ", file)
      data, err := ioutil.ReadFile(filepath.Join(notesPath, file))

      if err != nil {
        log.Fatalln("Cant open note file", err)
      }

      encFile := filepath.Join(encPath, file) + ".gpg"
      encData, err := encryptData(data, publicKey)

      if err != nil {
        log.Fatalln("Cant encrypt data", err)
      }

      writeNote(encFile, encData)
    }
}

func pullNotes(notesPath string, encPath string, passphrase string, privateKey []byte){
  notesList := getNotes(encPath)
    for _, file := range(notesList) {
      log.Println("Processing ", file)
      data, err := ioutil.ReadFile(filepath.Join(encPath, file))

      if err != nil {
        log.Fatalln("Cant open note file", err)
      }

      decFile := filepath.Join(notesPath, strings.TrimRight(file, ".gpg"))
      decData, err := decryptData(data, []byte(passphrase), privateKey)

      if err != nil {
        log.Fatalln("Cant decrypt data", err)
      }

      writeNote(decFile, decData)
    }

}

func gracefulShutdown() {
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

  homeDir, _ := homedir.Dir()
  notesDir := "zettelkasten"
  encDir := "enc"
  notesPath := filepath.Join(homeDir, notesDir)
  encPath := filepath.Join(homeDir, encDir)

  if action == "pull" {
    privateKeyPath:= filepath.Join(homeDir, ".gnupg/private.gpg")
    privateKey, err := ioutil.ReadFile(privateKeyPath)

    if err != nil {
      log.Fatalln("Cant open private key file", err)
    }

    pullNotes(notesPath, encPath, "", privateKey)
  } else if action == "push" {
    publicKeyPath:= filepath.Join(homeDir, ".gnupg/public.gpg")
    publicKey, err := ioutil.ReadFile(publicKeyPath)

    if err != nil {
      log.Fatalln("Cant open public key file", err)
    }

    pushNotes(notesPath, encPath, publicKey)
  } else {
    gracefulShutdown()
  }

}
