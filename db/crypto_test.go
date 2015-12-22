/*
 * Copyright (c) 2015 Tom Swindell (t.swindell@rubyx.co.uk)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */

package crdb

import (
    "testing"

    "bytes"
    "fmt"
    "strings"
)

var (
    methods []CryptoMethod
)

func InitMethods(t *testing.T) {
    if methods != nil { return }

    methods = make([]CryptoMethod, 0)

    var m CryptoMethod

    // AES CBC Based Encryption Methods
    m, _ = NewAESCryptoMethod(AES_128_KEY_SIZE)
    methods = append(methods, m)
    m, _ = NewAESCryptoMethod(AES_194_KEY_SIZE)
    methods = append(methods, m)
    m, _ = NewAESCryptoMethod(AES_256_KEY_SIZE)
    methods = append(methods, m)

    // RSA Based Encryption Methods
    m, _ = NewRSACryptoMethod(1028) // Insecure
    methods = append(methods, m)
    m, _ = NewRSACryptoMethod(2048)
    methods = append(methods, m)
    m, _ = NewRSACryptoMethod(4096)
    methods = append(methods, m)
}

func TestCryptoMethods(t *testing.T) {
    InitMethods(t)

    for _, m := range methods {
        fmt.Printf("Testing cryptography method: %s\n", m.TypeId())

        fmt.Printf("  Generating key...")
        key := m.GenerateKey()
        fmt.Println("Done")

        if key.TypeId() != m.TypeId() { t.Fatal("Key and method type mismatch") }
        if len(key.KeyData()) == 0 { t.Fatal("Key length is zero!") }

        clearText := []byte("Hello, world!")
        encrypted, e := m.Encrypt(key, clearText)
        if e != nil { t.Fatal(e) }

        decrypted, e := m.Decrypt(key, encrypted)
        if e != nil { t.Fatal(e) }

        if !bytes.Equal(clearText, decrypted) {
            t.Fatal("Message data mismatch.")
        }
    }
}

func TestAESGenerateKey(t *testing.T) {
    method, _ := NewAESCryptoMethod(AES_256_KEY_SIZE)
    key := method.GenerateKey()

    parts := strings.SplitN(string(key), ":", 2)

    if parts[0] != key.TypeId() {
        t.Errorf("Key format error, missing type prefix. %s != %s", parts[0], key.TypeId())
    }
}

func TestAESEncryptionMethod(t *testing.T) {
    method, _ := NewAESCryptoMethod(AES_256_KEY_SIZE)

    if method.TypeId() != "aes-256-cbc" {
        t.Error("Crypto method TypeId is wrong: %s", method.TypeId())
    }

    orig := []byte("Hello, world!")
    key := method.GenerateKey()

    text, e := method.Encrypt(key, orig)
    if e != nil { t.Error(e) }

    result, e := method.Decrypt(key, text)
    if e != nil { t.Error(e) }

    if !bytes.Equal(result, orig) {
        t.Error("Encryption failed!")
    }
}

