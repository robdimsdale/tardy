package templates

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type _escLocalFS struct{}

var _escLocal _escLocalFS

type _escStaticFS struct{}

var _escStatic _escStaticFS

type _escDir struct {
	fs   http.FileSystem
	name string
}

type _escFile struct {
	compressed string
	size       int64
	local      string
	isDir      bool

	data []byte
	once sync.Once
	name string
}

func (_escLocalFS) Open(name string) (http.File, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	return os.Open(f.local)
}

func (_escStaticFS) prepare(name string) (*_escFile, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	var err error
	f.once.Do(func() {
		f.name = path.Base(name)
		if f.size == 0 {
			return
		}
		var gr *gzip.Reader
		b64 := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(f.compressed))
		gr, err = gzip.NewReader(b64)
		if err != nil {
			return
		}
		f.data, err = ioutil.ReadAll(gr)
	})
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fs _escStaticFS) Open(name string) (http.File, error) {
	f, err := fs.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.File()
}

func (dir _escDir) Open(name string) (http.File, error) {
	return dir.fs.Open(dir.name + name)
}

func (f *_escFile) File() (http.File, error) {
	type httpFile struct {
		*bytes.Reader
		*_escFile
	}
	return &httpFile{
		Reader:   bytes.NewReader(f.data),
		_escFile: f,
	}, nil
}

func (f *_escFile) Close() error {
	return nil
}

func (f *_escFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f *_escFile) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *_escFile) Name() string {
	return f.name
}

func (f *_escFile) Size() int64 {
	return f.size
}

func (f *_escFile) Mode() os.FileMode {
	return 0
}

func (f *_escFile) ModTime() time.Time {
	return time.Time{}
}

func (f *_escFile) IsDir() bool {
	return f.isDir
}

func (f *_escFile) Sys() interface{} {
	return f
}

// FS returns a http.Filesystem for the embedded assets. If useLocal is true,
// the filesystem's contents are instead used.
func FS(useLocal bool) http.FileSystem {
	if useLocal {
		return _escLocal
	}
	return _escStatic
}

// Dir returns a http.Filesystem for the embedded assets on a given prefix dir.
// If useLocal is true, the filesystem's contents are instead used.
func Dir(useLocal bool, name string) http.FileSystem {
	if useLocal {
		return _escDir{fs: _escLocal, name: name}
	}
	return _escDir{fs: _escStatic, name: name}
}

// FSByte returns the named file from the embedded assets. If useLocal is
// true, the filesystem's contents are instead used.
func FSByte(useLocal bool, name string) ([]byte, error) {
	if useLocal {
		f, err := _escLocal.Open(name)
		if err != nil {
			return nil, err
		}
		return ioutil.ReadAll(f)
	}
	f, err := _escStatic.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.data, nil
}

// FSMustByte is the same as FSByte, but panics if name is not present.
func FSMustByte(useLocal bool, name string) []byte {
	b, err := FSByte(useLocal, name)
	if err != nil {
		panic(err)
	}
	return b
}

// FSString is the string version of FSByte.
func FSString(useLocal bool, name string) (string, error) {
	b, err := FSByte(useLocal, name)
	return string(b), err
}

// FSMustString is the string version of FSMustByte.
func FSMustString(useLocal bool, name string) string {
	return string(FSMustByte(useLocal, name))
}

var _escData = map[string]*_escFile{

	"/templates/head.html.tmpl": {
		local: "web/assets/templates/head.html.tmpl",
		size:  564,
		compressed: `
H4sIAAAJbogA/5SSsU/rMBDG9/wVfje/xK99C0McCZUOTDAUCUbXvjYOjp3a15Qo6v+O2wgoSAxVhnz3
3X33k2yPo8aNccigRqnheMzKP3cPi9XL45LV1NoqK08/ZqXbCkAHVcZYeZo9iSRbJMlULUNEErCnTX4D
l62aqMtxtze9gOf86TZf+LaTZNYWgSnvCF3K3S8F6i1+SzrZooDe4KHzgS6GD0ZTLTT2RmF+Lv4y4wwZ
afOopEUx+1hEhixWKxn0UPKpyKZOVMF0xGjoEoTwjXgjezm5wGJQAjhXXmPR7PYYhkL5lk8ynxez9LXG
FU2EquRTqrpisUPSThZr7ylSkJ3S7gz4NPj/Yl784038sn4DWuNeWUArINJgMdaICVQH3FxDUvEnKjlw
Pq2ST/c9juh0eiHvAQAA//8fFfhvNAIAAA==
`,
	},

	"/templates/home.html.tmpl": {
		local: "web/assets/templates/home.html.tmpl",
		size:  237,
		compressed: `
H4sIAAAJbogA/1yPQarDMAxE1/+fQvVehHTt5hS9gBqpTcCJQ2zaBpO7V8bg0O5G85hBkxLLfZwFzOAn
WeghZt//U4oyLY5i9oU4ewD25nnrVKjk8Qm9oxAupvdzJK1YTWHfdPWv6v/mHL4Dtmflf4qGtrvSyptt
VB2JRiO1+DiKBHtChPoBIGZqm/KpNsXJdTpHZtYJnwAAAP//3YAUhe0AAAA=
`,
	},

	"/": {
		isDir: true,
		local: "web/assets",
	},

	"/templates": {
		isDir: true,
		local: "web/assets/templates",
	},
}
