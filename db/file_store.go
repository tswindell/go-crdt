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
    return d
}

func (d *FileStore) Type() string { return "file" }

// The HasResource instance method returns true if this store has a specific resource.
func (d *FileStore) HasResource(resourceId ResourceId) bool {
    if _, e := os.Stat(path.Join(d.basepath, "store", resourceId.GetId())); os.IsNotExist(e) {
        return false
    }
    return true
}

// The GetResourceData instance method.
func (d *FileStore) GetData(resourceId ResourceId) ([]byte, error) {
    if !d.HasResource(resourceId) { return nil, E_UNKNOWN_RESOURCE }
    data, e := ioutil.ReadFile(path.Join(d.basepath, "store", resourceId.GetId()))
    return data, e
}

// The SetResourceData instance method.
func (d *FileStore) SetData(resourceId ResourceId, data []byte) error {
    filepath := path.Join(d.basepath, "store", resourceId.GetId())
    return ioutil.WriteFile(filepath, data, 0644)
}

