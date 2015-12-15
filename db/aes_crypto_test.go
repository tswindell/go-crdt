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
    "strings"
)

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

