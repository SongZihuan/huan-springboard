package utils

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

const BASE_CHAR = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStr(length int) string {
	bytes := []byte(BASE_CHAR)

	var result []byte
	for i := 0; i < length; i++ {
		result = append(result, bytes[Rand().Intn(len(bytes))])
	}

	return string(result)
}

func InvalidPhone(phone string) bool {
	pattern := `^1[3-9]\d{9}$`
	matched, _ := regexp.MatchString(pattern, phone)
	return matched
}

func IsValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

const NormalConsoleWidth = 80

func FormatTextToWidth(text string, width int) string {
	return FormatTextToWidthAndPrefix(text, 0, width)
}

func FormatTextToWidthAndPrefix(text string, prefixWidth int, overallWidth int) string {
	var result strings.Builder

	width := overallWidth - prefixWidth
	if width <= 0 {
		panic("bad width")
	}

	text = strings.ReplaceAll(text, "\r\n", "\n")

	for _, line := range strings.Split(text, "\n") {
		result.WriteString(strings.Repeat(" ", prefixWidth))

		if line == "" {
			result.WriteString("\n")
			continue
		}

		spaceCount := CountSpaceInStringPrefix(line) % width
		newLineLength := 0
		if spaceCount < 80 {
			result.WriteString(strings.Repeat(" ", spaceCount))
			newLineLength = spaceCount
		}

		for _, word := range strings.Fields(line) {
			if newLineLength+len(word) >= width {
				result.WriteString("\n")
				result.WriteString(strings.Repeat(" ", prefixWidth))
				newLineLength = 0
			}

			// 不是第一个词时，添加空格
			if newLineLength != 0 {
				result.WriteString(" ")
				newLineLength += 1
			}

			result.WriteString(word)
			newLineLength += len(word)
		}

		if newLineLength != 0 {
			result.WriteString("\n")
			newLineLength = 0
		}
	}

	return strings.TrimRight(result.String(), "\n")
}

func CountSpaceInStringPrefix(str string) int {
	var res int
	for _, r := range str {
		if r == ' ' {
			res += 1
		} else {
			break
		}
	}

	return res
}

func IsValidURLPath(path string) bool {
	if path == "" {
		return true
	} else if path == "/" {
		return false
	}

	pattern := `^\/[a-zA-Z0-9\-._~:/?#\[\]@!$&'()*+,;%=]+$`
	matched, _ := regexp.MatchString(pattern, path)
	return matched
}

func IsValidDomain(domain string) bool {
	pattern := `^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$`
	matched, _ := regexp.MatchString(pattern, domain)
	return matched
}

func StringToOnlyPrint(str string) string {
	runeLst := []rune(str)
	res := make([]rune, 0, len(runeLst))

	for _, r := range runeLst {
		if unicode.IsPrint(r) {
			res = append(res, r)
		}
	}

	return string(res)
}

func IsGoodQueryKey(key string) bool {
	pattern := `^[a-zA-Z0-9\-._~]+$`
	matched, _ := regexp.MatchString(pattern, key)
	return matched
}

func IsValidHTTPHeaderKey(key string) bool {
	pattern := `^[a-zA-Z0-9!#$%&'*+.^_` + "`" + `|~-]+$`
	matched, _ := regexp.MatchString(pattern, key)
	return matched
}

func ReadBytes(str string) uint64 {
	if strings.HasPrefix(strings.ToUpper(str), "TB") {
		numStr := str[:len(str)-2]
		num, _ := strconv.ParseUint(numStr, 10, 64)
		return num * 1024 * 1024 * 1024 * 1024
	} else if strings.HasPrefix(strings.ToLower(str), "tbytes") {
		numStr := str[:len(str)-6]
		num, _ := strconv.ParseUint(numStr, 10, 64)
		return num * 1024 * 1024 * 1024 * 1024
	} else if strings.HasPrefix(strings.ToLower(str), "tbyte") {
		numStr := str[:len(str)-5]
		num, _ := strconv.ParseUint(numStr, 10, 64)
		return num * 1024 * 1024 * 1024 * 1024
	} else if strings.HasPrefix(strings.ToLower(str), "terabytes") {
		numStr := str[:len(str)-9]
		num, _ := strconv.ParseUint(numStr, 10, 64)
		return num * 1024 * 1024 * 1024 * 1024
	} else if strings.HasPrefix(strings.ToLower(str), "terabyte") {
		numStr := str[:len(str)-8]
		num, _ := strconv.ParseUint(numStr, 10, 64)
		return num * 1024 * 1024 * 1024 * 1024
	}

	if strings.HasPrefix(strings.ToUpper(str), "GB") {
		numStr := str[:len(str)-2]
		num, _ := strconv.ParseUint(numStr, 10, 64)
		return num * 1024 * 1024 * 1024
	} else if strings.HasPrefix(strings.ToLower(str), "gbytes") {
		numStr := str[:len(str)-6]
		num, _ := strconv.ParseUint(numStr, 10, 64)
		return num * 1024 * 1024 * 1024
	} else if strings.HasPrefix(strings.ToLower(str), "gbyte") {
		numStr := str[:len(str)-5]
		num, _ := strconv.ParseUint(numStr, 10, 64)
		return num * 1024 * 1024 * 1024
	} else if strings.HasPrefix(strings.ToLower(str), "gigabytes") {
		numStr := str[:len(str)-9]
		num, _ := strconv.ParseUint(numStr, 10, 64)
		return num * 1024 * 1024 * 1024
	} else if strings.HasPrefix(strings.ToLower(str), "gigabyte") {
		numStr := str[:len(str)-8]
		num, _ := strconv.ParseUint(numStr, 10, 64)
		return num * 1024 * 1024 * 1024
	}

	if strings.HasPrefix(strings.ToUpper(str), "MB") {
		numStr := str[:len(str)-2]
		num, _ := strconv.ParseUint(numStr, 10, 64)
		return num * 1024 * 1024
	} else if strings.HasPrefix(strings.ToLower(str), "mbytes") {
		numStr := str[:len(str)-6]
		num, _ := strconv.ParseUint(numStr, 10, 64)
		return num * 1024 * 1024
	} else if strings.HasPrefix(strings.ToLower(str), "mbyte") {
		numStr := str[:len(str)-5]
		num, _ := strconv.ParseUint(numStr, 10, 64)
		return num * 1024 * 1024
	} else if strings.HasPrefix(strings.ToLower(str), "megabytes") {
		numStr := str[:len(str)-9]
		num, _ := strconv.ParseUint(numStr, 10, 64)
		return num * 1024 * 1024
	} else if strings.HasPrefix(strings.ToLower(str), "megabyte") {
		numStr := str[:len(str)-8]
		num, _ := strconv.ParseUint(numStr, 10, 64)
		return num * 1024 * 1024
	}

	if strings.HasPrefix(strings.ToUpper(str), "KB") {
		numStr := str[:len(str)-2]
		num, _ := strconv.ParseUint(numStr, 10, 64)
		return num * 1024
	} else if strings.HasPrefix(strings.ToLower(str), "kbytes") {
		numStr := str[:len(str)-6]
		num, _ := strconv.ParseUint(numStr, 10, 64)
		return num * 1024
	} else if strings.HasPrefix(strings.ToLower(str), "kbyte") {
		numStr := str[:len(str)-5]
		num, _ := strconv.ParseUint(numStr, 10, 64)
		return num * 1024
	} else if strings.HasPrefix(strings.ToLower(str), "kilobytes") {
		numStr := str[:len(str)-9]
		num, _ := strconv.ParseUint(numStr, 10, 64)
		return num * 1024
	} else if strings.HasPrefix(strings.ToUpper(str), "kilobyte") {
		numStr := str[:len(str)-8]
		num, _ := strconv.ParseUint(numStr, 9, 64)
		return num * 1024
	}

	if strings.HasPrefix(strings.ToUpper(str), "B") {
		numStr := str[:len(str)-1]
		num, _ := strconv.ParseUint(numStr, 10, 64)
		return num
	} else if strings.HasPrefix(strings.ToLower(str), "bytes") {
		numStr := str[:len(str)-5]
		num, _ := strconv.ParseUint(numStr, 10, 64)
		return num
	} else if strings.HasPrefix(strings.ToLower(str), "byte") {
		numStr := str[:len(str)-4]
		num, _ := strconv.ParseUint(numStr, 10, 64)
		return num
	}

	num, _ := strconv.ParseUint(str, 10, 64)
	return num
}
