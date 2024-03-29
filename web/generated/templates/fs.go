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
		size:  793,
		compressed: `
H4sIAAAJbogA/5RTzW7UMBC+71MYn4lNW4EQiiOh0gMnOBQJjrP2bO3g2Kk9mzaK+u44yZaWRZW6yiHj
sb8ff5lMk8GdC8i4RTD84WFTv/ny7fL61/crZqnzzaaeX8xDuFEcA282jNXz2bkoZYcETFtIGUnxPe2q
j/z5liXqK7zdu0Hxn9WPz9Vl7Hogt/XImY6BMBTc1yuF5gb/QQboUPHB4V0fEz07fOcMWWVwcBqrZfGW
ueDIga+yBo/q7JGIHHlsriGZsZbrYrPuZJ1cT4zGvogQ3pNsYYC1y1lOWnEpdTQo2ts9plHo2Mm1rM7F
WXk6F0SbeVPLFdWcQByQTACxjZEyJei1CYvA34a8EOfinWzzU+slQe/Cb5bQK55p9JgtYhGyCXenKOl8
LFU6/Citxf38RfOnEo4JbRbax73ZeUi40EIL99K7bZbmojC/Fx9K8ej8vzl5uskrrpKpjI1ejNrY4cHf
qzM/wNsD+jjIWq5DPU0YTPkN/gQAAP//nN+ZSRkDAAA=
`,
	},

	"/templates/home.html.tmpl": {
		local: "web/assets/templates/home.html.tmpl",
		size:  270,
		compressed: `
H4sIAAAJbogA/1yPQc7CIBCF1/9/ipE9IXWNnMILjGVsm9DSAKk2pHd3KorR3cz7eI95OVu6DhOB6P1I
M3Yktu0/50Tj7DDtOqHdNQB98XY1PPBohwVahzGeROunhBwRRGHfNPhb1X99Tt6jbI7M/xj1jTljsKtW
PH0cii01uCzvLS5dDesxJGG0Yu114vMt6IOUUE8EKXeqVanCX6XRGe5Lk+WOjwAAAP//+BtKZQ4BAAA=
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
