package crdb

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/hmac"
    "crypto/rand"
    "crypto/sha256"
    "fmt"
)

const (
    AES_128_KEY_SIZE = 16
    AES_194_KEY_SIZE = 24
    AES_256_KEY_SIZE = 32
)

type AESCryptoMethod struct {
    keysize int
    cryptoType string

    ivsize int
    macsize int
    ckeysize int
    mkeysize int
}

func NewAESCryptoMethod(keysize int) (*AESCryptoMethod, error) {
    if keysize != AES_128_KEY_SIZE &&
       keysize != AES_194_KEY_SIZE &&
       keysize != AES_256_KEY_SIZE {
        return nil, E_INVALID_KEY
    }

    d := new(AESCryptoMethod)
    d.cryptoType = fmt.Sprintf("aes-%d-cbc", keysize * 8)

    d.ivsize   = aes.BlockSize
    d.macsize  = keysize
    d.ckeysize = keysize
    d.mkeysize = keysize
    d.keysize  = d.ckeysize + d.mkeysize

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
    if len(data) == 0 { return nil }
    pbyte := data[len(data) - 1]
    if int(pbyte) > len(data) || pbyte >= aes.BlockSize || pbyte == 0 { return nil }
    for i := len(data) - 1; i > len(data) - int(pbyte) - 1; i-- {
        if data[i] != pbyte { return nil }
    }
    return data[:len(data) - int(pbyte)]
}


// The Encrypt() instance method
func (d *AESCryptoMethod) Encrypt(resourceKey ResourceKey, data []byte) ([]byte, error) {
    keydata := resourceKey.KeyData()

    // Validate key length.
    if len(keydata) != d.keysize { return nil, E_INVALID_KEY }

    // Build initialization vector.
    iv := make([]byte, d.ivsize)
    _, e := rand.Read(iv)
    if e != nil { return nil, e }

    // Pad to aes.BlockSize
    data = __append_padding(data)
    text := make([]byte, len(data))

    // Encrypt AES-%d-CBC
    ci, _ := aes.NewCipher(keydata[:d.ckeysize])
    cbc := cipher.NewCBCEncrypter(ci, iv)

    cbc.CryptBlocks(text, data)

    // Apply HMAC
    hm := hmac.New(sha256.New, keydata[d.ckeysize:])
    text = append(iv, text...)
    hm.Write(text)

    return hm.Sum(text), nil
}

// The Decrypt instance method
func (d *AESCryptoMethod) Decrypt(resourceKey ResourceKey, data []byte) ([]byte, error) {
    keydata := resourceKey.KeyData()

    // Validate key length.
    if len(keydata) != d.keysize { return nil, E_INVALID_KEY }

    // Validate HMAC size
    if (len(data) % aes.BlockSize) != 0 { return nil, E_INVALID_RESOURCE_DATA }

    // Check against minimum message length (HMAC + IV + 1 or more Message Blocks)
    if len(data) < (aes.BlockSize + d.ivsize + d.macsize) { return nil, E_INVALID_RESOURCE_DATA }

    // Extract HMAC
    macs := len(data) - d.macsize
    mtag := data[macs:]
    text := data[:macs]

    // Check HMAC
    hm := hmac.New(sha256.New, keydata[d.ckeysize:])
    hm.Write(text)
    mac := hm.Sum(nil)
    if !hmac.Equal(mac, mtag) { return nil, fmt.Errorf("Invalid HMAC in data.") }

    // Decrypt AES-%d-CBC
    ci, _ := aes.NewCipher(keydata[:d.ckeysize])
    cbc := cipher.NewCBCDecrypter(ci, text[:d.ivsize])

    result := make([]byte, macs - d.ivsize)
    cbc.CryptBlocks(result, text[d.ivsize:])

    result = __remove_padding(result)
    return result, nil
}

