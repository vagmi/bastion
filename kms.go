package bastion

import (
	"bytes"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/gob"
	"encoding/pem"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

type EncryptedData struct {
	CipherKey     string
	CipherContent string
}

type CryptService interface {
	EncryptAsSSHKey(x509Data []byte) ([]byte, error)
	DecryptAsSSHKey(x509Data []byte) (ssh.Signer, error)
}

type KmsCryptService struct {
	kmsService *kms.KMS
}

func NewKmsCryptService() (*KmsCryptService, error) {
	kmsSvc, err := getKmsSvc()
	if err != nil {
		return nil, err
	}
	return &KmsCryptService{
		kmsService: kmsSvc,
	}, nil
}

func getKmsSvc() (*kms.KMS, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	kmsSvc := kms.New(sess)
	return kmsSvc, nil
}

// CreateMasterKey creates a master key for
// encrypting the CA private key
func (cs *KmsCryptService) CreateMasterKey() (*string, error) {
	res, err := cs.kmsService.CreateKey(&kms.CreateKeyInput{
		Tags: []*kms.Tag{
			{
				TagKey:   aws.String("CreatedBy"),
				TagValue: aws.String("Bastion"),
			},
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create key")
	}
	keyName := "alias/bastion/mainKey"
	_, err = cs.kmsService.CreateAlias(&kms.CreateAliasInput{
		TargetKeyId: res.KeyMetadata.KeyId,
		AliasName:   &keyName,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create an alias for key %s", res.KeyMetadata.KeyId)
	}
	return res.KeyMetadata.KeyId, nil
}

// GetMasterKey returns a key if exists or
// creates one if it doesnt exist
func (cs *KmsCryptService) GetMasterKey() (*string, error) {
	keyName := "alias/bastion/mainKey"
	keyOutput, err := cs.kmsService.DescribeKey(&kms.DescribeKeyInput{
		KeyId: &keyName,
	})
	if err != nil {
		switch err.(type) {
		case *kms.NotFoundException:
			return cs.CreateMasterKey()
		default:
			return nil, err
		}
	}
	return keyOutput.KeyMetadata.KeyId, nil
}

// EncryptAsSSHKey encrypts given x509 data as a
// encrypted PEM block
func (cs *KmsCryptService) EncryptAsSSHKey(x509Bytes []byte) ([]byte, error) {
	k, err := cs.GetMasterKey()
	if err != nil {
		return nil, err
	}
	keySpec := "AES_256"
	dataKeyOutput, err := cs.kmsService.GenerateDataKey(&kms.GenerateDataKeyInput{
		KeyId:   k,
		KeySpec: &keySpec,
	})
	if err != nil {
		return nil, err
	}
	block, err := x509.EncryptPEMBlock(rand.Reader, "RSA PRIVATE KEY", x509Bytes, dataKeyOutput.Plaintext, x509.PEMCipherAES256)
	pKey := pem.EncodeToMemory(block)
	encryptedData := EncryptedData{
		CipherContent: string(pKey),
		CipherKey:     base64.RawStdEncoding.EncodeToString(dataKeyOutput.CiphertextBlob),
	}
	gobBytes := bytes.NewBuffer([]byte{})
	encoder := gob.NewEncoder(gobBytes)
	err = encoder.Encode(encryptedData)

	if err != nil {
		return nil, err
	}
	return gobBytes.Bytes(), nil
}

// DecryptAsSSHKey decrypts the given ssh data and returns a signer
func (cs *KmsCryptService) DecryptAsSSHKey(data []byte) (ssh.Signer, error) {
	var encryptedData EncryptedData
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	err := decoder.Decode(&encryptedData)
	if err != nil {
		return nil, err
	}
	cipherBlob, err := base64.RawStdEncoding.DecodeString(encryptedData.CipherKey)
	if err != nil {
		return nil, err
	}
	decryptOut, err := cs.kmsService.Decrypt(&kms.DecryptInput{
		CiphertextBlob: cipherBlob,
	})
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKeyWithPassphrase([]byte(encryptedData.CipherContent), decryptOut.Plaintext)
}
