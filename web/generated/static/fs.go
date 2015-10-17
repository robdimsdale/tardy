package static

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

	"/static/css/home.css": {
		local: "web/assets/static/css/home.css",
		size:  189,
		compressed: `
H4sIAAAJbogA/0zMwQrCMAyA4XufIuDVjXqtZx8kc9lWDGlpKlRk727cFL0l5M/XD1jg6abIHEArEQ98
p7NbneuxRYVKrb6DJDXAyecGiqKdUonTX5axLsfPzFHIXgB2VZIYCKaXdKMAB+/9ti+YqSsko1kyB7iW
qPkyzqQ73H70xo12Znx8xdW9AgAA//97HuIFvQAAAA==
`,
	},

	"/static/js/home.js": {
		local: "web/assets/static/js/home.js",
		size:  2360,
		compressed: `
H4sIAAAJbogA/7RW227qOBR9z1dYnnlwRGoCnc4DaB4qzev5gqo6chMDPk1iZG8gEeLfz7YdLgbSVpVO
JCTf9vLy2jfoxkpiwagC5jRJ/malLja1bCDlRoqyI4wsNk0BSjcs3SfJcUKWEv4XIFmZkn1CiJGwMQ1p
5I6EZV5u5M8Sh+k8OSTJVhhSC7NUDfmP7EGvZ2SaZ8So5Qpm5BGHbxpA12FcyQWu/pMfMoQmZKdKWKHd
ZPJvTh56HO4OnWceKRxfSTfG80/5xXG88zwJl80Tf/70JiObUhqGrEV4lfsc8xbBykcOqpbcFqKSLO13
rz9e6lqohuFp2QLq6NGys4hOr6NaZw3n5JAOQhrRLCV7QV28EK9pzztw6wI3T4tXqpHCfMruZWCfOKTa
0XekLzjve8boVtHZ+SHNPkQQ7fcRXj/TIbg3y690aJ9bZXsttksucBbp0Duujda0Uc5HNMQDvVL2a4jd
XUQXnRd4xUoY4GK9xhBjdEkjGwFgGC0qYS3NCG2Ju+reEUAJ7EKb2h3zk8rFT55RMjqG/YjQNLZFlhXz
8nyPT3eHj8fs7mFaWckCnnGb8jdhYjMXFSG/Tqsc5cKkS09kDJp/yOcGNey3dCDNWnZONJ9pt7bdkG3H
fghY+YjG/AvBOwQSHDCE5HHEm2Ud62GwHHUsH0LzuY5gUxLHV+OUUMV7dE+mzgUrfDvVlHrHNWqKzADW
djYe7zauwFXKAi90Pf5rDMK+27GLnpIrLEMXEIeba2uNzUJvpbl+YnSxyxYfAgxWGB3RHvcxq0I3wQ5h
hB9O8/zqXCzB5HrbQoeJRxeqqh70WhQKnANz/pR+4QEb+ES6P/CC6ZceMBmmf0wOLEQzUFDJOAMAmw27
G3bOr64mZMHJvvjG2M7phz6JXdnzadyXvSAD5X7t4sqr54UGPYo68yjqzDemp2Q5la2LPj2K+/Sl8Z2K
9XFxpDe8ghbRda5k9hLgs39ZFy1jsVbj7STkCGKGfwfu7wz+fgcAAP//oImuMTgJAAA=
`,
	},

	"/": {
		isDir: true,
		local: "web/assets",
	},

	"/static": {
		isDir: true,
		local: "web/assets/static",
	},

	"/static/css": {
		isDir: true,
		local: "web/assets/static/css",
	},

	"/static/js": {
		isDir: true,
		local: "web/assets/static/js",
	},
}
