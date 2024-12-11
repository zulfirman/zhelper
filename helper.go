package zhelper

import (
	"errors"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo/v4"
	"github.com/rs/xid"
	"github.com/thedevsaddam/gojsonq/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

var (
	poolReq = &sync.Pool{
		New: func() interface{} {
			return resty.New()
		},
	}
)

type H map[string]interface{}

// Rs handles JSON responses with a status code and path
func Rs(c echo.Context, result Response) error {
	if result.Code == 0 {
		result.Code = 200
	}
	result.Status = http.StatusText(result.Code)
	result.Path = Substr(c.Request().RequestURI, 150)
	return c.JSON(result.Code, result)
}

// RsMessage returns a response with a custom message and code
func RsMessage(c echo.Context, code int, message interface{}) error {
	result := Response{
		Code: code,
		Content: H{
			"message": message,
		},
	}
	return Rs(c, result)
}

// RsSuccess returns a generic success message
func RsSuccess(c echo.Context) error {
	return RsMessage(c, 200, "success")
}

// RsError returns a response with error details
func RsError(c echo.Context, code int, message interface{}) error {
	return RsMessage(c, code, message)
}

// KeyExists checks if a key exists in a decoded JSON map
func KeyExists(decoded map[string]interface{}, key string) bool {
	val, ok := decoded[key]
	return ok && val != nil
}

// ReadyBodyJson binds JSON data from request body to map
func ReadyBodyJson(c echo.Context, json map[string]interface{}) map[string]interface{} {
	if err := c.Bind(&json); err != nil {
		return nil
	}
	return json
}

// GetReq sends a GET request and checks for status code
func GetReq(url, token string) (*resty.Response, error) {
	client := poolReq.Get().(*resty.Client)
	defer poolReq.Put(client)

	resp, err := client.R().EnableTrace().SetAuthToken(token).Get(url)
	if err != nil {
		return resp, err
	}
	if resp.StatusCode() != 200 {
		fmt.Printf("Request success but status code is %d: %s\n", resp.StatusCode(), resp.String())
		return resp, errors.New("errCode")
	}
	return resp, err
}

// PostReq sends a POST request with an auth token and JSON body
func PostReq(url, token string, body interface{}) (*resty.Response, error) {
	client := poolReq.Get().(*resty.Client)
	defer poolReq.Put(client)

	resp, err := client.R().
		EnableTrace().
		SetAuthToken(token).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(url)

	if err != nil {
		return resp, err
	}
	if resp.StatusCode() != 200 {
		fmt.Printf("Request success but status code is %d: %s\n", resp.StatusCode(), resp.String())
		return resp, errors.New("errCode")
	}
	return resp, err
}

// JsonToMap converts a JSON string to a map
func JsonToMap(jsonStr string) map[string]interface{} {
	var result map[string]interface{}
	if err := sonic.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil
	}
	return result
}

// MarshalBinary serializes an object to a binary format
func MarshalBinary(i interface{}) []byte {
	data, err := sonic.Marshal(i)
	if err != nil {
		log.Printf("Error marshalling binary: %s", err.Error())
		return nil
	}
	return data
}

// ReadJson retrieves values from a JSON string with optional path
func ReadJson(jsonString, path string) interface{} {
	q := gojsonq.New().FromString(jsonString)
	if path != "" {
		return q.Find(path)
	}
	return q.Get()
}

// RemoveField removes specified fields from a JSON object
func RemoveField(obj interface{}, ignoreFields ...string) (interface{}, error) {
	jsonData, err := sonic.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var mapData map[string]interface{}
	if err := sonic.Unmarshal(jsonData, &mapData); err != nil {
		return nil, err
	}

	for _, field := range ignoreFields {
		delete(mapData, field)
	}

	return mapData, nil
}

// Substr limits a string to a maximum length
func Substr(input string, limit int) string {
	if len([]rune(input)) > limit {
		return input[:limit]
	}
	return input
}

// UniqueId generates a unique identifier using xid
func UniqueId() string {
	return xid.New().String()
}

// DateNow formats the current time in various formats
func DateNow(typeFormat int) string {
	switch typeFormat {
	case 1:
		return time.Now().UTC().Format("2006-01-02")
	case 2:
		return time.Now().Format("15:04:05")
	case 3:
		return time.Now().UTC().String()
	default:
		return ""
	}
}

// GormTime converts a time object to GORM datatypes.Time
func GormTime(timeParam time.Time) datatypes.Time {
	return datatypes.NewTime(timeParam.Hour(), timeParam.Minute(), timeParam.Second(), 0)
}

// DeletedAt returns the current time as a valid GORM deleted timestamp
func DeletedAt() gorm.DeletedAt {
	return gorm.DeletedAt{
		Time:  time.Now(),
		Valid: true,
	}
}

// IntString converts an int to string
func IntString(param int) string {
	return strconv.Itoa(param)
}

// StringInt converts a string to int
func StringInt(param string) int {
	intVar, _ := strconv.Atoi(param)
	return intVar
}

// Int64String converts an int64 to string
func Int64String(param int64) string {
	return strconv.FormatInt(param, 10)
}

// StringInt64 converts a string to int64
func StringInt64(param string) int64 {
	intVal, _ := strconv.ParseInt(param, 10, 64)
	return intVal
}

// BlankString checks if a string is empty or consists only of whitespace
func BlankString(s string) bool {
	return strings.TrimSpace(s) == "" || strings.HasPrefix(s, " ")
}

// FailOnError logs a fatal error message
func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

// GetParamPagination extracts pagination parameters from query
func GetParamPagination(c echo.Context) Pagination {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	page, _ := strconv.Atoi(c.QueryParam("page"))
	sort := c.QueryParam("sort")
	asc, _ := strconv.Atoi(c.QueryParam("asc"))

	if limit == 0 {
		limit = 15
	}
	if limit > 100 {
		limit = 100
	}
	if page <= 1 {
		page = 0
	} else {
		page--
	}
	ascFinal := "desc"
	if asc == 1 {
		ascFinal = "asc"
	}
	if sort != "" {
		sort = ToSnakeCase(sort)
		sort = fmt.Sprintf(`"%s" %s`, sort, ascFinal)
	}
	return Pagination{
		Limit:  limit,
		Page:   page,
		Sort:   sort,
		Search: c.QueryParam("search"),
		Field:  ToSnakeCase(c.QueryParam("field")),
	}
}

// Paginate handles pagination and returns the result and metadata
func Paginate(pagination Pagination, qry *gorm.DB, total int64) (*gorm.DB, H) {
	offset := pagination.Page * pagination.Limit
	qryData := qry.Limit(pagination.Limit).Offset(offset)
	if pagination.Sort != "" {
		qryData = qryData.Order(pagination.Sort)
	}
	return qryData, PaginateInfo(pagination, total)
}

// PaginateInfo generates pagination metadata
func PaginateInfo(paging Pagination, totalData int64) H {
	totalPages := math.Ceil(float64(totalData) / float64(paging.Limit))
	nextPage := paging.Page + 1
	if paging.Page >= int(totalPages)-1 {
		nextPage = 0
	}
	previousPage := paging.Page - 1
	if previousPage < 0 {
		previousPage = 0
	}
	return H{
		"nextPage":     nextPage,
		"previousPage": previousPage,
		"currentPage":  paging.Page + 1,
		"totalPages":   totalPages,
		"totalData":    totalData,
	}
}

// ToSnakeCase converts camelCase to snake_case
func ToSnakeCase(camel string) string {
	var result []byte
	for i, c := range camel {
		if unicode.IsUpper(c) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, byte(unicode.ToLower(c)))
		} else {
			result = append(result, byte(c))
		}
	}
	return string(result)
}

func Includes(haystack []interface{}, needle interface{}) bool {
	for _, sliceItem := range haystack {
		if sliceItem == needle {
			return true
		}
	}
	return false
}