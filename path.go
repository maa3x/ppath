package path

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type Path string

func New(v ...string) Path {
	return Path(filepath.Join(v...))
}

func (p Path) String() string {
	return string(p)
}

func (p Path) Join(v ...string) Path {
	return Path(filepath.Join(append([]string{string(p)}, v...)...))
}

func (p Path) Base() Path {
	return Path(filepath.Base(string(p)))
}

func (p Path) Dir() Path {
	return Path(filepath.Dir(string(p)))
}

func (p Path) NthParent(n int) Path {
	v := p
	for range n {
		v = v.Dir()
	}
	return v
}

func (p Path) Ext() Path {
	return Path(filepath.Ext(string(p)))
}

func (p Path) Split() (dir, file Path) {
	p1, p2 := filepath.Split(string(p))
	return Path(p1), Path(p2)
}

func (p Path) Rel(r Path) (Path, error) {
	rel, err := filepath.Rel(string(r), string(p))
	return Path(rel), err
}

func (p Path) Abs() (Path, error) {
	abs, err := filepath.Abs(string(p))
	return Path(abs), err
}

func (p Path) Delete() error {
	return os.RemoveAll(string(p))
}

func (p Path) Rename(n string) error {
	return os.Rename(string(p), n)
}

func (p Path) Copy(dst Path) error {
	if p.IsDir() {
		if dst.IsExist() && !dst.IsDir() {
			return errors.New("destination is not a directory, cannot copy directory to file")
		}
		if !dst.IsExist() {
			if err := os.MkdirAll(string(dst), 0o755); err != nil {
				return err
			}
		}
		return os.CopyFS(string(dst), os.DirFS(string(p)))
	}

	src, err := p.Open()
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer src.Close()

	if dst.IsDir() {
		dst = dst.Join(p.Base().String())
	}

	var dest io.WriteCloser
	if dst.IsExist() {
		if dest, err = dst.Open(); err != nil {
			return err
		}
	} else {
		if dest, err = dst.Create(); err != nil {
			return err
		}
	}
	defer dest.Close()

	_, err = io.Copy(dest, src)
	return err
}

func (p Path) Open() (*os.File, error) {
	return os.Open(string(p))
}

func (p Path) Create() (*os.File, error) {
	if p.IsExist() {
		return nil, errors.New("already exists")
	}

	if !p.Dir().IsExist() {
		if err := os.MkdirAll(p.Dir().String(), 0o755); err != nil {
			return nil, fmt.Errorf("create parent directory: %w", err)
		}
	}

	return os.Create(string(p))
}

func (p Path) ReadFile() ([]byte, error) {
	return os.ReadFile(string(p))
}

func (p Path) WriteFile(data []byte) error {
	return os.WriteFile(string(p), data, 0o644)
}

func (p Path) IsAbs() bool {
	return filepath.IsAbs(string(p))
}

func (p Path) IsLocal() bool {
	return filepath.IsLocal(string(p))
}

func (p Path) IsValid() bool {
	return fs.ValidPath(string(p))
}

func (p Path) IsRegular() bool {
	fi, err := p.Stat()
	if err != nil {
		return false
	}
	return fi.Mode().IsRegular()
}

func (p Path) IsDir() bool {
	fi, err := p.Stat()
	if err != nil {
		return false
	}
	return fi.IsDir()
}

func (p Path) IsSymlink() bool {
	fi, err := p.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&fs.ModeSymlink != 0
}

func (p Path) IsDev() bool {
	fi, err := p.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&fs.ModeDevice != 0
}

func (p Path) IsExist() bool {
	_, err := p.Stat()
	return err == nil
}

func (p Path) Match(pattern string) bool {
	v, err := filepath.Match(pattern, string(p))
	return err == nil && v
}

func (p Path) VolumeName() string {
	return filepath.VolumeName(string(p))
}

func (p Path) Stat() (fs.FileInfo, error) {
	return os.Stat(string(p))
}

func (p Path) Size() (int64, error) {
	fi, err := p.Stat()
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

func (p Path) SizeX() int64 {
	size, _ := p.Size()
	return size
}

func (p Path) Walk(fn fs.WalkDirFunc) error {
	return filepath.WalkDir(string(p), fn)
}
