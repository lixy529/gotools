// 文件相关函数测试
//   变更历史
//     2017-02-20  lixiaoya  新建
package utils

import (
	"testing"
)

// TestFileCtime FileCtime函数测试
func TestFileCtime(t *testing.T) {
	sec, nsec, err := FileCtime("/tmp/test.txt")
	if err != nil {
		t.Error("FileCtime err", err.Error())
		return
	}
	t.Log(sec, nsec)
}

// TestFileMtime FileMtime函数测试
func TestFileMtime(t *testing.T) {
	sec, nsec, err := FileMtime("/tmp/test.txt")
	if err != nil {
		t.Error("FileMtime err", err.Error())
		return
	}
	t.Log(sec, nsec)
}

// TestFileAtime FileAtime函数测试
func TestFileAtime(t *testing.T) {
	sec, nsec, err := FileAtime("/tmp/test.txt")
	if err != nil {
		t.Error("FileAtime err", err.Error())
		return
	}
	t.Log(sec, nsec)
}
