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

package set

import (
    "bytes"
    "encoding/base64"
    "encoding/binary"
    "fmt"
    "hash/crc32"
    "io"
    "sync"
)

// Common Go representation of a grow only set.
type GSet struct {
    sync.RWMutex
    contents map[interface{}]struct{}
}

func NewGSet() GSet {
    return GSet{contents: make(map[interface{}]struct{})}
}

func (s GSet) Insert(item interface{}) bool {
    s.Lock()
    defer s.Unlock()

    _, found := s.contents[item]
    s.contents[item] = struct{}{}

    return !found
}

func (s GSet) Contains(item interface{}) bool {
    s.RLock()
    defer s.RUnlock()

    _, found := s.contents[item]

    return found
}

func (s GSet) Length() int {
    s.RLock()
    defer s.RUnlock()

    return len(s.contents)
}

func (s GSet) Equals(other GSet) bool {
    s.RLock()
    defer s.RUnlock()

    if len(s.contents) != other.Length() {
        return false
    }

    for i := range s.contents {
        if !other.Contains(i) {
            return false
        }
    }
    return true
}

func (s GSet) Clone() GSet {
    s.RLock()
    defer s.RUnlock()

    result := NewGSet()
    for i := range s.contents {
        result.Insert(i)
    }

    return result
}

func (s GSet) Merge(other GSet) {
    for i := range other.Iterate() {
        s.Insert(i)
    }
}

func (s GSet) Iterate() <-chan interface{} {
    ch := make(chan interface{})
    go func() {
        s.RLock()
        defer s.RUnlock()

        for i := range s.contents { ch <- i }

        close(ch)
    }()
    return ch
}

func (s GSet) ToSlice() []interface{} {
    s.RLock()
    defer s.RUnlock()

    result := make([]interface{}, 0, len(s.contents))
    for i := range s.contents { result = append(result, i) }

    return result
}

var GSET_HEADER_MAGIC = []byte{'c', 'r', 'd', 't', ':', 'g', 's', 'e', 't', 0x00}

func (s GSet) Serialize(buff *bytes.Buffer) error {
    s.RLock()
    defer s.RUnlock()

    buff.Write(GSET_HEADER_MAGIC)
    binary.Write(buff, binary.LittleEndian, uint32(len(s.contents)))

    for i := range s.contents {
        data, e := base64.StdEncoding.DecodeString(i.(string))
        if e != nil { return nil }

        binary.Write(buff, binary.LittleEndian, uint64(len(data)))
        binary.Write(buff, binary.LittleEndian, crc32.ChecksumIEEE(data))
        buff.Write(data)
    }

    return nil
}

func (s GSet) Deserialize(buff *bytes.Buffer) error {
    if buff.Len() < len(GSET_HEADER_MAGIC) { return fmt.Errorf("data too small") }

    header := make([]byte, len(GSET_HEADER_MAGIC))
    _, e := buff.Read(header)
    if e != nil || !bytes.Equal(header, GSET_HEADER_MAGIC) { return fmt.Errorf("invalid header") }

    var sizeof uint32
    if e := binary.Read(buff, binary.LittleEndian, &sizeof); e != nil {
        return e
    }

    for i := uint32(0); i < sizeof; i++ {
        var datl uint64
        var datc uint32

        if e := binary.Read(buff, binary.LittleEndian, &datl); e != nil {
            if e == io.EOF { return nil }
        }
        if e := binary.Read(buff, binary.LittleEndian, &datc); e != nil {
            return e
        }

        object := make([]byte, datl)
        if l, e := buff.Read(object); e != nil || l != int(datl) {
            return fmt.Errorf("invalid format")
        }

        if datc != crc32.ChecksumIEEE(object) { return fmt.Errorf("crc32 failure") }

        s.Insert(base64.StdEncoding.EncodeToString(object))
    }

    return nil
}

