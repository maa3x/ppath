package ppath

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var testContent = []byte("test content")

func errorIf(t *testing.T, err error) {
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNew(t *testing.T) {
	p := New("a", "b", "c")
	expected := filepath.Join("a", "b", "c")
	if p.String() != expected {
		t.Errorf("expected %s, got %s", expected, p.String())
	}
}

func TestJoin(t *testing.T) {
	p := New("a", "b")
	p = p.Join("c", "d")
	expected := filepath.Join("a", "b", "c", "d")
	if p.String() != expected {
		t.Errorf("expected %s, got %s", expected, p.String())
	}
}

func TestBase(t *testing.T) {
	p := New("a", "b", "c")
	expected := "c"
	if p.Base().String() != expected {
		t.Errorf("expected %s, got %s", expected, p.Base().String())
	}
}

func TestDir(t *testing.T) {
	p := New("a", "b", "c")
	expected := filepath.Join("a", "b")
	if p.Dir().String() != expected {
		t.Errorf("expected %s, got %s", expected, p.Dir().String())
	}
}

func TestNthParent(t *testing.T) {
	p := New("a", "b", "c", "d")
	expected := filepath.Join("a", "b")
	if p.NthParent(2).String() != expected {
		t.Errorf("expected %s, got %s", expected, p.NthParent(2).String())
	}
}

func TestExt(t *testing.T) {
	p := New("a", "b", "c.txt")
	expected := ".txt"
	if p.Ext().String() != expected {
		t.Errorf("expected %s, got %s", expected, p.Ext().String())
	}
}

func TestSplit(t *testing.T) {
	p := New("a", "b", "c.txt")
	dir, file := p.Split()
	expectedDir, expectedFile := filepath.Split(filepath.Join("a", "b", "c.txt"))
	if dir.String() != expectedDir || file.String() != expectedFile {
		t.Errorf("expected (%s, %s), got (%s, %s)", expectedDir, expectedFile, dir.String(), file.String())
	}
}

func TestRel(t *testing.T) {
	p := New("a", "b", "c", "d")
	r := New("a", "b")
	expected := filepath.Join("c", "d")
	rel, err := p.Rel(r)
	if err != nil || rel.String() != expected {
		t.Errorf("expected %s, got %s, error: %v", expected, rel.String(), err)
	}
}

func TestAbs(t *testing.T) {
	p := New(".")
	abs, err := p.Abs()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expected, _ := filepath.Abs(".")
	if abs.String() != expected {
		t.Errorf("expected %s, got %s", expected, abs.String())
	}
}

func TestDelete(t *testing.T) {
	p := New("testdir")
	os.Mkdir(p.String(), 0o755)
	err := p.Delete()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if p.IsExist() {
		t.Errorf("expected directory to be deleted")
	}
}

func TestIsAbs(t *testing.T) {
	p := New("/absolute/path")
	if !p.IsAbs() {
		t.Errorf("expected path to be absolute")
	}
}

func TestIsLocal(t *testing.T) {
	p := New("relative/path")
	if !p.IsLocal() {
		t.Errorf("expected path to be local")
	}
}

func TestIsValid(t *testing.T) {
	p := New("valid/path")
	if !p.IsValid() {
		t.Errorf("expected path to be valid")
	}
}

func TestIsRegular(t *testing.T) {
	p := New("testfile.txt")
	os.WriteFile(p.String(), []byte("test"), 0o644)
	if !p.IsRegular() {
		t.Errorf("expected path to be a regular file")
	}
	os.Remove(p.String())
}

func TestIsDir(t *testing.T) {
	p := New("testdir")
	os.Mkdir(p.String(), 0o755)
	if !p.IsDir() {
		t.Errorf("expected path to be a directory")
	}
	os.Remove(p.String())
}

func TestIsSymlink(t *testing.T) {
	p := New("testdir")
	if err := os.Mkdir(p.String(), 0o755); err != nil {
		t.Errorf("os.Mkdir: %v", err)
	}
	symlink := New("symlink")
	if err := os.Symlink(p.String(), symlink.String()); err != nil {
		t.Errorf("os.Symlink: %v", err)
	}
	if !symlink.IsSymlink() {
		t.Errorf("expected path to be a symlink")
	}
	os.Remove(p.String())
	os.Remove(symlink.String())
}

func TestIsDev(t *testing.T) {
	// This test is platform dependent and might not work on all systems.
	// It is generally difficult to create a device file in a cross-platform manner.
	// Therefore, this test is more of a placeholder to illustrate the usage.
	// On Unix-like systems, you might need root privileges to create a device file.
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on Windows systems")
	}

	// Example of a device file on Unix-like systems (this will not work on Windows)
	devPath := New("/dev/null")
	if !devPath.IsDev() {
		t.Errorf("expected /dev/null to be a device file")
	}
}

func TestIsExist(t *testing.T) {
	p := New("testfile.txt")
	if err := p.WriteFile([]byte("test")); err != nil {
		t.Errorf("WriteFile: %v", err)
	}
	if !p.IsExist() {
		t.Errorf("expected path to exist")
	}
}

func TestMatch(t *testing.T) {
	p := New("testfile.txt")
	if !p.Match("*.txt") {
		t.Errorf("expected path to match pattern")
	}
}

func TestVolumeName(t *testing.T) {
	p := New("C:\\path\\to\\file")
	expected := ""
	if runtime.GOOS == "windows" {
		expected = "C:"
	}
	if p.VolumeName() != expected {
		t.Errorf("expected %s, got %s", expected, p.VolumeName())
	}
}

func TestSize(t *testing.T) {
	p := New("testfile.txt")
	os.WriteFile(p.String(), []byte("test"), 0o644)
	expected := int64(4)
	size, err := p.Size()
	if err != nil || size != expected {
		t.Errorf("expected %d, got %d, error: %v", expected, size, err)
	}
	os.Remove(p.String())
}

func TestWalk(t *testing.T) {
	p := New("testdir")
	os.Mkdir(p.String(), 0o755)
	os.WriteFile(p.Join("file1.txt").String(), []byte("test"), 0o644)
	os.WriteFile(p.Join("file2.txt").String(), []byte("test"), 0o644)

	var files []string
	err := p.Walk(func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedFiles := []string{
		p.String(),
		p.Join("file1.txt").String(),
		p.Join("file2.txt").String(),
	}
	for _, ef := range expectedFiles {
		found := false
		for _, f := range files {
			if f == ef {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected file %s to be found", ef)
		}
	}

	os.RemoveAll(p.String())
}

func TestReadFile(t *testing.T) {
	p := New("testfile.txt")
	if err := p.WriteFile(testContent); err != nil {
		t.Errorf("WriteFile: %v", err)
	}

	content, err := p.ReadFile()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(content) != string(testContent) {
		t.Errorf("expected %s, got %s", testContent, content)
	}

	os.Remove(p.String())
}

func TestMkdirIfNotExist(t *testing.T) {
	p := New("testdir")

	// Ensure the directory does not exist before the test
	if p.IsExist() {
		p.Delete()
	}

	// Test creating a new directory
	err := p.MkdirIfNotExist()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !p.IsDir() {
		t.Errorf("expected path to be a directory")
	}

	// Test calling MkdirIfNotExist on an existing directory
	err = p.MkdirIfNotExist()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test calling MkdirIfNotExist on a path that exists but is not a directory
	filePath := p.Join("testfile.txt")
	err = filePath.WriteFile([]byte("test"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	err = filePath.MkdirIfNotExist()
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	// Clean up
	p.Delete()
}

func TestSizeX(t *testing.T) {
	p := New("testfile.txt")
	if err := p.WriteFile(testContent); err != nil {
		t.Errorf("WriteFile: %v", err)
	}
	defer p.Delete()

	expected := int64(len(testContent))
	size := p.SizeX()
	if size != expected {
		t.Errorf("expected %d, got %d", expected, size)
	}
}

func TestIsWritable(t *testing.T) {
	t.Run("NonExistentPath", func(t *testing.T) {
		p := New("nonexistentfile.txt")
		if p.IsWritable() {
			t.Errorf("expected path to be non-writable")
		}
	})

	t.Run("WritableFile", func(t *testing.T) {
		p := New("writablefile.txt")
		if err := p.WriteFile(testContent); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		defer p.Delete()

		if !p.IsWritable() {
			t.Errorf("expected path to be writable")
		}
	})

	t.Run("NonWritableFile", func(t *testing.T) {
		p := New("nonwritablefile.txt")
		if err := p.WriteFile(testContent); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		defer p.Delete()

		if err := os.Chmod(p.String(), 0o444); err != nil {
			t.Fatalf("Chmod: %v", err)
		}

		if p.IsWritable() {
			t.Errorf("expected path to be non-writable")
		}
	})

	t.Run("WritableDirectory", func(t *testing.T) {
		p := New("writabledir")
		if err := p.MkdirIfNotExist(); err != nil {
			t.Fatalf("MkdirIfNotExist: %v", err)
		}
		defer p.Delete()

		if !p.IsWritable() {
			t.Errorf("expected path to be writable")
		}
	})

	t.Run("NonWritableDirectory", func(t *testing.T) {
		p := New("nonwritabledir")
		if err := p.MkdirIfNotExist(); err != nil {
			t.Fatalf("MkdirIfNotExist: %v", err)
		}
		defer p.Delete()

		if err := os.Chmod(p.String(), 0o555); err != nil {
			t.Fatalf("Chmod: %v", err)
		}

		if p.IsWritable() {
			t.Errorf("expected path to be non-writable")
		}
	})
}

func TestOpenFile(t *testing.T) {
	p := New("testfile.txt")
	defer p.Delete()

	// Test creating a new file
	f, err := p.OpenFile(os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	f.Close()

	if !p.IsExist() {
		t.Errorf("expected file to be created")
	}

	// Test opening an existing file for reading and writing
	f, err = p.OpenFile(os.O_RDWR, 0o644)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	f.Close()

	// Test opening a non-existent file without create flag
	nonExistentFile := New("nonexistentfile.txt")
	_, err = nonExistentFile.OpenFile(os.O_RDWR, 0o644)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestJoinP(t *testing.T) {
	p := New("a", "b")
	p1 := New("c")
	p2 := New("d", "e")
	result := p.JoinPath(p1, p2)
	expected := filepath.Join("a", "b", "c", "d", "e")
	if result.String() != expected {
		t.Errorf("expected %s, got %s", expected, result.String())
	}

	// Test with no additional paths
	result = p.JoinPath()
	expected = filepath.Join("a", "b")
	if result.String() != expected {
		t.Errorf("expected %s, got %s", expected, result.String())
	}

	// Test with one additional path
	result = p.JoinPath(p1)
	expected = filepath.Join("a", "b", "c")
	if result.String() != expected {
		t.Errorf("expected %s, got %s", expected, result.String())
	}
}

func TestSourceFile(t *testing.T) {
	// Test that SourceFile returns the correct path of the current file
	expected := WD().Join("path_test.go").String()
	log.Println(ThisFile())
	sourceFile := ThisFile().String()
	if sourceFile != expected {
		t.Errorf("expected %s, got %s", expected, sourceFile)
	}
}

func TestWD(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}

	p := WD()
	if p.String() != wd {
		t.Errorf("expected %s, got %s", wd, p.String())
	}
}

func TestStringP(t *testing.T) {
	result := New("a", "b", "c").StringP()
	if result == nil {
		t.Errorf("expected non-nil pointer, got nil")
		return
	}
	if expected := filepath.Join("a", "b", "c"); *result != expected {
		t.Errorf("expected %s, got %s", expected, *result)
	}
}

func TestReadDir(t *testing.T) {
	// Test reading a directory with files
	dir := New("testdir")
	if err := dir.MkdirIfNotExist(); err != nil {
		t.Fatalf("MkdirIfNotExist: %v", err)
	}
	defer dir.Delete()

	file1 := dir.Join("file1.txt")
	if err := file1.WriteFile(testContent); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	file2 := dir.Join("file2.txt")
	if err := file2.WriteFile(testContent); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	entries, err := dir.ReadDir()
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}

	expectedEntries := map[string]bool{
		"file1.txt": true,
		"file2.txt": true,
	}

	for _, entry := range entries {
		if _, ok := expectedEntries[entry.Name()]; !ok {
			t.Errorf("unexpected entry: %s", entry.Name())
		}
		delete(expectedEntries, entry.Name())
	}

	if len(expectedEntries) != 0 {
		t.Errorf("missing entries: %v", expectedEntries)
	}

	// Test reading a non-directory path
	file := New("testfile.txt")
	if err := file.WriteFile(testContent); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	defer file.Delete()

	_, err = file.ReadDir()
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("expected 'not a directory' error, got %v", err)
	}
}

func TestBaseWithoutExt(t *testing.T) {
	tests := []struct {
		input    Path
		expected string
	}{
		{New("."), "."},
		{New("a", "b", "c.txt"), "c"},
		{New("a", "b", "c.tar.gz"), "c.tar"},
		{New("a", "b", "c"), "c"},
		{New("a", "b", ".hiddenfile"), ".hiddenfile"},
		{New("a", "b", "c.d.e.f"), "c.d.e"},
	}

	for _, test := range tests {
		result := test.input.BaseWithoutExt()
		if result.String() != test.expected {
			t.Errorf("expected %s, got %s", test.expected, result.String())
		}
	}
}

func TestIsChildOf(t *testing.T) {
	tests := []struct {
		child    Path
		parent   Path
		expected bool
	}{
		{New("a/b/c"), New("a/b"), true},
		{New("a/b/c"), New("a/b/c"), true},
		{New("a/b/c"), New("a/b/c/d"), false},
		{New("a/b/c"), New("a/b/x"), false},
		{New("/a/b/c"), New("/a/b"), true},
		{New("/a/b/c"), New("/a/b/c"), true},
		{New("/a/b/c"), New("/a/b/c/d"), false},
		{New("/a/b/c"), New("/a/b/x"), false},
	}

	for _, test := range tests {
		result := test.child.IsChildOf(test.parent)
		if result != test.expected {
			t.Errorf("expected %v, got %v for child %s and parent %s", test.expected, result, test.child, test.parent)
		}
	}
}

func TestIsParentOf(t *testing.T) {
	tests := []struct {
		parent   Path
		child    Path
		expected bool
	}{
		{New("a/b"), New("a/b/c"), true},
		{New("a/b/c"), New("a/b/c"), true},
		{New("a/b/c/d"), New("a/b/c"), false},
		{New("a/b/x"), New("a/b/c"), false},
		{New("/a/b"), New("/a/b/c"), true},
		{New("/a/b/c"), New("/a/b/c"), true},
		{New("/a/b/c/d"), New("/a/b/c"), false},
		{New("/a/b/x"), New("/a/b/c"), false},
	}

	for _, test := range tests {
		result := test.parent.IsParentOf(test.child)
		if result != test.expected {
			t.Errorf("expected %v, got %v for parent %s and child %s", test.expected, result, test.parent, test.child)
		}
	}
}

func TestThisDir(t *testing.T) {
	// Test that ThisDir returns the directory of the current file
	expected := WD().String()
	thisDir := ThisDir().String()
	if thisDir != expected {
		t.Errorf("expected %s, got %s", expected, thisDir)
	}

	// Test that ThisDir returns the correct directory when called from another function
	func() {
		expected := WD().String()
		thisDir := ThisDir().String()
		if thisDir != expected {
			t.Errorf("expected %s, got %s", expected, thisDir)
		}
	}()
}

func TestS(t *testing.T) {
	tests := []struct {
		input    Path
		expected string
	}{
		{New("a", "b", "c"), filepath.Join("a", "b", "c")},
		{New(""), ""},
		{New("a"), "a"},
	}

	for _, test := range tests {
		result := test.input.Str()
		if result != test.expected {
			t.Errorf("expected %s, got %s", test.expected, result)
		}
	}
}

func TestRemove(t *testing.T) {
	// Test removing a file
	t.Run("RemoveFile", func(t *testing.T) {
		p := New("testfile.txt")
		if err := p.WriteFile(testContent); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		defer p.Delete()

		if err := p.Remove(); err != nil {
			t.Errorf("Remove: %v", err)
		}

		if p.IsExist() {
			t.Errorf("expected file to be removed")
		}
	})

	// Test removing a directory
	t.Run("RemoveDirectory", func(t *testing.T) {
		p := New("testdir")
		if err := p.MkdirIfNotExist(); err != nil {
			t.Fatalf("MkdirIfNotExist: %v", err)
		}
		defer p.Delete()

		if err := p.Remove(); err != nil {
			t.Errorf("Remove: %v", err)
		}

		if p.IsExist() {
			t.Errorf("expected directory to be removed")
		}
	})
}

func TestRename(t *testing.T) {
	t.Run("RenameFile", func(t *testing.T) {
		src := New("srcfile.txt")
		dst := New("dstfile.txt")
		if err := src.WriteFile(testContent); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		defer src.Delete()
		defer dst.Delete()

		if err := src.Rename(dst.String()); err != nil {
			t.Fatalf("Rename: %v", err)
		}

		if src.IsExist() {
			t.Errorf("expected source file to be renamed")
		}
		if !dst.IsExist() {
			t.Errorf("expected destination file to exist")
		}

		dstContent, err := dst.ReadFile()
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if string(dstContent) != string(testContent) {
			t.Errorf("expected %s, got %s", testContent, dstContent)
		}
	})

	t.Run("RenameFileToDirectory", func(t *testing.T) {
		src := New("srcfile.txt")
		dstDir := New("dstdir")
		dst := dstDir.Join("srcfile.txt")
		if err := src.WriteFile(testContent); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		defer src.Delete()
		defer dstDir.Delete()

		if err := src.Rename(dst.String()); err != nil {
			t.Fatalf("Rename: %v", err)
		}

		if src.IsExist() {
			t.Errorf("expected source file to be renamed")
		}
		if !dst.IsExist() {
			t.Errorf("expected destination file to exist")
		}

		dstContent, err := dst.ReadFile()
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if string(dstContent) != string(testContent) {
			t.Errorf("expected %s, got %s", testContent, dstContent)
		}
	})

	t.Run("RenameDirectory", func(t *testing.T) {
		srcDir := New("srcdir")
		dstDir := New("dstdir")
		if err := srcDir.MkdirIfNotExist(); err != nil {
			t.Fatalf("MkdirIfNotExist: %v", err)
		}
		defer srcDir.Delete()
		defer dstDir.Delete()

		srcFile := srcDir.Join("file.txt")
		if err := srcFile.WriteFile(testContent); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		if err := srcDir.Rename(dstDir.String()); err != nil {
			t.Fatalf("Rename: %v", err)
		}

		if srcDir.IsExist() {
			t.Errorf("expected source directory to be renamed")
		}
		if !dstDir.IsExist() {
			t.Errorf("expected destination directory to exist")
		}

		dstFile := dstDir.Join("file.txt")
		dstContent, err := dstFile.ReadFile()
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if string(dstContent) != string(testContent) {
			t.Errorf("expected %s, got %s", testContent, dstContent)
		}
	})
}

func TestCopy(t *testing.T) {
	t.Run("CopyFile", func(t *testing.T) {
		src := New("srcfile.txt")
		dst := New("dstfile.txt")
		if err := src.WriteFile(testContent); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		defer src.Delete()
		defer dst.Delete()

		if err := src.Copy(dst); err != nil {
			t.Fatalf("Copy: %v", err)
		}

		if !dst.IsExist() {
			t.Errorf("expected destination file to exist")
		}

		dstContent, err := dst.ReadFile()
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if string(dstContent) != string(testContent) {
			t.Errorf("expected %s, got %s", testContent, dstContent)
		}
	})

	t.Run("CopyFileToDirectory", func(t *testing.T) {
		src := New("srcfile.txt")
		if err := src.WriteFile(testContent); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		defer src.Delete()

		dstDir := New("dstdir")
		defer dstDir.Delete()
		dst := dstDir.Join("srcfile.txt")
		if err := src.Copy(dst); err != nil {
			t.Fatalf("Copy: %v", err)
		}

		if !dst.IsExist() {
			t.Errorf("expected destination file to exist")
		}

		dstContent, err := dst.ReadFile()
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if string(dstContent) != string(testContent) {
			t.Errorf("expected %s, got %s", testContent, dstContent)
		}
	})

	t.Run("CopyDirectory", func(t *testing.T) {
		srcDir := New("srcdir")
		dstDir := New("dstdir")
		if err := srcDir.MkdirIfNotExist(); err != nil {
			t.Fatalf("MkdirIfNotExist: %v", err)
		}
		defer srcDir.Delete()
		defer dstDir.Delete()

		srcFile := srcDir.Join("file.txt")
		if err := srcFile.WriteFile(testContent); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		if err := srcDir.Copy(dstDir); err != nil {
			t.Fatalf("Copy: %v", err)
		}

		if !dstDir.IsExist() {
			t.Errorf("expected destination directory to exist")
		}

		dstFile := dstDir.Join("file.txt")
		dstContent, err := dstFile.ReadFile()
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if string(dstContent) != string(testContent) {
			t.Errorf("expected %s, got %s", testContent, dstContent)
		}
	})
}

func TestMove(t *testing.T) {
	t.Run("MoveFile", func(t *testing.T) {
		src := New("srcfile.txt")
		dst := New("dstfile.txt")
		if err := src.WriteFile(testContent); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		defer src.Delete()
		defer dst.Delete()

		if err := src.Move(dst); err != nil {
			t.Fatalf("Move: %v", err)
		}

		if src.IsExist() {
			t.Errorf("expected source file to be moved")
		}
		if !dst.IsExist() {
			t.Errorf("expected destination file to exist")
		}

		dstContent, err := dst.ReadFile()
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if string(dstContent) != string(testContent) {
			t.Errorf("expected %s, got %s", testContent, dstContent)
		}
	})

	t.Run("MoveFileToDirectory", func(t *testing.T) {
		src := New("srcfile.txt")
		dstDir := New("dstdir")
		dst := dstDir.Join("srcfile.txt")
		if err := src.WriteFile(testContent); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		defer src.Delete()
		defer dstDir.Delete()

		if err := src.Move(dst); err != nil {
			t.Fatalf("Move: %v", err)
		}

		if src.IsExist() {
			t.Errorf("expected source file to be moved")
		}
		if !dst.IsExist() {
			t.Errorf("expected destination file to exist")
		}

		dstContent, err := dst.ReadFile()
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if string(dstContent) != string(testContent) {
			t.Errorf("expected %s, got %s", testContent, dstContent)
		}
	})

	t.Run("MoveDirectory", func(t *testing.T) {
		srcDir := New("srcdir")
		dstDir := New("dstdir")
		if err := srcDir.MkdirIfNotExist(); err != nil {
			t.Fatalf("MkdirIfNotExist: %v", err)
		}
		defer srcDir.Delete()
		defer dstDir.Delete()

		srcFile := srcDir.Join("file.txt")
		if err := srcFile.WriteFile(testContent); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		if err := srcDir.Move(dstDir); err != nil {
			t.Fatalf("Move: %v", err)
		}

		if srcDir.IsExist() {
			t.Errorf("expected source directory to be moved")
		}
		if !dstDir.IsExist() {
			t.Errorf("expected destination directory to exist")
		}

		dstFile := dstDir.Join("file.txt")
		dstContent, err := dstFile.ReadFile()
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if string(dstContent) != string(testContent) {
			t.Errorf("expected %s, got %s", testContent, dstContent)
		}
	})

	t.Run("MoveNonExistentFile", func(t *testing.T) {
		src := New("nonexistentfile.txt")
		dst := New("dstfile.txt")

		err := src.Move(dst)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "source file does not exist") {
			t.Errorf("expected 'source file does not exist' error, got %v", err)
		}
	})
}

func TestOpenOrCreate(t *testing.T) {
	p := New("testfile.txt")
	defer p.Delete()

	// Test creating a new file
	f, err := p.OpenOrCreate()
	if err != nil {
		t.Fatalf("OpenOrCreate: %v", err)
	}
	f.Close()

	if !p.IsExist() {
		t.Errorf("expected file to be created")
	}

	// Test opening an existing file for reading and writing
	f, err = p.OpenOrCreate()
	if err != nil {
		t.Fatalf("OpenOrCreate: %v", err)
	}
	f.Close()

	// Test writing to the file
	content := []byte("test content")
	if err := p.WriteFile(content); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Test reading from the file
	readContent, err := p.ReadFile()
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(readContent) != string(content) {
		t.Errorf("expected %s, got %s", content, readContent)
	}
}

func TestCreate(t *testing.T) {
	t.Run("CreateNewFile", func(t *testing.T) {
		p := New("testfile.txt")
		defer p.Delete()

		f, err := p.Create()
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		f.Close()

		if !p.IsExist() {
			t.Errorf("expected file to be created")
		}
	})

	t.Run("CreateExistingFile", func(t *testing.T) {
		p := New("testfile.txt")
		defer p.Delete()

		// Create the file first
		f, err := p.Create()
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		f.Close()

		// Try to create the file again
		_, err = p.Create()
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("expected 'already exists' error, got %v", err)
		}
	})

	t.Run("CreateInNonExistentDirectory", func(t *testing.T) {
		p := New("nonexistentdir", "testfile.txt")
		defer p.Dir().Delete()

		f, err := p.Create()
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		f.Close()

		if !p.IsExist() {
			t.Errorf("expected file to be created")
		}
		if !p.Dir().IsExist() {
			t.Errorf("expected parent directory to be created")
		}
	})
}

func TestUsage(t *testing.T) {
	// Test Usage on an existing directory
	t.Run("ExistingDirectory", func(t *testing.T) {
		p := New(".")
		usage, err := p.Usage()
		if err != nil {
			t.Fatalf("Usage: %v", err)
		}
		if usage.Total == 0 {
			t.Errorf("expected non-zero total usage")
		}
		if usage.Free == 0 {
			t.Errorf("expected non-zero free usage")
		}
		if usage.Used == 0 {
			t.Errorf("expected non-zero used usage")
		}
		if usage.UsedPercent == 0 {
			t.Errorf("expected non-zero used percent")
		}
	})

	// Test Usage on a non-existent path
	t.Run("NonExistentPath", func(t *testing.T) {
		p := New("nonexistentpath")
		_, err := p.Usage()
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestSHA256String(t *testing.T) {
	p := New("testfile.txt")
	defer p.Delete()

	// Write test content to the file
	if err := p.WriteFile(testContent); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Calculate the expected SHA256 hash
	expectedHash := sha256.Sum256(testContent)
	expectedHashStr := hex.EncodeToString(expectedHash[:])

	// Get the SHA256 hash from the method
	hashStr := p.SHA256String()

	if hashStr != expectedHashStr {
		t.Errorf("expected %s, got %s", expectedHashStr, hashStr)
	}
}

func TestSHA1String(t *testing.T) {
	p := New("testfile.txt")
	defer p.Delete()

	// Write test content to the file
	if err := p.WriteFile(testContent); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Calculate the expected SHA1 hash
	expectedHash := sha1.Sum(testContent)
	expectedHashStr := hex.EncodeToString(expectedHash[:])

	// Get the SHA1 hash from the method
	hashStr := p.SHA1String()

	if hashStr != expectedHashStr {
		t.Errorf("expected %s, got %s", expectedHashStr, hashStr)
	}
}

func TestMD5String(t *testing.T) {
	p := New("testfile.txt")
	defer p.Delete()

	// Write test content to the file
	if err := p.WriteFile(testContent); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Calculate the expected MD5 hash
	expectedHash := md5.Sum(testContent)
	expectedHashStr := hex.EncodeToString(expectedHash[:])

	// Get the MD5 hash from the method
	hashStr := p.MD5String()

	if hashStr != expectedHashStr {
		t.Errorf("expected %s, got %s", expectedHashStr, hashStr)
	}
}

func TestQueryHas(t *testing.T) {
	tests := []struct {
		path     Path
		key      string
		expected bool
	}{
		{New("/example/path/for/test?foo=bar"), "foo", true},
		{New("/example/path/for/test?foo=bar&baz=qux"), "baz", true},
		{New("/example/path/for/test?foo=bar&baz=qux"), "quux", false},
		{New("/example/path/for/test"), "foo", false},
		{New("/example/path/for/test?foo="), "foo", true},
	}

	for _, test := range tests {
		result := test.path.QueryHas(test.key)
		if result != test.expected {
			t.Errorf("expected %v, got %v for path %s and key %s", test.expected, result, test.path, test.key)
		}
	}
}

func TestQueryDel(t *testing.T) {
	tests := []struct {
		path     Path
		key      string
		expected string
	}{
		{New("/example/path/for/test?foo=bar"), "foo", "/example/path/for/test"},
		{New("/example/path/for/test?foo=bar&baz=qux"), "foo", "/example/path/for/test?baz=qux"},
		{New("/example/path/for/test?foo=bar&baz=qux"), "baz", "/example/path/for/test?foo=bar"},
		{New("/example/path/for/test?foo=bar&baz=qux"), "quux", "/example/path/for/test?baz=qux&foo=bar"},
		{New("/example/path/for/test"), "foo", "/example/path/for/test"},
		{New("/example/path/for/test?foo="), "foo", "/example/path/for/test"},
	}

	for _, test := range tests {
		result := test.path.QueryDel(test.key)
		if result.String() != test.expected {
			t.Errorf("expected %s, got %s for path %s and key %s", test.expected, result.String(), test.path, test.key)
		}
	}
}

func TestQueryAdd(t *testing.T) {
	tests := []struct {
		path     Path
		key      string
		value    any
		expected string
	}{
		{New("/example/path/for/test"), "foo", "bar", "/example/path/for/test?foo=bar"},
		{New("/example/path/for/test?foo=bar"), "baz", "qux", "/example/path/for/test?baz=qux&foo=bar"},
		{New("/example/path/for/test?foo=bar"), "foo", "baz", "/example/path/for/test?foo=bar&foo=baz"},
		{New("/example/path/for/test"), "foo", 123, "/example/path/for/test?foo=123"},
		{New("/example/path/for/test?foo=bar"), "baz", true, "/example/path/for/test?baz=true&foo=bar"},
	}

	for _, test := range tests {
		result := test.path.QueryAdd(test.key, test.value)
		if result.String() != test.expected {
			t.Errorf("expected %s, got %s for path %s, key %s, and value %v", test.expected, result.String(), test.path, test.key, test.value)
		}
	}
}

func TestQuerySet(t *testing.T) {
	tests := []struct {
		path     Path
		key      string
		value    any
		expected string
	}{
		{New("/example/path/for/test"), "foo", "bar", "/example/path/for/test?foo=bar"},
		{New("/example/path/for/test?foo=bar"), "baz", "qux", "/example/path/for/test?baz=qux&foo=bar"},
		{New("/example/path/for/test?foo=bar"), "foo", "baz", "/example/path/for/test?foo=baz"},
		{New("/example/path/for/test"), "foo", 123, "/example/path/for/test?foo=123"},
		{New("/example/path/for/test?foo=bar"), "baz", true, "/example/path/for/test?baz=true&foo=bar"},
	}

	for _, test := range tests {
		result := test.path.QuerySet(test.key, test.value)
		if result.String() != test.expected {
			t.Errorf("expected %s, got %s for path %s, key %s, and value %v", test.expected, result.String(), test.path, test.key, test.value)
		}
	}
}

func TestQuery(t *testing.T) {
	tests := []struct {
		path     Path
		expected string
	}{
		{New("/example/path/for/test?foo=bar"), "foo=bar"},
		{New("/example/path/for/test?foo=bar&baz=qux"), "foo=bar&baz=qux"},
		{New("/example/path/for/test"), ""},
		{New("/example/path/for/test?"), ""},
		{New("/example/path/for/test?foo="), "foo="},
	}

	for _, test := range tests {
		result := test.path.Query()
		if result != test.expected {
			t.Errorf("expected %s, got %s for path %s", test.expected, result, test.path)
		}
	}
}

func TestWithQuery(t *testing.T) {
	tests := []struct {
		path     Path
		query    string
		expected string
	}{
		{New("/example/path/for/test"), "foo=bar", "/example/path/for/test?foo=bar"},
		{New("/example/path/for/test?existing=query"), "foo=bar", "/example/path/for/test?foo=bar"},
		{New("/example/path/for/test"), "", "/example/path/for/test"},
		{New("/example/path/for/test?existing=query"), "", "/example/path/for/test"},
		{New("/example/path/for/test"), "foo=bar&baz=qux", "/example/path/for/test?foo=bar&baz=qux"},
	}

	for _, test := range tests {
		result := test.path.WithQuery(test.query)
		if result.String() != test.expected {
			t.Errorf("expected %s, got %s for path %s and query %s", test.expected, result.String(), test.path, test.query)
		}
	}
}

func TestWithoutQuery(t *testing.T) {
	tests := []struct {
		path     Path
		expected string
	}{
		{New("/example/path/for/test?foo=bar"), "/example/path/for/test"},
		{New("/example/path/for/test?foo=bar&baz=qux"), "/example/path/for/test"},
		{New("/example/path/for/test"), "/example/path/for/test"},
		{New("/example/path/for/test?"), "/example/path/for/test"},
		{New("/example/path/for/test?foo="), "/example/path/for/test"},
	}

	for _, test := range tests {
		result := test.path.WithoutQuery()
		if result.String() != test.expected {
			t.Errorf("expected %s, got %s for path %s", test.expected, result.String(), test.path)
		}
	}
}

func TestHasQuery(t *testing.T) {
	tests := []struct {
		path     Path
		expected bool
	}{
		{New("/example/path/for/test?foo=bar"), true},
		{New("/example/path/for/test?foo=bar&baz=qux"), true},
		{New("/example/path/for/test"), false},
		{New("/example/path/for/test?"), true},
		{New("/example/path/for/test?foo="), true},
	}

	for _, test := range tests {
		result := test.path.HasQuery()
		if result != test.expected {
			t.Errorf("expected %v, got %v for path %s", test.expected, result, test.path)
		}
	}
}

func TestMergeMove_SourceDoesNotExist(t *testing.T) {
	src := New("nonexistent.txt")
	dst := New("dst.txt")
	err := src.MergeMove(dst)
	if err == nil {
		t.Fatal("expected error for non-existent source, got nil")
	}
}

func TestMergeMove_MoveFileToNonExistentDst(t *testing.T) {
	// Create a temporary file as source
	tempDir := t.TempDir()
	srcPath := New(filepath.Join(tempDir, "src.txt"))
	dstPath := New(filepath.Join(tempDir, "dst.txt"))

	// Write some content to the source file
	if err := srcPath.WriteFile([]byte("merge move test")); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Ensure destination does not exist
	os.Remove(dstPath.String())

	// Call MergeMove, should perform a rename
	if err := srcPath.MergeMove(dstPath); err != nil {
		t.Fatalf("MergeMove: %v", err)
	}

	// Source should no longer exist and destination file should now exist
	if srcPath.Exists() {
		t.Errorf("expected source file to be moved (non-existent)")
	}
	if !dstPath.Exists() {
		t.Errorf("expected destination file to exist")
	}
}

func TestMergeMove_MoveFileToExistingDirectory(t *testing.T) {
	// Create a temporary file as source and a directory as destination
	tempDir := t.TempDir()
	srcPath := New(filepath.Join(tempDir, "src.txt"))
	dstDir := New(filepath.Join(tempDir, "destDir"))

	// Write content to the source file
	if err := srcPath.WriteFile([]byte("file to dir")); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	// Create the destination directory
	if err := dstDir.MkdirIfNotExist(); err != nil {
		t.Fatalf("MkdirIfNotExist: %v", err)
	}

	// Call MergeMove; it should move src file inside dst directory
	if err := srcPath.MergeMove(dstDir); err != nil {
		t.Fatalf("MergeMove: %v", err)
	}

	// The moved file should now be at dstDir joined with base name of srcPath.
	movedFile := dstDir.JoinPath(srcPath.Base())
	if srcPath.Exists() {
		t.Errorf("expected source file to be moved")
	}
	if !movedFile.Exists() {
		t.Errorf("expected moved file (%s) to exist", movedFile.String())
	}
}

func TestMergeMove_MoveFileToExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	// Create a source and destination file paths
	srcPath := New(filepath.Join(tempDir, "src.txt"))
	dstPath := New(filepath.Join(tempDir, "dst.txt"))

	// Write different contents to source and destination
	srcContent := []byte("source content")
	dstContent := []byte("old destination")

	if err := srcPath.WriteFile(srcContent); err != nil {
		t.Fatalf("WriteFile src: %v", err)
	}
	if err := dstPath.WriteFile(dstContent); err != nil {
		t.Fatalf("WriteFile dst: %v", err)
	}

	// Call MergeMove: it should delete the destination and rename the source to dstPath.
	if err := srcPath.MergeMove(dstPath); err != nil {
		t.Fatalf("MergeMove: %v", err)
	}

	// Source should not exist, and destination should have source content.
	if srcPath.Exists() {
		t.Errorf("expected source file to be moved")
	}
	if !dstPath.Exists() {
		t.Errorf("expected destination file to exist")
	}
	result, err := dstPath.ReadFile()
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(result) != string(srcContent) {
		t.Errorf("expected destination file content %q, got %q", srcContent, result)
	}
}

func TestMergeMove_MergeMoveDirectory(t *testing.T) {
	tempDir := t.TempDir()
	// Create source directory with multiple files
	srcDir := New(filepath.Join(tempDir, "srcDir"))
	if err := srcDir.MkdirIfNotExist(); err != nil {
		t.Fatalf("MkdirIfNotExist srcDir: %v", err)
	}

	// Create a couple of files inside source directory
	file1 := srcDir.Join("file1.txt")
	file2 := srcDir.Join("file2.txt")
	if err := file1.WriteFile([]byte("content1")); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := file2.WriteFile([]byte("content2")); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Create destination directory where the merge will occur; it already exists
	dstDir := New(filepath.Join(tempDir, "dstDir"))
	if err := dstDir.MkdirIfNotExist(); err != nil {
		t.Fatalf("MkdirIfNotExist dstDir: %v", err)
	}

	// MergeMove srcDir to dstDir; expect the files to be moved into dstDir
	if err := srcDir.MergeMove(dstDir); err != nil {
		t.Fatalf("MergeMove: %v", err)
	}

	// Source directory should be deleted.
	if srcDir.Exists() {
		t.Errorf("expected source directory to be deleted")
	}

	// Files should now exist in dstDir.
	movedFile1 := dstDir.Join("file1.txt")
	movedFile2 := dstDir.Join("file2.txt")
	if !movedFile1.Exists() {
		t.Errorf("expected file1.txt to exist in destination")
	}
	if !movedFile2.Exists() {
		t.Errorf("expected file2.txt to exist in destination")
	}

	// Verify contents
	content1, _ := movedFile1.ReadFile()
	if string(content1) != "content1" {
		t.Errorf("expected content1, got %s", content1)
	}
	content2, _ := movedFile2.ReadFile()
	if string(content2) != "content2" {
		t.Errorf("expected content2, got %s", content2)
	}
}

func TestMergeMove_DirectoryToNonDirectory(t *testing.T) {
	// Create source directory with one file
	tempDir := t.TempDir()
	srcDir := New(filepath.Join(tempDir, "srcDir"))
	if err := srcDir.MkdirIfNotExist(); err != nil {
		t.Fatalf("MkdirIfNotExist: %v", err)
	}
	srcFile := srcDir.Join("file.txt")
	if err := srcFile.WriteFile([]byte("content")); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Create a destination regular file
	dstFile := New(filepath.Join(tempDir, "dstFile.txt"))
	if err := dstFile.WriteFile([]byte("dst content")); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Attempting to merge move a directory into a non-directory should error.
	err := srcDir.MergeMove(dstFile)
	if err == nil {
		t.Fatal("expected error when merging directory to non-directory, got nil")
	}
}

func TestReadJSON(t *testing.T) {
	t.Run("ReadValidJSONObject", func(t *testing.T) {
		p := New("testfile.json")
		defer p.Delete()

		jsonData := map[string]any{
			"name": "test",
			"age":  30,
		}
		if err := p.WriteJSON(jsonData); err != nil {
			t.Fatalf("WriteJSON: %v", err)
		}

		result, err := p.ReadJSON()
		if err != nil {
			t.Fatalf("ReadJSON: %v", err)
		}

		resultMap, ok := result.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", result)
		}

		if resultMap["name"] != "test" {
			t.Errorf("expected name=test, got %v", resultMap["name"])
		}
		if resultMap["age"].(float64) != 30 {
			t.Errorf("expected age=30, got %v", resultMap["age"])
		}
	})

	t.Run("ReadValidJSONArray", func(t *testing.T) {
		p := New("testfile.json")
		defer p.Delete()

		jsonData := []any{"item1", "item2", "item3"}
		if err := p.WriteJSON(jsonData); err != nil {
			t.Fatalf("WriteJSON: %v", err)
		}

		result, err := p.ReadJSON()
		if err != nil {
			t.Fatalf("ReadJSON: %v", err)
		}

		resultArray, ok := result.([]any)
		if !ok {
			t.Fatalf("expected []any, got %T", result)
		}

		if len(resultArray) != 3 {
			t.Errorf("expected 3 items, got %d", len(resultArray))
		}
	})

	t.Run("ReadInvalidJSON", func(t *testing.T) {
		p := New("testfile.json")
		defer p.Delete()

		if err := p.WriteFile([]byte("{invalid json}")); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		_, err := p.ReadJSON()
		if err == nil {
			t.Errorf("expected error for invalid JSON, got nil")
		}
	})

	t.Run("ReadNonExistentFile", func(t *testing.T) {
		p := New("nonexistent.json")

		_, err := p.ReadJSON()
		if err == nil {
			t.Errorf("expected error for non-existent file, got nil")
		}
	})

	t.Run("ReadEmptyJSON", func(t *testing.T) {
		p := New("testfile.json")
		defer p.Delete()

		if err := p.WriteFile([]byte("")); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		_, err := p.ReadJSON()
		if err == nil {
			t.Errorf("expected error for empty JSON, got nil")
		}
	})
}

func TestWriteJSON(t *testing.T) {
	t.Run("WriteJSONObject", func(t *testing.T) {
		p := New("testfile.json")
		defer p.Delete()

		jsonData := map[string]any{
			"name": "test",
			"age":  30,
		}

		if err := p.WriteJSON(jsonData); err != nil {
			t.Fatalf("WriteJSON: %v", err)
		}

		if !p.IsExist() {
			t.Errorf("expected file to exist")
		}

		content, err := p.ReadFile()
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}

		if !strings.Contains(string(content), "name") {
			t.Errorf("expected JSON to contain 'name'")
		}
	})

	t.Run("WriteJSONToNonExistentDirectory", func(t *testing.T) {
		p := New("nonexistentdir", "testfile.json")
		defer p.Dir().Delete()

		jsonData := map[string]string{"key": "value"}
		if err := p.WriteJSON(jsonData); err != nil {
			t.Fatalf("WriteJSON: %v", err)
		}

		if !p.IsExist() {
			t.Errorf("expected file to exist")
		}
	})
}

func TestReadFrom(t *testing.T) {
	t.Run("ReadFromReader", func(t *testing.T) {
		p := New("testfile.txt")
		defer p.Delete()

		content := "test content from reader"
		reader := strings.NewReader(content)

		n, err := p.ReadFrom(reader)
		if err != nil {
			t.Fatalf("ReadFrom: %v", err)
		}

		if n != int64(len(content)) {
			t.Errorf("expected %d bytes written, got %d", len(content), n)
		}

		readContent, err := p.ReadFile()
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}

		if string(readContent) != content {
			t.Errorf("expected %s, got %s", content, readContent)
		}
	})
}

func TestReadFromPath(t *testing.T) {
	t.Run("ReadFromExistingFile", func(t *testing.T) {
		src := New("srcfile.txt")
		dst := New("dstfile.txt")
		defer src.Delete()
		defer dst.Delete()

		if err := src.WriteFile(testContent); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		n, err := dst.ReadFromPath(src)
		if err != nil {
			t.Fatalf("ReadFromPath: %v", err)
		}

		if n != int64(len(testContent)) {
			t.Errorf("expected %d bytes, got %d", len(testContent), n)
		}

		dstContent, err := dst.ReadFile()
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}

		if string(dstContent) != string(testContent) {
			t.Errorf("expected %s, got %s", testContent, dstContent)
		}
	})
}

func TestWriteTo(t *testing.T) {
	t.Run("WriteToWriter", func(t *testing.T) {
		p := New("testfile.txt")
		defer p.Delete()

		if err := p.WriteFile(testContent); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		var buf strings.Builder
		n, err := p.WriteTo(&buf)
		if err != nil {
			t.Fatalf("WriteTo: %v", err)
		}

		if n != int64(len(testContent)) {
			t.Errorf("expected %d bytes, got %d", len(testContent), n)
		}

		if buf.String() != string(testContent) {
			t.Errorf("expected %s, got %s", testContent, buf.String())
		}
	})
}

func TestWriteToPath(t *testing.T) {
	t.Run("WriteToNewFile", func(t *testing.T) {
		src := New("srcfile.txt")
		dst := New("dstfile.txt")
		defer src.Delete()
		defer dst.Delete()

		if err := src.WriteFile(testContent); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		if err := src.WriteToPath(dst); err != nil {
			t.Fatalf("WriteToPath: %v", err)
		}

		dstContent, err := dst.ReadFile()
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}

		if string(dstContent) != string(testContent) {
			t.Errorf("expected %s, got %s", testContent, dstContent)
		}
	})
}

func TestTruncate(t *testing.T) {
	t.Run("TruncateFile", func(t *testing.T) {
		p := New("testfile.txt")
		defer p.Delete()

		if err := p.WriteFile(testContent); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		if err := p.Truncate(); err != nil {
			t.Fatalf("Truncate: %v", err)
		}

		size, err := p.Size()
		if err != nil {
			t.Fatalf("Size: %v", err)
		}

		if size != 0 {
			t.Errorf("expected size 0, got %d", size)
		}
	})

	t.Run("TruncateDirectory", func(t *testing.T) {
		dir := New("testdir")
		defer dir.Delete()

		if err := dir.MkdirIfNotExist(); err != nil {
			t.Fatalf("MkdirIfNotExist: %v", err)
		}

		file := dir.Join("file.txt")
		if err := file.WriteFile(testContent); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		if err := dir.Truncate(); err != nil {
			t.Fatalf("Truncate: %v", err)
		}

		if !dir.IsEmpty() {
			t.Errorf("expected directory to be empty")
		}
	})
}

func TestIsEmpty(t *testing.T) {
	t.Run("EmptyFile", func(t *testing.T) {
		p := New("testfile.txt")
		defer p.Delete()

		if err := p.WriteFile([]byte("")); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		if !p.IsEmpty() {
			t.Errorf("expected file to be empty")
		}
	})

	t.Run("NonEmptyFile", func(t *testing.T) {
		p := New("testfile.txt")
		defer p.Delete()

		if err := p.WriteFile(testContent); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		if p.IsEmpty() {
			t.Errorf("expected file to be non-empty")
		}
	})

	t.Run("EmptyDirectory", func(t *testing.T) {
		dir := New("testdir")
		defer dir.Delete()

		if err := dir.MkdirIfNotExist(); err != nil {
			t.Fatalf("MkdirIfNotExist: %v", err)
		}

		if !dir.IsEmpty() {
			t.Errorf("expected directory to be empty")
		}
	})

	t.Run("NonEmptyDirectory", func(t *testing.T) {
		dir := New("testdir")
		defer dir.Delete()

		if err := dir.MkdirIfNotExist(); err != nil {
			t.Fatalf("MkdirIfNotExist: %v", err)
		}

		file := dir.Join("file.txt")
		if err := file.WriteFile(testContent); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		if dir.IsEmpty() {
			t.Errorf("expected directory to be non-empty")
		}
	})

	t.Run("NonExistentPath", func(t *testing.T) {
		p := New("nonexistent.txt")
		if !p.IsEmpty() {
			t.Errorf("expected non-existent path to be considered empty")
		}
	})
}

func TestIsEqual(t *testing.T) {
	t.Run("SamePaths", func(t *testing.T) {
		p1 := New("test.txt")
		p2 := New("test.txt")

		if !p1.IsEqual(p2) {
			t.Errorf("expected paths to be equal")
		}
	})

	t.Run("DifferentPaths", func(t *testing.T) {
		p1 := New("test1.txt")
		p2 := New("test2.txt")

		if p1.IsEqual(p2) {
			t.Errorf("expected paths to be different")
		}
	})

	t.Run("RelativeAndAbsoluteSame", func(t *testing.T) {
		p1 := New("test.txt")
		abs, _ := p1.Abs()

		if !p1.IsEqual(abs) {
			t.Errorf("expected relative and absolute paths to be equal")
		}
	})
}

func TestHasPrefix(t *testing.T) {
	tests := []struct {
		path     Path
		prefix   string
		expected bool
	}{
		{New("/home/user/file.txt"), "/home", true},
		{New("/home/user/file.txt"), "/var", false},
		{New("relative/path"), "relative", true},
		{New("relative/path"), "absolute", false},
	}

	for _, test := range tests {
		result := test.path.HasPrefix(test.prefix)
		if result != test.expected {
			t.Errorf("expected %v for path %s with prefix %s, got %v", test.expected, test.path, test.prefix, result)
		}
	}
}

func TestHasSuffix(t *testing.T) {
	tests := []struct {
		path     Path
		suffix   string
		expected bool
	}{
		{New("/home/user/file.txt"), ".txt", true},
		{New("/home/user/file.txt"), ".pdf", false},
		{New("document.pdf"), ".pdf", true},
		{New("document"), ".pdf", false},
	}

	for _, test := range tests {
		result := test.path.HasSuffix(test.suffix)
		if result != test.expected {
			t.Errorf("expected %v for path %s with suffix %s, got %v", test.expected, test.path, test.suffix, result)
		}
	}
}

func TestHasExt(t *testing.T) {
	tests := []struct {
		path     Path
		ext      string
		expected bool
	}{
		{New("file.txt"), ".txt", true},
		{New("file.txt"), "txt", true},
		{New("file.pdf"), ".txt", false},
		{New("file"), "", true},
		{New("file.tar.gz"), ".gz", true},
	}

	for _, test := range tests {
		result := test.path.HasExt(test.ext)
		if result != test.expected {
			t.Errorf("expected %v for path %s with ext %s, got %v", test.expected, test.path, test.ext, result)
		}
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		path     Path
		sub      string
		expected bool
	}{
		{New("/home/user/documents"), "user", true},
		{New("/home/user/documents"), "admin", false},
		{New("test_file.txt"), "test", true},
		{New("test_file.txt"), "demo", false},
	}

	for _, test := range tests {
		result := test.path.Contains(test.sub)
		if result != test.expected {
			t.Errorf("expected %v for path %s with substring %s, got %v", test.expected, test.path, test.sub, result)
		}
	}
}

func TestTrim(t *testing.T) {
	tests := []struct {
		input    Path
		expected string
	}{
		{New("  /path/to/file  "), "/path/to/file"},
		{New("\t/path/to/file\n"), "/path/to/file"},
		{New("/path/to/file"), "/path/to/file"},
		{New("  "), ""},
	}

	for _, test := range tests {
		result := test.input.Trim()
		if result.String() != test.expected {
			t.Errorf("expected %s, got %s", test.expected, result.String())
		}
	}
}

func TestClean(t *testing.T) {
	tests := []struct {
		input    Path
		expected string
	}{
		{New("a/b/../c"), filepath.Clean("a/b/../c")},
		{New("a//b//c"), filepath.Clean("a//b//c")},
		{New("./a/b"), filepath.Clean("./a/b")},
		{New("a/./b"), filepath.Clean("a/./b")},
	}

	for _, test := range tests {
		result := test.input.Clean()
		if result.String() != test.expected {
			t.Errorf("expected %s, got %s", test.expected, result.String())
		}
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		input    Path
		expected string
	}{
		{New("a/b/../c"), filepath.Clean("a/b/../c")},
		{New("a//b//c"), filepath.Clean("a//b//c")},
	}

	for _, test := range tests {
		result := test.input.Normalize()
		if result.String() != test.expected {
			t.Errorf("expected %s, got %s", test.expected, result.String())
		}
	}
}

func TestXXH64(t *testing.T) {
	p := New("testfile.txt")
	defer p.Delete()

	if err := p.WriteFile(testContent); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	hash := p.XXH64()
	if hash == nil {
		t.Errorf("expected non-nil hash")
	}
	if len(hash) == 0 {
		t.Errorf("expected non-empty hash")
	}
}

func TestXXH64String(t *testing.T) {
	p := New("testfile.txt")
	defer p.Delete()

	if err := p.WriteFile(testContent); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	hashStr := p.XXH64String()
	if hashStr == "" {
		t.Errorf("expected non-empty hash string")
	}
	if len(hashStr) != 16 {
		t.Errorf("expected 16 character hash string, got %d", len(hashStr))
	}
}

func TestTemp(t *testing.T) {
	p := Temp()
	defer p.Delete()

	if p.String() == "" {
		t.Errorf("expected non-empty path")
	}
}

func TestTempFile(t *testing.T) {
	f, err := TempFile("testpattern-*.txt")
	if err != nil {
		t.Fatalf("TempFile: %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	if !New(f.Name()).IsExist() {
		t.Errorf("expected temporary file to exist")
	}
}

func TestTempDir(t *testing.T) {
	p := TempDir("subdir", "nested")

	if !p.Contains("subdir") {
		t.Errorf("expected path to contain 'subdir'")
	}
	if !p.Contains("nested") {
		t.Errorf("expected path to contain 'nested'")
	}
}

func TestSegments(t *testing.T) {
	if runtime.GOOS == "windows" {
		p := New("C:\\path\\to\\file")
		segments := p.Segments()
		if len(segments) != 1 {
			t.Errorf("expected 1 segment on Windows, got %d", len(segments))
		}
	} else {
		p := New("/usr/local/bin")
		segments := p.Segments()
		if len(segments) != 1 {
			t.Errorf("expected 1 segment on Unix, got %d", len(segments))
		}
	}
}

func TestTotalSize(t *testing.T) {
	// Test regular file
	t.Run("RegularFile", func(t *testing.T) {
		p := New("testfile_totalsize.txt")
		content := []byte("test content for size")
		if err := p.WriteFile(content); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		defer p.Delete()

		size, err := p.TotalSize()
		if err != nil {
			t.Fatalf("TotalSize() error: %v", err)
		}

		expected := int64(len(content))
		if size != expected {
			t.Errorf("expected %d, got %d", expected, size)
		}
	})

	// Test directory with files
	t.Run("DirectoryWithFiles", func(t *testing.T) {
		dir := New("testdir_totalsize")
		if err := dir.MkdirIfNotExist(); err != nil {
			t.Fatalf("MkdirIfNotExist: %v", err)
		}
		defer dir.Delete()

		// Create files in directory
		file1 := dir.Join("file1.txt")
		content1 := []byte("content1")
		if err := file1.WriteFile(content1); err != nil {
			t.Fatalf("WriteFile file1: %v", err)
		}

		file2 := dir.Join("file2.txt")
		content2 := []byte("content2 is longer")
		if err := file2.WriteFile(content2); err != nil {
			t.Fatalf("WriteFile file2: %v", err)
		}

		totalSize, err := dir.TotalSize()
		if err != nil {
			t.Fatalf("TotalSize() error: %v", err)
		}

		// Total should be dir size + file1 size + file2 size
		dirInfo, _ := dir.Stat()
		expectedMin := dirInfo.Size() + int64(len(content1)) + int64(len(content2))
		if totalSize < expectedMin {
			t.Errorf("expected at least %d, got %d", expectedMin, totalSize)
		}
	})

	// Test nested directory structure
	t.Run("NestedDirectories", func(t *testing.T) {
		dir := New("testdir_nested_totalsize")
		if err := dir.MkdirIfNotExist(); err != nil {
			t.Fatalf("MkdirIfNotExist: %v", err)
		}
		defer dir.Delete()

		// Create nested structure
		subdir := dir.Join("subdir")
		if err := subdir.MkdirIfNotExist(); err != nil {
			t.Fatalf("MkdirIfNotExist subdir: %v", err)
		}

		file1 := dir.Join("top.txt")
		content1 := []byte("top level")
		if err := file1.WriteFile(content1); err != nil {
			t.Fatalf("WriteFile top: %v", err)
		}

		file2 := subdir.Join("nested.txt")
		content2 := []byte("nested content")
		if err := file2.WriteFile(content2); err != nil {
			t.Fatalf("WriteFile nested: %v", err)
		}

		totalSize, err := dir.TotalSize()
		if err != nil {
			t.Fatalf("TotalSize() error: %v", err)
		}

		// Total should include both directories and both files
		dirInfo, _ := dir.Stat()
		subdirInfo, _ := subdir.Stat()
		expectedMin := dirInfo.Size() + subdirInfo.Size() + int64(len(content1)) + int64(len(content2))
		if totalSize < expectedMin {
			t.Errorf("expected at least %d, got %d", expectedMin, totalSize)
		}
	})
}

func TestTotalSizeX(t *testing.T) {
	dir := New("testdir_totalsizex")
	if err := dir.MkdirIfNotExist(); err != nil {
		t.Fatalf("MkdirIfNotExist: %v", err)
	}
	defer dir.Delete()

	file := dir.Join("test.txt")
	content := []byte("test content")
	if err := file.WriteFile(content); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	totalSize := dir.TotalSizeX()
	if totalSize == 0 {
		t.Error("expected non-zero total size")
	}

	// Should be at least the size of the file
	if totalSize < int64(len(content)) {
		t.Errorf("expected at least %d, got %d", int64(len(content)), totalSize)
	}
}
