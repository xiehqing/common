package util

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/xiehqing/common/pkg/logs"
	"gopkg.in/yaml.v3"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func ToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatInt(int64(v), 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	}
	return ""
}

func ToMap(obj interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &result)
	return result, err
}

func Int64Join(elements []int64, sep string) string {
	var dats []string
	for _, ele := range elements {
		dats = append(dats, strconv.FormatInt(ele, 10))
	}
	return strings.Join(dats, sep)
}

func DivideInt64(a, b int64, precision int) float64 {
	if b == 0 {
		return 0 // 或者返回错误
	}
	result := float64(a) / float64(b)
	factor := math.Pow(10, float64(precision))
	return math.Round(result*factor) / factor
}

func BytesToGB(bytes uint64) float64 {
	return float64(bytes) / (1024 * 1024 * 1024)
}

// Convert 对象转换
func Convert[T interface{}](src interface{}) (*T, error) {
	b, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	var dat T
	err = json.Unmarshal(b, &dat)
	if err != nil {
		return nil, err
	}
	return &dat, nil
}

// ToFloat64 将任意数字类型转换为float64
func ToFloat64(input interface{}) (float64, error) {
	// Check for the kind of the value first
	if input == nil {
		return 0, fmt.Errorf("unsupported type: %T", input)
	}

	kind := reflect.TypeOf(input).Kind()
	switch kind {
	case reflect.Float64:
		return input.(float64), nil
	case reflect.Float32:
		return float64(input.(float32)), nil
	case reflect.Int, reflect.Int32, reflect.Int64, reflect.Int8, reflect.Int16:
		return float64(reflect.ValueOf(input).Int()), nil
	case reflect.Uint, reflect.Uint32, reflect.Uint64, reflect.Uint8, reflect.Uint16:
		return float64(reflect.ValueOf(input).Uint()), nil
	case reflect.String:
		return strconv.ParseFloat(input.(string), 64)
	case reflect.Bool:
		if input.(bool) {
			return 1.0, nil
		}
		return 0.0, nil
	default:
		return 0, fmt.Errorf("unsupported number type: %T", input)
	}
}

// ToInt64 将任意数字类型转换为int64
func ToInt64(input interface{}) (int64, error) {
	if input == nil {
		return 0, fmt.Errorf("unsupported type: %T", input)
	}
	kind := reflect.TypeOf(input).Kind()
	switch kind {
	case reflect.Float64:
		return int64(input.(float64)), nil
	case reflect.Float32:
		return int64(input.(float32)), nil
	case reflect.Int, reflect.Int32, reflect.Int64, reflect.Int8, reflect.Int16:
		return int64(reflect.ValueOf(input).Int()), nil
	case reflect.Uint, reflect.Uint32, reflect.Uint64, reflect.Uint8, reflect.Uint16:
		return int64(reflect.ValueOf(input).Uint()), nil
	case reflect.String:
		return strconv.ParseInt(input.(string), 10, 64)
	case reflect.Bool:
		if input.(bool) {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("unsupported number type: %T", input)
	}
	return 0, nil
}

// ToJson 对象转换为json
func ToJson(o interface{}) (string, error) {
	b, err := json.Marshal(o)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ToJsonIgnoreError 对象转换为json，忽略错误
func ToJsonIgnoreError(o interface{}) string {
	if o == nil {
		logs.Errorf("[ToJsonIgnoreError]对象为nil")
		return ""
	}
	b, err := json.Marshal(o)
	if err != nil {
		logs.Errorf("[ToJsonIgnoreError]对象转换为json失败：%s", err.Error())
		return ""
	}
	return string(b)
}

// Yml2Json yml转json
func Yml2Json(content string) (string, error) {
	ymlContent := strings.ReplaceAll(content, "\t", "")
	var yamlObj interface{}
	if err := yaml.Unmarshal([]byte(ymlContent), &yamlObj); err != nil {
		logs.Errorf("解析yaml错误：%v，尝试预处理后再解析", err)
		ymlContent = preprocessYAML(ymlContent)
		if err := yaml.Unmarshal([]byte(ymlContent), &yamlObj); err != nil {
			logs.Errorf("解析yaml错误：%v", err)
			return "", err
		}
	}
	jsonBytes, err := json.MarshalIndent(yamlObj, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// preprocessYAML 预处理YAML内容
func preprocessYAML(yamlContent string) string {
	lines := strings.Split(yamlContent, "\n")

	for i, line := range lines {
		// 检查schedule行并确保格式正确
		if strings.Contains(line, "schedule:") && !strings.Contains(line, `"`) && !strings.Contains(line, `'`) {
			// 检查是否包含特殊字符
			if strings.ContainsAny(line, "/*") {
				// 在值周围添加引号
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					value := strings.TrimSpace(parts[1])
					if !strings.HasPrefix(value, `"`) && !strings.HasPrefix(value, `'`) {
						lines[i] = parts[0] + ": '" + value + "'"
					}
				}
			}
		}
	}

	return strings.Join(lines, "\n")
}

// Yml2Map yml转map
func Yml2Map(content string) (map[string]interface{}, error) {
	ymlJson, err := Yml2Json(content)
	if err != nil {
		return nil, err
	}
	var yamlObj map[string]interface{}
	if err := json.Unmarshal([]byte(ymlJson), &yamlObj); err != nil {
		return nil, err
	}
	return yamlObj, nil
}

// IsBase64URLSafe 判断是否为URL安全的Base64编码
func IsBase64URLSafe(str string) (bool, []byte) {
	cleanStr := strings.TrimSpace(str)

	// URL安全的Base64允许长度不是4的倍数
	content, err := base64.RawURLEncoding.DecodeString(cleanStr)
	return err == nil, content
}

// CanConvertFloatToInt64 判断float64是否可以安全转换为int64
func CanConvertFloatToInt64(f float64) bool {
	// 检查是否为NaN或Infinity
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return false
	}

	// 检查是否超出int64范围
	if f < math.MinInt64 || f > math.MaxInt64 {
		return false
	}

	// 检查是否有小数部分（如果要求必须是整数）
	// 注意：这里只是检查是否可以转换为int64，不要求一定是整数
	// 如果需要必须是整数，可以添加以下检查：
	if math.Trunc(f) != f {
		return false
	}

	return true
}

// isPureDigits 检查是否为纯数字
func isPureDigits(str string) bool {
	// 匹配纯数字（可能包含前导0）
	matched, _ := regexp.MatchString(`^\d+$`, str)
	return matched
}

// isValidBase64URLChars 检查是否只包含Base64 URL安全字符
func isValidBase64URLChars(str string) bool {
	// Base64 URL安全字符集: A-Z, a-z, 0-9, -, _
	// 注意：= 是填充字符，在RawURLEncoding中通常不需要
	base64URLRegex := regexp.MustCompile(`^[A-Za-z0-9\-_]+$`)
	return base64URLRegex.MatchString(str)
}

// AnalysisUrl 解析url
func AnalysisUrl(url string, defaultPort string) (string, string) {
	// 处理前将url中的单引号或者双引号去掉
	if strings.HasSuffix(url, "/") {
		url = url[:len(url)-1]
	}
	url = strings.ReplaceAll(url, "'", "")
	url = strings.ReplaceAll(url, "\"", "")
	url = strings.ReplaceAll(url, "[", "")
	url = strings.ReplaceAll(url, "]", "")
	parts := strings.Split(url, "://")
	var host, port string
	if len(parts) == 2 {
		hostPort := strings.Split(parts[1], ":")
		if len(hostPort) == 2 {
			host = hostPort[0]
			port = hostPort[1]
		} else {
			host = hostPort[0]
			port = defaultPort
		}
	} else {
		hostPort := strings.Split(url, ":")
		if len(hostPort) == 2 {
			host = hostPort[0]
			port = hostPort[1]
		} else {
			host = hostPort[0]
			port = defaultPort
		}
	}
	return host, port
}

// IsCronExpression 判断字符串是否为合法的 Cron 表达式
func IsCronExpression(s string) bool {
	// Cron 表达式的基本正则规则（简化版，可根据需求扩展）
	// 格式：秒 分 时 日 月 周（年可选）
	cronRegex := `^(\*|(\d+(-\d+)?)(/\d+)?(,(\d+(-\d+)?)(/\d+)?)*$`
	match, _ := regexp.MatchString(cronRegex, s)
	return match
}

// ContainsInt64 判断切片中是否包含指定元素
func ContainsInt64(elements []int64, element int64) bool {
	for _, e := range elements {
		if e == element {
			return true
		}
	}
	return false
}

// ContainsInt 判断切片中是否包含指定元素
func ContainsInt(elements []int, element int) bool {
	for _, e := range elements {
		if e == element {
			return true
		}
	}
	return false
}

// TitleCase 首字母大写
func TitleCase(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

// RemoveStrBothSideDoubleQuotes 去除字符串两边的双引号
func RemoveStrBothSideDoubleQuotes(str string) string {
	if strings.HasPrefix(str, "\"") && strings.HasSuffix(str, "\"") {
		return str[1 : len(str)-1]
	}
	return str
}

type RoundedFloat float64

// RoundTo 保留指定小数位数
func (f RoundedFloat) RoundTo(precision int) float64 {
	multiplier := math.Pow10(precision)
	return math.Round(float64(f)*multiplier) / multiplier
}
