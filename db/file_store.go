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
    "io/ioutil"
    "os"
    "path"
)

// The FileStore implements the Datastore interface to provide flat file
// resource data storage.
type FileStore struct {
    basepath string
}

// The NewFileStore function returns a new FileStore instance.
func NewFileStore(basepath string) *FileStore {
    d := new(FileStore)
    d.basepath = basepath

    // Make directory if not exist.
    if _, e := os.Stat(basepath); os.IsNotExist(e) {
        os.MkdirAll(basepath, 0755)
    }

    return d
}

func (d *FileStore) TypeId() string { return "file" }

// The HasResource instance method returns true if this store has a specific resource.
func (d *FileStore) HasResource(resourceId ResourceId) bool {
    if _, e := os.Stat(path.Join(d.basepath, resourceId.GetId())); os.IsNotExist(e) {
        return false
    }
    return true
}

// The GetResourceData instance method.
func (d *FileStore) GetData(resourceId ResourceId) ([]byte, error) {
    if !d.HasResource(resourceId) { return nil, E_UNKNOWN_RESOURCE }
    data, e := ioutil.ReadFile(path.Join(d.basepath, resourceId.GetId()))
    return data, e
}

// The SetResourceData instance method.
func (d *FileStore) SetData(resourceId ResourceId, data []byte) error {
    filepath := path.Join(d.basepath, resourceId.GetId())
    return ioutil.WriteFile(filepath, data, 0644)
}

