package ppath

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
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
	if err.Error() != "not a directory" {
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
		if err.Error() != "source file does not exist" {
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
		if err.Error() != "already exists" {
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
