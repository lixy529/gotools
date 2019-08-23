package utils

import (
	"fmt"
	"os"
	"testing"
)

// TestIsFile test IsFile function.
func TestIsFile(t *testing.T) {
	name := "./utils.go"

	if r, _ := IsFile(name); !r {
		t.Error("StarWithStr failed. Got false, expected true.")
		return
	}
}

// TestMkDir test MkDir function.
func TestMkDir(t *testing.T) {
	name := "./test/dd/cc/aa.pid"

	if err := MkDir(name, 0777, true); err != nil {
		t.Error("MkDir failed")
		return
	}
}

// TestWriteFile test WriteFile function.
func TestWriteFile(t *testing.T) {
	name := "/tmp/test.log"
	for i := 0; i < 10; i++ {
		data := fmt.Sprintf("message_%02d\n", i)
		_, err := WriteFile(name, []byte(data), os.O_RDWR|os.O_CREATE|os.O_APPEND, os.FileMode(0660))
		if err != nil {
			t.Errorf("WriteFile err: %s", err.Error())
			return
		}
	}
}
