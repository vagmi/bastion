package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/vagmi/bastion"
)

func main() {
	ks, err := bastion.NewKmsCryptService()
	if err != nil {
		panic(err)
	}
	sshKey := bastion.NewSSHKey()

	ciphered, err := ks.EncryptAsSSHKey(sshKey)
	if err != nil {
		panic(err)
	}

	signer, err := ks.DecryptAsSSHKey(ciphered)
	if err != nil {
		panic(err)
	}

	file, _ := os.Open("/Users/vagmi/.ssh/id_rsa.pub")
	defer file.Close()
	pubKey, _ := ioutil.ReadAll(file)
	cert, err := bastion.SignCert(pubKey, signer)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(cert))
}
