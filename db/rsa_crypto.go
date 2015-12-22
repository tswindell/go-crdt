package crdb

import (
    "crypto/rsa"
    "crypto/rand"
    "crypto/sha1"
    "crypto/x509"
    "fmt"
)

type RSACryptoMethod struct {
    keysize int
    typeId  string
}

func NewRSACryptoMethod(keysize int) (*RSACryptoMethod, error) {
    d := new(RSACryptoMethod)
    d.keysize = keysize

    d.typeId = fmt.Sprintf("rsa-%d-sha1", d.keysize)

    return d, nil
}

func (d *RSACryptoMethod) TypeId() string {
    return d.typeId
}

func (d *RSACryptoMethod) GenerateKey() ResourceKey {
    key, e := rsa.GenerateKey(rand.Reader, d.keysize)
    if e != nil { return ResourceKey("") }
    return NewResourceKey(d.TypeId(), x509.MarshalPKCS1PrivateKey(key))
}

func (d *RSACryptoMethod) Encrypt(key ResourceKey, data []byte) ([]byte, error) {
    h := sha1.New()
    secret, e := x509.ParsePKCS1PrivateKey(key.KeyData())
    if e != nil { return nil, e }
    pubkey := secret.PublicKey
    return rsa.EncryptOAEP(h, rand.Reader, &pubkey, data, []byte{})
}

func (d *RSACryptoMethod) Decrypt(key ResourceKey, data []byte) ([]byte, error) {
    h := sha1.New()
    secret, e := x509.ParsePKCS1PrivateKey(key.KeyData())
    if e != nil { return nil, e }
    return rsa.DecryptOAEP(h, rand.Reader, secret, data, []byte{})
}

