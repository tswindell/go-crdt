package crdb

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/hmac"
    "crypto/rand"
    "crypto/sha256"
    "fmt"
)

type AESCryptoMethod struct {
    keysize int
    cryptoType string
}

//TODO Validate key size, 16, 24, 32 only
func NewAESCryptoMethod(keysize int) (*AESCryptoMethod, error) {
    d := new(AESCryptoMethod)
    d.keysize = keysize
    d.cryptoType = fmt.Sprintf("aes-%d-cbc", d.keysize * 8)
    return d, nil
}

func (d *AESCryptoMethod) Type() string {
    return d.cryptoType
}

func (d *AESCryptoMethod) GenerateKey() ResourceKey {
    key := make([]byte, d.keysize)
    _, e := rand.Read(key)
    if e != nil { return ResourceKey("") }
    return NewResourceKey(d.Type(), key)
}

// PKCS #7 padding implementation
func __append_padding(data []byte) []byte {
    padding := aes.BlockSize - (len(data) % aes.BlockSize)
    for i := 0; i < padding; i++ { data = append(data, byte(padding)) }
    return data
}

func __remove_padding(data []byte) []byte {
    pn := data[len(data)-1]
    if int(pn) > len(data) || pn > aes.BlockSize || pn == 0 { return data }
    for i := 0; i < int(pn); i++ { if data[len(data) - i] != pn { return data } }
    return data[:len(data) - int(pn)]
}


// The Encrypt() instance method
func (d *AESCryptoMethod) Encrypt(resourceKey ResourceKey, data []byte) ([]byte, error) {
    keydata := resourceKey.KeyData()

    // Validate key length.
    if len(keydata) != d.keysize { return nil, E_INVALID_KEY }

    // Build initialization vector.
    iv := make([]byte, aes.BlockSize)
    _, e := rand.Read(iv)
    if e != nil { return nil, e }

    // Pad to aes.BlockSize
    data = __append_padding(data)

    // Encrypt AES-%d-CBC
    c, _ := aes.NewCipher(keydata)
    cbc := cipher.NewCBCEncrypter(c, iv)

    result := make([]byte, len(data))
    cbc.CryptBlocks(result, data)

    // Apply HMAC
    hm := hmac.New(sha256.New, keydata[d.keysize:])
    result = append(iv, result...)
    hm.Write(result)

    return hm.Sum(result), nil
}

// The Decrypt instance method
func (d *AESCryptoMethod) Decrypt(resourceKey ResourceKey, data []byte) ([]byte, error) {
    keydata := resourceKey.KeyData()

    // Validate key length.
    if len(keydata) != d.keysize { return nil, E_INVALID_KEY }

    // Validate HMAC size
    if (len(data) % aes.BlockSize) != 0 { return nil, E_INVALID_RESOURCE_DATA }

    // Check against minimum message length (HMAC + IV)
    if len(data) < (4 * aes.BlockSize) { return nil, E_INVALID_RESOURCE_DATA }

    // Extract HMAC
    macoffset := len(data) - (2 * aes.BlockSize)
    macanswer := data[macoffset:]

    // Extract actual content.
    citext := data[:macoffset]

    // Check HMAC
    hm := hmac.New(sha256.New, keydata[d.keysize:])
    hm.Write(citext)
    mac := hm.Sum(nil)
    if !hmac.Equal(mac, macanswer) { return nil, E_INVALID_RESOURCE_DATA }

    // Decrypt AES-%d-CBC
    c, _ := aes.NewCipher(keydata)
    cbc := cipher.NewCBCDecrypter(c, citext[:aes.BlockSize])

    result := make([]byte, macoffset - (2 * aes.BlockSize))
    cbc.CryptBlocks(result, data[aes.BlockSize:])

    result = __remove_padding(result)
    return result, nil
}

