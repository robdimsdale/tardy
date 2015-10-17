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
		size:  2500,
		compressed: `
H4sIAAAJbogA/7RWzW7bOBC++ykIdg8UrFByul2gMfYQFNjbPkHRAy3RNreUaJCjRELgd98hqTii/NMg
QA0EkDQz33z8OD+hnZPEgVUVrOli8QerTdU1soWMWynqgTCy7doKlGlZ9rJYPAlLWmMbob8ZbSz5m1AH
UuqN7iRdk6IgtdyKTgOpvD34a7Xbg7SngE9fH799/WcV3Ucj2XRAnGqUxgAwY46YrxF2p1qMfAFzeCD3
ZU6sj3ogn/FxYwBME5+13OLXP8tjviD4e1Y17DFutfqrJHcjDvdOb28BKbrvpX9G/y/lxB1zvr3EZOtF
8H8VhljZ1tKyWoDIyEuwEeKZ9whWf+auElpyrVopLMtG+/zHa9MI1TL0lz3gDQS8/E3+GrExFXS2JTVX
9Zocs6tgVrQ7yb6jJkGEH9nIOfIaPsDr+xU78UiNJ+7pTti+nLjWYnDrY5bfRBD9xxF+/EqHeLV5OdOh
f+yVG7V42nGBb4kOUSHWJ9+MVf52aKwFOlP2fYjDRURfmRO8ai8scHE4YHkxuqNJjACwjFZaOEdzQnvi
U11yAZTAbbGdvFt40QIkK3NKlq8lvyQ0S2ORpWZBno/xGS7wCZjDJUwntazgEc2Ub4RNw3xVxN46feUo
FzZcdiJjMfwmnzPUaO/plQbDYsQWy3yPnUcN16IG9q+Afahl7LxYttkcBIfeiBPVvwY23s0doo5Il9nc
RgmExMaxE0pALM9ojWhhXCDYPUlLtPViqupnkidXb/Mu/p5VW5tnbvBakBnAwT0UxXPn56NWDnhlmuJT
AcL9dIUvwKjyBOJ4lrYxuKPMk7TzIyaJfcOFKmKwxwJLbDyUvYpLjNedFeHxvixnfqkEq7nZwYC9S7dK
6ztzEJUCXwkl/5K94wAd/EK633CC+3cdYHWd/mt/4Sx7AAVapk0EuKnY1QXlx0oeLznM7xTbX/pxnAN+
coZJME7OKAPl4dsk5ex4cb8vk8W+TBb7WeipWU6Tb7Lml+manwZfGHq35ys94xW1SNL5qTtKgFNh+m9E
bAnU4j/nS6gQB1U8rWLjYKLoil5H/Ps/AAD//7Ji4MXECQAA
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
