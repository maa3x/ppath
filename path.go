package ppath

import (
	"archive/zip"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/maa3x/errz"
	"github.com/shirou/gopsutil/v4/disk"
)

type Path string

func New(v ...string) Path {
	return Path(filepath.Join(v...))
}

// ThisFile retrieves the path of the source file from which it was invoked.
func ThisFile() Path {
	_, f, _, _ := runtime.Caller(1)
	return New(f)
}

// WD returns the path of the application’s current working directory.
func WD() Path {
	v := "."
	if wd, err := os.Getwd(); err == nil {
		v = wd
	}

	return New(v)
}

// ThisDir retrieves the path of the directory containing the source file from which it was invoked.
func ThisDir() Path {
	_, f, _, ok := runtime.Caller(1)
	if ok {
		return New(f).Dir()
	}

	return WD()
}

func Temp(dirs ...string) Path {
	f, err := TempFile("", dirs...)
	if err != nil {
		return New()
	}
	f.Close()
	return New(f.Name())
}

func TempFile(pattern string, dirs ...string) (*os.File, error) {
	return os.CreateTemp(filepath.Join(dirs...), pattern)
}

func TempDir(dirs ...string) Path {
	return New(append([]string{os.TempDir()}, dirs...)...)
}

func (p Path) String() string {
	return string(p)
}

func (p Path) Str() string {
	return string(p)
}

func (p Path) StringP() *string {
	return (*string)(&p)
}

func (p Path) Join(v ...string) Path {
	return Path(filepath.Join(append([]string{string(p)}, v...)...))
}

func (p Path) JoinPath(v ...Path) Path {
	s := make([]string, len(v))
	for i := range v {
		s[i] = string(v[i])
	}

	return p.Join(s...)
}

func (p Path) Base() Path {
	return Path(filepath.Base(string(p)))
}

func (p Path) BaseWithoutExt() Path {
	base := p.Base()
	segs := strings.Split(string(base), ".")
	if len(segs) == 1 || (len(segs) == 2 && segs[0] == "") {
		return base
	}
	return Path(strings.Join(segs[:len(segs)-1], "."))
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

func (p Path) Segments() []string {
	return filepath.SplitList(string(p))
}

func (p Path) Rel(r Path) (Path, error) {
	rel, err := filepath.Rel(string(r), string(p))
	return Path(rel), err
}

func (p Path) Abs() (Path, error) {
	if p.IsAbs() {
		return p, nil
	}

	abs, err := filepath.Abs(string(p))
	return Path(abs), err
}

func (p Path) IsChildOf(parent Path) bool {
	return strings.HasPrefix(string(p), string(parent))
}

func (p Path) IsParentOf(child Path) bool {
	return child.IsChildOf(p)
}

func (p Path) Delete() error {
	return os.RemoveAll(string(p))
}

func (p Path) Remove() error {
	return p.Delete()
}

func (p Path) Rename(n string) error {
	if err := Path(n).Dir().MkdirIfNotExist(); err != nil {
		return fmt.Errorf("create parent directory: %w", err)
	}
	return os.Rename(string(p), n)
}

func (p Path) Copy(dst Path) error {
	if p.IsDir() {
		if err := dst.MkdirIfNotExist(); err != nil {
			return err
		}
		return os.CopyFS(string(dst), os.DirFS(string(p)))
	}

	src, err := p.Open()
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer src.Close()

	if dst.IsDir() {
		dst = dst.JoinPath(p.Base())
	}
	if err := dst.Dir().MkdirIfNotExist(); err != nil {
		return fmt.Errorf("create parent directory: %w", err)
	}
	dest, err := dst.OpenFile(os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, src)
	return err
}

// MergeMove moves a file or directory from path p to dst.
//   - If dst doesn't exist: performs a straight move
//   - If p is a file and dst is a directory: moves p into dst
//   - If p is a file and dst is a file: replaces dst with p
//   - If p is a directory and dst is a directory: recursively merges contents
func (p Path) MergeMove(dst Path) error {
	if !p.Exists() {
		return errz.E("source file does not exist")
	}

	if !dst.Exists() {
		if err := dst.Dir().MkdirIfNotExist(); err != nil {
			return errz.E(err, "create parent directory")
		}
		if err := os.Rename(string(p), string(dst)); err != nil {
			return errz.E(err, "rename file")
		}
		return nil
	}

	if p.IsRegular() {
		if dst.IsDir() {
			dst = dst.JoinPath(p.Base())
			if err := os.Rename(string(p), string(dst)); err != nil {
				return errz.E(err, "rename file")
			}
			return nil
		}
		if !dst.IsRegular() {
			return errz.E("destination is not a regular file")
		}

		if err := dst.Delete(); err != nil {
			return errz.E(err, "delete old file")
		}
		if err := os.Rename(string(p), string(dst)); err != nil {
			return errz.E(err, "rename file")
		}
		return nil
	}

	if !p.IsDir() {
		return errz.E("source must be a regular file or directory")
	}
	if !dst.IsDir() {
		return errz.E("destination is not a directory")
	}

	entries, err := p.ReadDir()
	if err != nil {
		return errz.E(err, "reading directory entries")
	}
	for i := range entries {
		entryName := entries[i].Name()
		srcPath := p.Join(entryName)
		dstPath := dst.Join(entryName)
		if err := srcPath.MergeMove(dstPath); err != nil {
			return errz.E(err, "move file", "name", entryName)
		}
	}

	return p.Delete()
}

func (p Path) Move(dst Path) error {
	if !p.IsExist() {
		return errz.E("source file does not exist")
	}

	if err := dst.Dir().MkdirIfNotExist(); err != nil {
		return fmt.Errorf("make parent directory: %w", err)
	}

	return p.Rename(dst.String())
}

func (p Path) Truncate() error {
	if p.IsRegular() {
		if err := os.Truncate(string(p), 0); err != nil {
			return errz.E("truncate file", err)
		}
		return nil
	}

	if p.IsDir() {
		if err := p.Delete(); err != nil {
			return errz.E(err, "delete directory")
		}
		if err := p.MkdirIfNotExist(); err != nil {
			return errz.E("recreate directory", err)
		}
		return nil
	}

	return errz.E("unsupported target")
}

func (p Path) OpenFile(flag int, perm os.FileMode) (*os.File, error) {
	if p.IsDir() {
		return nil, errz.E("can not open a directory")
	}
	if err := p.Dir().MkdirIfNotExist(); err != nil {
		return nil, fmt.Errorf("create parent directory: %w", err)
	}
	return os.OpenFile(string(p), flag, perm)
}

func (p Path) Open() (*os.File, error) {
	return p.OpenFile(os.O_RDONLY, 0)
}

func (p Path) OpenOrCreate() (*os.File, error) {
	return p.OpenFile(os.O_RDWR|os.O_CREATE, 0o644)
}

func (p Path) Create() (*os.File, error) {
	if p.IsExist() {
		return nil, errz.E("already exists")
	}

	if err := p.Dir().MkdirIfNotExist(); err != nil {
		return nil, fmt.Errorf("create parent directory: %w", err)
	}

	return os.Create(string(p))
}

func (p Path) MkdirIfNotExist() error {
	if !p.IsExist() {
		return os.MkdirAll(string(p), 0o755)
	}

	if !p.IsDir() {
		return errz.E("already exists but not a directory")
	}

	return nil
}

func (p Path) ReadDir() ([]fs.DirEntry, error) {
	if !p.IsDir() {
		return nil, errz.E("not a directory")
	}

	entries, err := os.ReadDir(string(p))
	if err != nil {
		return nil, fmt.Errorf("read directory: %w", err)
	}
	return entries, nil
}

func (p Path) ReadFile() ([]byte, error) {
	return os.ReadFile(string(p))
}

func (p Path) ReadJSON() (any, error) {
	f, err := os.Open(string(p))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var v any
	err = json.NewDecoder(f).Decode(&v)
	return v, err
}

func (p Path) ReadFrom(r io.Reader) (int64, error) {
	dest, err := p.Create()
	if err != nil {
		return 0, err
	}
	defer dest.Close()

	return dest.ReadFrom(r)
}

func (p Path) ReadFromPath(p2 Path) (int64, error) {
	src, err := p2.Open()
	if err != nil {
		return 0, err
	}
	defer src.Close()

	return p.ReadFrom(src)
}

func (p Path) WriteFile(data []byte) error {
	if p.IsDir() {
		return errz.E("can not write to a directory")
	}
	if err := p.Dir().MkdirIfNotExist(); err != nil {
		return fmt.Errorf("create parent directory: %w", err)
	}
	return os.WriteFile(string(p), data, 0o644)
}

func (p Path) WriteJSON(v any) error {
	f, err := p.OpenFile(os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return errz.E(err, "open file")
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(v)
}

func (p Path) WriteTo(w io.Writer) (int64, error) {
	src, err := p.Open()
	if err != nil {
		return 0, err
	}
	defer src.Close()

	return src.WriteTo(w)
}

func (p Path) WriteToPath(p2 Path) error {
	dest, err := p2.Create()
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = p.WriteTo(dest)
	return err
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
	fi, err := os.Lstat(string(p))
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

func (p Path) Exists() bool {
	return p.IsExist()
}

func (p Path) DoesNotExist() bool {
	return !p.IsExist()
}

func (p Path) IsEqual(p2 Path) bool {
	if p == p2 {
		return true
	}

	abs1, err := p.Abs()
	if err != nil {
		return false
	}
	ab2, err := p2.Abs()
	if err != nil {
		return false
	}

	return abs1 == ab2
}

func (p Path) IsWritable() bool {
	if !p.IsExist() {
		return false
	}

	if p.IsDir() {
		tp := p.Join(".tmp_check_write")
		f, err := os.OpenFile(string(tp), os.O_WRONLY|os.O_CREATE, 0o600)
		if err != nil {
			return false
		}
		f.Close()
		tp.Delete()
		return true
	}

	if !p.IsRegular() {
		return false
	}

	f, err := os.OpenFile(string(p), os.O_WRONLY, 0o600)
	if err != nil {
		return false
	}
	f.Close()
	return true
}

func (p Path) IsEmpty() bool {
	if p.DoesNotExist() {
		return true
	}

	if p.IsDir() {
		entries, err := p.ReadDir()
		if err != nil {
			return false
		}
		return len(entries) == 0
	}

	size, err := p.Size()
	if err != nil {
		return false
	}
	return size == 0
}

func (p Path) HasPrefix(prefix string) bool {
	return strings.HasPrefix(string(p), prefix)
}

func (p Path) HasSuffix(suffix string) bool {
	return strings.HasSuffix(string(p), suffix)
}

func (p Path) HasExt(ext string) bool {
	if ext == "" {
		return true
	}
	if ext[0] != '.' {
		ext = "." + ext
	}
	return strings.HasSuffix(string(p), ext)
}

func (p Path) Contains(sub string) bool {
	return strings.Contains(string(p), sub)
}

func (p Path) Trim() Path {
	return Path(strings.TrimSpace(string(p)))
}

func (p Path) Match(pattern string) bool {
	v, err := filepath.Match(pattern, string(p))
	return err == nil && v
}

func (p Path) VolumeName() string {
	return filepath.VolumeName(string(p))
}

func (p Path) Clean() Path {
	return Path(filepath.Clean(string(p)))
}

func (p Path) Normalize() Path {
	return p.Clean()
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

// TotalSize returns the size of the file. For directories, it recursively
// calculates the total size including all entries.
func (p Path) TotalSize() (int64, error) {
	fi, err := p.Stat()
	if err != nil {
		return 0, err
	}

	// For regular files, return the file size
	if fi.Mode().IsRegular() {
		return fi.Size(), nil
	}

	if !fi.IsDir() {
		// For other types (symlinks, devices, etc.), return just the size
		return fi.Size(), nil
	}

	// For directories, calculate total size recursively
	total := fi.Size()

	entries, err := p.ReadDir()
	if err != nil {
		return 0, fmt.Errorf("read directory: %w", err)
	}

	for _, entry := range entries {
		entryPath := p.Join(entry.Name())
		size, err := entryPath.TotalSize()
		if err != nil {
			return 0, fmt.Errorf("get size of %s: %w", entry.Name(), err)
		}
		total += size
	}

	return total, nil
}

// TotalSizeX is a convenience method that returns TotalSize() without error.
func (p Path) TotalSizeX() int64 {
	size, _ := p.TotalSize()
	return size
}

func (p Path) Walk(fn fs.WalkDirFunc) error {
	return filepath.WalkDir(string(p), fn)
}

func (p Path) HasQuery() bool {
	return strings.Contains(string(p), "?")
}

func (p Path) WithoutQuery() Path {
	if !p.HasQuery() {
		return p
	}
	return Path(strings.Split(string(p), "?")[0])
}

func (p Path) WithQuery(q string) Path {
	if q == "" {
		return p.WithoutQuery()
	}
	return Path(string(p.WithoutQuery()) + "?" + q)
}

func (p Path) Query() string {
	if !p.HasQuery() {
		return ""
	}
	return strings.Split(string(p), "?")[1]
}

func (p Path) QuerySet(k string, v any) Path {
	if q, err := url.ParseQuery(p.Query()); err == nil {
		q.Set(k, toString(v))
		return p.WithQuery(q.Encode())
	}
	return p
}

func (p Path) QueryAdd(k string, v any) Path {
	if q, err := url.ParseQuery(p.Query()); err == nil {
		q.Add(k, toString(v))
		return p.WithQuery(q.Encode())
	}
	return p
}

func (p Path) QueryDel(k string) Path {
	if q, err := url.ParseQuery(p.Query()); err == nil {
		q.Del(k)
		return p.WithQuery(q.Encode())
	}
	return p
}

func (p Path) QueryHas(k string) bool {
	if q, err := url.ParseQuery(p.Query()); err == nil {
		return q.Has(k)
	}
	return false
}

func (p Path) hashFile(h hash.Hash) []byte {
	f, err := os.Open(string(p))
	if err != nil {
		return nil
	}
	f.WriteTo(h)
	f.Close()
	return h.Sum(nil)
}

func (p Path) MD5() []byte {
	return p.hashFile(md5.New())
}

func (p Path) SHA1() []byte {
	return p.hashFile(sha1.New())
}

func (p Path) SHA256() []byte {
	return p.hashFile(sha256.New())
}

func (p Path) XXH64() []byte {
	return p.hashFile(xxhash.New())
}

func (p Path) hashFileString(h hash.Hash) string {
	if s := p.hashFile(h); len(s) > 0 {
		return hex.EncodeToString(s)
	}
	return ""
}

func (p Path) MD5String() string {
	return p.hashFileString(md5.New())
}

func (p Path) SHA1String() string {
	return p.hashFileString(sha1.New())
}

func (p Path) SHA256String() string {
	return p.hashFileString(sha256.New())
}

func (p Path) XXH64String() string {
	return p.hashFileString(xxhash.New())
}

type Usage struct {
	Total       uint64
	Used        uint64
	Free        uint64
	UsedPercent float64
}

func (p Path) Times() (created, modified, accessed time.Time) {
	return getTimes(string(p))
}

func (p Path) Usage() (u Usage, err error) {
	var s *disk.UsageStat
	if s, err = disk.Usage(string(p)); err != nil {
		return Usage{}, err
	}

	return Usage{
		Total:       s.Total,
		Used:        s.Used,
		Free:        s.Free,
		UsedPercent: s.UsedPercent,
	}, nil
}

func (p Path) WriteZipArchive(zipFilePath Path) (retErr error) {
	if !p.Exists() {
		return errz.E("src directory does not exist")
	}

	zipFile, err := zipFilePath.Create()
	if err != nil {
		return errz.E("create zip file", err)
	}
	defer func() {
		if err := zipFile.Close(); err != nil && retErr == nil {
			retErr = errz.E("close zip file", err)
		}
	}()

	archive := zip.NewWriter(zipFile)
	defer func() {
		if err := archive.Close(); err != nil && retErr == nil {
			retErr = errz.E("close zip archive writer", err)
		}
	}()

	if !p.IsDir() {
		stat, err := p.Stat()
		if err != nil {
			return errz.E("read file stat", err)
		}
		header, err := zip.FileInfoHeader(stat)
		if err != nil {
			return errz.E("create zip header from file info", err)
		}
		return p.writeZipFile(string(p.Base()), archive, header)
	}

	walkErr := p.Walk(func(path string, d fs.DirEntry, itemErr error) error {
		if itemErr != nil {
			return errz.E("walk item error", itemErr)
		}
		if path == string(p) || path == string(zipFilePath) {
			return nil
		}

		relPath, err := filepath.Rel(string(p), path)
		if err != nil {
			return errz.E("calculate relative path", err)
		}

		info, err := d.Info()
		if err != nil {
			return errz.E("get file info", err)
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return errz.E("create zip header from file info", err)
		}

		if d.IsDir() {
			header.Name = filepath.ToSlash(relPath) + "/"
			if _, err := archive.CreateHeader(header); err != nil {
				return errz.E("create zip directory header", err)
			}
			return nil
		}

		return Path(path).writeZipFile(relPath, archive, header)
	})
	if walkErr != nil {
		return errz.E("walk directory for zip archive", walkErr)
	}

	return nil
}

func (p Path) writeZipFile(relPath string, archive *zip.Writer, header *zip.FileHeader) error {
	header.Method = zip.Deflate
	header.Name = filepath.ToSlash(relPath)
	w, err := archive.CreateHeader(header)
	if err != nil {
		return errz.E("create zip file header", err)
	}
	if _, err := Path(p).WriteTo(w); err != nil {
		return errz.E("write file content to zip", err)
	}
	return nil
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}
