package path

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

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

func TestRename(t *testing.T) {
	p := New("testfile.txt")
	os.WriteFile(p.String(), []byte("test"), 0o644)
	newName := "newtestfile.txt"
	err := p.Rename(newName)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !New(newName).IsExist() {
		t.Errorf("expected file to be renamed")
	}
	os.Remove(newName)
}

func TestCopy(t *testing.T) {
	src := New("srcfile.txt")
	dst := New("dstfile.txt")
	os.WriteFile(src.String(), []byte("test"), 0o644)
	err := src.Copy(dst)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !dst.IsExist() {
		t.Errorf("expected file to be copied")
	}
	os.Remove(src.String())
	os.Remove(dst.String())
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
