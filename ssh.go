package bastion

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

// NewEncryptedSSHKey returns a SSH key and encrypts it with the given key
func NewSSHKey() []byte {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		panic(err)
	}
	return x509.MarshalPKCS1PrivateKey(privateKey)
}

// SignCert signs a public key based on a signer and
// marshals the output to be stored as an authorized key file
// that can be added to ssh-agent
func SignCert(pubKeyBytes []byte, signer ssh.Signer) ([]byte, error) {
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(pubKeyBytes)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse Public Key")
	}
	cert := &ssh.Certificate{
		Serial:          1,
		ValidPrincipals: []string{"user"},
		SignatureKey:    signer.PublicKey(),
		KeyId:           "user",
		Key:             pubKey,
		CertType:        ssh.UserCert,
		ValidAfter:      uint64(time.Now().Add(time.Hour * 24 * 7).Unix()),
		Permissions: ssh.Permissions{
			Extensions: map[string]string{
				"permit-X11-forwarding":   "",
				"permit-agent-forwarding": "",
				"permit-port-forwarding":  "",
				"permit-pty":              "",
				"permit-user-rc":          "",
			},
		},
	}
	cert.SignCert(rand.Reader, signer)
	return ssh.MarshalAuthorizedKey(cert), nil
}
