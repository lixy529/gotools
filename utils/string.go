package utils

import (
	"strings"
	"regexp"
)

// DelRepeat remove duplicate characters.
// eg: a//b//c////c => a/b/c
func DelRepeat(s string, c byte) string {
	b := []byte{}
	j := 0
	flag := false
	for i := 0; i < len(s); i++ {
		if flag && s[i] == c {
			continue
		}

		if s[i] == c {
			flag = true
		} else {
			flag = false
		}

		b = append(b, s[i])
		j++
	}

	return string(b)
}

// Replace replace string.
// If the number of olds and news is different, the data is less.
// n: Replace the number, 0 is not replaced, less than 0 is all replaced.
// eg: news-["m1", "m2"] olds-["+", "=] 1m12m23 => 1+2=3
func Replace(src string, olds []string, news []string, n int) string {
	oldCnt := len(olds)
	newCnt := len(news)
	if src == "" || oldCnt == 0 || newCnt == 0 || n == 0 {
		return src
	}

	cnt := oldCnt
	if cnt > newCnt {
		cnt = newCnt
	}

	dst := src
	for i := 0; i < cnt; i++ {
		dst = strings.Replace(dst, olds[i], news[i], n)
	}

	return dst
}

// Substr returns sub string.
// start: Intercept pos
// length: Intercept length
func Substr(str string, start, length int) string {
	l := len(str)
	if start >= l {
		return ""
	}

	if start < 0 {
		start = l + start
		if start < 0 {
			start = 0
		}
	}

	end := start + length
	if end > l {
		end = l
	}

	return str[start:end]
}

// Empty check the string is empty.
func Empty(str string) bool {
	if len(str) == 0 {
		return true
	}

	return false
}

// MbLen get the number of characters in the string.
func MbLen(str string) int {
	b := []rune(str)

	return len(b)
}

// GetSafeSql delete sql special characters
func GetSafeSql(str string) string {
	if Empty(str) {
		return ""
	}
	pattern := `\b(?i:sleep|delay|waitfor|and|exec|execute|insert|select|delete|update|count|master|char|declare|net user|xp_cmdshell|or|create|drop|table|from|grant|use|group_concat|column_name|information_schema.columns|table_schema|union|where|orderhaving|having|by|truncate|like)\b`
	reg := regexp.MustCompile(pattern)

	return reg.ReplaceAllString(str, "")
}

// Html2Str delete html special characters.
func Html2Str(html string) string {
	src := string(html)

	re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	src = re.ReplaceAllStringFunc(src, strings.ToLower)

	//remove STYLE
	re, _ = regexp.Compile("\\<style[\\S\\s]+?\\</style\\>")
	src = re.ReplaceAllString(src, "")

	//remove SCRIPT
	re, _ = regexp.Compile("\\<script[\\S\\s]+?\\</script\\>")
	src = re.ReplaceAllString(src, "")

	re, _ = regexp.Compile("\\<[\\S\\s]+?\\>")
	src = re.ReplaceAllString(src, "\n")

	return strings.TrimSpace(src)
}

// StrSplit split str by the specified length.
// Return an empty slice if the src length is less than or equal to 0
func StrSplit(src string, length ...int) []string {
	length = append(length, 1)
	res := []string{}
	if length[0] <= 0 {
		return res
	}
	srcLen := len(src)
	if srcLen <= length[0] {
		return append(res, src)
	}

	pos := 0
	tmp := make([]byte, length[0])
	for i := 0; i < srcLen; i++ {
		tmp[pos] = src[i]
		pos++
		if pos == length[0] {
			res = append(res, string(tmp))
			tmp = make([]byte, length[0])
			pos = 0
		}
	}

	if pos > 0 {
		res = append(res, string(tmp))
	}

	return res
}
