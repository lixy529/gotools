// 文件相关函数测试
//   变更历史
//     2017-02-20  lixiaoya  新建
package utils

import (
	"fmt"
	"testing"
)

// TestFileCtime FileCtime函数测试
func TestFileCtime(t *testing.T) {
	sec, nsec, err := FileCtime("/tmp/test.txt")
	fmt.Println(sec, nsec, err)
}

// TestFileMtime FileMtime函数测试
func TestFileMtime(t *testing.T) {
	sec, nsec, err := FileMtime("/tmp/test.txt")
	fmt.Println(sec, nsec, err)
}

// TestFileAtime FileAtime函数测试
func TestFileAtime(t *testing.T) {
	sec, nsec, err := FileAtime("/tmp/test.txt")
	fmt.Println(sec, nsec, err)
}
