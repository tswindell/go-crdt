package crdb

import (
    "testing"

    "bytes"
    "fmt"
)

func TestAESEncryptionMethod(t *testing.T) {
    method, _ := NewAESCryptoMethod(32)
    orig := []byte("Hello, world!")
    key := method.GenerateKey()

    fmt.Printf("KEY LENGTH: %d\n", len(key))

    fmt.Printf("ORIG: %x\n", orig)

    text, e := method.Encrypt(key, orig)
    if e != nil { t.Error(e) }

    fmt.Printf("TEXT: %x\n", text)

    result, e := method.Decrypt(key, text)
    if e != nil { t.Error(e) }

    fmt.Printf("DEC: %x\n", result)

    if !bytes.Equal(result, orig) {
        t.Error("Encryption failed!")
    }
}

