// 文件相关函数
//   变更历史
//     2019-08-23  lixiaoya  新建
package utils

import (
	"errors"
	"os"
	"syscall"
)

// FileCtime 获取文件的创建时间
// linux版本使用Ctim，mac版本使用Ctimespec，为发兼容只能使用反射去找对应的字段名
//   参数
//     path: 文件路径
//   返回
//     秒、纳秒、错误信息
func FileCtime(path string) (int64, int64, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return -1, -1, err
	}

	sysInfo := fi.Sys()
	if stat, ok := sysInfo.(*syscall.Stat_t); ok {
		// linux使用Ctim, mac使用Ctimespec
		return stat.Ctim.Sec, stat.Ctim.Nsec, nil
	}

	return -1, -1, errors.New("Assertion error")
}

// FileMtime 获取文件的修改时间
// linux版本使用Mtim，mac版本使用Mtimespec
//   参数
//     path: 文件路径
//   返回
//     秒、纳秒、错误信息
func FileMtime(path string) (int64, int64, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return -1, -1, err
	}

	sysInfo := fi.Sys()
	if stat, ok := sysInfo.(*syscall.Stat_t); ok {
		// linux使用Mtim, mac使用Mtimespec
		return stat.Mtim.Sec, stat.Mtim.Nsec, nil
	}

	return -1, -1, errors.New("Assertion error")
}

// FileAtime 获取文件的访问时间
// linux版本使用Atim，mac版本使用Atimespec
//   参数
//     path: 文件路径
//   返回
//     秒、纳秒、错误信息
func FileAtime(path string) (int64, int64, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return -1, -1, err
	}

	sysInfo := fi.Sys()
	if stat, ok := sysInfo.(*syscall.Stat_t); ok {
		// linux使用Atim, mac使用Atimespec
		return stat.Atim.Sec, stat.Atim.Nsec, nil
	}

	return -1, -1, errors.New("Assertion error")
}
