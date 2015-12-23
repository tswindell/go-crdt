package crdb

import (
    "crypto/rsa"
    "crypto/rand"
    "crypto/sha1"
    "crypto/x509"
    "encoding/hex"
    "fmt"
)

type RSAPrivateKeyRing ThreadSafeMap

func (d *RSAPrivateKeyRing) GetPrivateKey(resourceKey ResourceKey) *rsa.PrivateKey {
    fingerprint := hex.EncodeToString(resourceKey.KeyData())

    value := ThreadSafeMap(*d).GetValue(fingerprint)

    if value == nil {
        key, e := x509.ParsePKCS1PrivateKey(resourceKey.KeyData())
        if e != nil { return nil }

        ThreadSafeMap(*d).Insert(fingerprint, key)

        key.Precompute()

        return key
    }

    return value.(*rsa.PrivateKey)
}

type RSACryptoMethod struct {
    keysize int
    typeId  string

    keyring RSAPrivateKeyRing
}

func NewRSACryptoMethod(keysize int) (*RSACryptoMethod, error) {
    d := new(RSACryptoMethod)
    d.keysize = keysize
    d.typeId = fmt.Sprintf("rsa-%d-sha1", d.keysize)
    d.keyring = RSAPrivateKeyRing(NewThreadSafeMap())

    return d, nil
}

func (d *RSACryptoMethod) TypeId() string {
    return d.typeId
}

func (d *RSACryptoMethod) GenerateKey() ResourceKey {
    key, e := rsa.GenerateKey(rand.Reader, d.keysize)
    if e != nil { return ResourceKey("") }

    serialized := x509.MarshalPKCS1PrivateKey(key)

    resourceKey := NewResourceKey(d.TypeId(), serialized)

    d.keyring.GetPrivateKey(resourceKey)

    return resourceKey
}

func (d *RSACryptoMethod) Encrypt(key ResourceKey, data []byte) ([]byte, error) {
    h := sha1.New()
    secret := d.keyring.GetPrivateKey(key)
    if secret == nil { return nil, E_INVALID_KEY }
    pubkey := secret.PublicKey
    return rsa.EncryptOAEP(h, rand.Reader, &pubkey, data, []byte{})
}

func (d *RSACryptoMethod) Decrypt(key ResourceKey, data []byte) ([]byte, error) {
    h := sha1.New()
    secret := d.keyring.GetPrivateKey(key)
    if secret == nil { return nil, E_INVALID_KEY }
    return rsa.DecryptOAEP(h, rand.Reader, secret, data, []byte{})
}

