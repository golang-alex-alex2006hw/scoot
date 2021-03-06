// Code generated by go-bindata.
// sources:
// config/config.go
// config/local.local
// config/local.memory
// DO NOT EDIT!

package config

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _configConfigGo = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x01\x00\x00\xff\xff\x00\x00\x00\x00\x00\x00\x00\x00")

func configConfigGoBytes() ([]byte, error) {
	return bindataRead(
		_configConfigGo,
		"config/config.go",
	)
}

func configConfigGo() (*asset, error) {
	bytes, err := configConfigGoBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "config/config.go", size: 0, mode: os.FileMode(420), modTime: time.Unix(1521487177, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _configLocalLocal = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x6c\x8e\x41\x4b\x03\x31\x10\x46\xef\xfb\x2b\x42\xce\x45\x8b\xe2\x65\xaf\xed\x49\x2c\x2e\xed\x82\xe7\x69\x76\x12\x43\x67\x77\x96\x99\x89\x58\xa4\xff\x5d\xa2\xad\x0a\xf5\x98\x79\x2f\x8f\xef\xa3\x71\xce\xaf\xa8\xa8\xa1\xf8\xd6\xd5\xa7\x73\xbe\x3f\xce\xe8\x5b\xe7\x89\x03\x90\x6f\x9c\x3b\x2d\xaa\xf7\xc2\x72\x40\xd1\x6b\x4f\xe6\xe0\x17\xdf\xa7\x8e\x89\xf2\x94\x3a\x94\xcc\x43\x65\x77\x0f\xcb\x51\x7f\x1b\xbb\xf0\x8a\x43\x21\x94\x15\x4f\x31\xa7\xeb\x96\x1a\x18\xc6\x42\x97\xe0\x06\xde\xb7\x68\x92\x51\x3b\x94\x1e\xf4\xe0\x5d\xeb\x96\x67\xb8\xc6\x7d\x49\x1b\x1e\xb0\x1e\x23\x90\xe2\x19\x6c\x31\xf0\x1b\xca\x23\xef\xf5\x79\xda\x19\x88\x95\xb9\x3a\x26\x05\x7f\xfe\x46\x28\x64\x35\xd9\xe7\x11\xb9\x58\x15\xfc\xfd\x72\xfc\xb3\x16\x12\x3c\xf1\x3f\x2b\x63\x26\xbc\x2c\x5c\x67\xc1\x60\x2c\xc7\x0a\x6e\x34\x30\xdb\x00\x06\xb7\xd5\x51\x48\x40\x9c\xbe\x8a\xcd\xa9\xf9\x0c\x00\x00\xff\xff\x1b\xaf\xe8\xca\x6f\x01\x00\x00")

func configLocalLocalBytes() ([]byte, error) {
	return bindataRead(
		_configLocalLocal,
		"config/local.local",
	)
}

func configLocalLocal() (*asset, error) {
	bytes, err := configLocalLocalBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "config/local.local", size: 367, mode: os.FileMode(420), modTime: time.Unix(1521481677, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _configLocalMemory = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x6c\x8f\xc1\x4a\x03\x41\x10\x44\xef\xfb\x15\x4d\x9f\x3d\xac\x78\x9b\xeb\xe6\x24\x06\x25\xd9\x1f\xe8\x24\x35\xeb\xe2\xcc\x76\x98\xe9\x16\x83\xe4\xdf\x65\x70\x45\xd0\x1c\xbb\xea\xf1\xa8\xfe\xec\x88\x78\x48\x5e\x0d\x85\x03\xb5\x93\x88\xc7\xcb\x19\x1c\x88\x33\xb2\x96\x0b\xdf\x7d\xa7\x83\xfa\x62\x1c\xe8\xbe\xef\x88\xae\x2d\xe4\xfd\xf1\x15\x27\x4f\x28\x83\x2e\x71\x9e\xfe\x1b\xaa\x89\x21\x7a\xfa\x71\x6c\xe5\x63\x07\x2b\x33\xea\x0b\xca\x28\xf5\x8d\x29\x50\xbf\x96\x1b\x1c\x7c\xda\xea\x09\x2d\x8c\x92\x2a\xd6\x62\x87\xa3\xbe\xa3\x3c\xea\xa1\x3e\x2f\x7b\x93\x62\x7e\xfe\xcb\x6c\x10\xc5\x93\x35\xe7\x38\x67\xa8\x5b\x23\xf8\xa1\xcf\xfc\x3b\x57\x26\x79\xd2\x1b\x33\xd7\x47\x1b\xd8\x5d\xbb\xaf\x00\x00\x00\xff\xff\xa5\xf8\x26\x2e\x15\x01\x00\x00")

func configLocalMemoryBytes() ([]byte, error) {
	return bindataRead(
		_configLocalMemory,
		"config/local.memory",
	)
}

func configLocalMemory() (*asset, error) {
	bytes, err := configLocalMemoryBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "config/local.memory", size: 277, mode: os.FileMode(420), modTime: time.Unix(1521481677, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"config/config.go":    configConfigGo,
	"config/local.local":  configLocalLocal,
	"config/local.memory": configLocalMemory,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"config": &bintree{nil, map[string]*bintree{
		"config.go":    &bintree{configConfigGo, map[string]*bintree{}},
		"local.local":  &bintree{configLocalLocal, map[string]*bintree{}},
		"local.memory": &bintree{configLocalMemory, map[string]*bintree{}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
