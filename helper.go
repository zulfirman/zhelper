package zhelper

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v4"
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
)

var (
	poolReq = &sync.Pool{
		New: func() interface{} {
			return resty.New()
		},
	}
)

func Rs(c echo.Context, result Response) error {
	if result.Code == 0 {
		result.Code = 200
	}
	result.Status = http.StatusText(result.Code)
	result.Path = Substr(c.Request().RequestURI, 150)
	return c.JSON(result.Code, result)
}

func RsMessage(c echo.Context, code int, message interface{}) error {
	var result Response
	result.Code=code
	if result.Code == 0 {
		result.Code = 200
	}
	result.Content = H{
		"message" : message,
	}
	result.Status = http.StatusText(result.Code)
	result.Path = Substr(c.Request().RequestURI, 150)
	return c.JSON(result.Code, result)
}

func KeyExists(decoded map[string]interface{}, key string) bool {
	val, ok := decoded[key]
	return ok && val != nil
}

// H is a shortcut for map[string]interface{}
type H map[string]interface{}

func RsSuccess(c echo.Context) error {
	return Rs(c, Response{
		Content: H{
			"message": "success",
		},
	})
}

func RsError(c echo.Context, code int, message interface{}) error {
	return Rs(c, Response{
		Code:    code,
		Content: message,
	})
}

func ReadyBodyJson(c echo.Context, json map[string]interface{}) map[string]interface{} {
	if err := c.Bind(&json); err != nil {
		return nil
	}
	return json
}

func GetReq(Url string, token string) (*resty.Response, error) {
	client := poolReq.Get().(*resty.Client)
	defer poolReq.Put(client)
	resp, err := client.R().EnableTrace().SetAuthToken(token).Get(Url)
	if err != nil {
		fmt.Println(err)
	}
	code := resp.StatusCode()
	if code != 200 {
		fmt.Println(resp.String())
		err = errors.New("response code is " + string(code))
	}
	return resp, err
}

func PostReq(Url string, token string, body interface{}) (*resty.Response, error) {
	client := poolReq.Get().(*resty.Client)
	defer poolReq.Put(client)
	resp, err := client.R().EnableTrace().
		SetAuthToken(token).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(Url)
	if err != nil {
		fmt.Println(err)
	}
	code := resp.StatusCode()
	if code != 200 {
		fmt.Println(resp.String())
		err = errors.New("response code is " + string(code))
	}
	return resp, err
}

func JsonToMap(jsonStr string) map[string]interface{} {
	result := make(map[string]interface{})
	err := sonic.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil
	}
	return result
}

func MarshalBinary(i interface{}) (data []byte) { //bytes to interface
	marshal, err := sonic.Marshal(i)
	if err != nil {
		println(err.Error())
	}
	return marshal
}

func ReadJson(jsonString string, path string) interface{} { //from json string to json
	if path == "" {
		return gojsonq.New().FromString(jsonString).Get()
	}
	return gojsonq.New().FromString(jsonString).Find(path)
}

func RemoveField(obj interface{}, ignoreFields ...string) (interface{}, error) {
	// Marshal the object to JSON.
	toJson, err := sonic.Marshal(obj)
	if err != nil {
		return obj, err
	}

	// If no fields are specified, return the object as is.
	if len(ignoreFields) == 0 {
		return obj, nil
	}

	// Unmarshal the JSON to a map.
	toMap := map[string]interface{}{}
	sonic.Unmarshal(toJson, &toMap)

	// Remove the specified fields from the map.
	for _, field := range ignoreFields {
		delete(toMap, field)
	}

	// Return the modified map.
	return toMap, nil
}

func IntString(result int) string {
	return strconv.Itoa(result)
}

func StringInt(result string) int {
	intVar, _ := strconv.Atoi(result)
	return intVar
}

func Substr(input string, limit int) string {
	if len([]rune(input)) >= limit {
		input = input[0:limit]
	}
	return input
}

func ArrUniqueStr(strSlice []string) []string {
	// Use a map to track the unique elements.
	uniqueMap := make(map[string]bool)
	for _, item := range strSlice {
		if item == "" {
			// Skip empty strings.
			continue
		}
		uniqueMap[item] = true
	}
	// Convert the map keys to a slice.
	uniqueSlice := make([]string, 0, len(uniqueMap))
	for key := range uniqueMap {
		uniqueSlice = append(uniqueSlice, key)
	}
	return uniqueSlice
}

func ArrUniqueInt(intSlice []int) []int {
	uniqueMap := make(map[int]bool)
	for _, item := range intSlice {
		if item == 0 {
			continue
		}
		uniqueMap[item] = true
	}
	uniqueSlice := make([]int, 0, len(uniqueMap))
	for key := range uniqueMap {
		uniqueSlice = append(uniqueSlice, key)
	}
	return intSlice
}

func ArrUnique64(intSlice []int64) []int64 {
	uniqueMap := make(map[int64]bool)
	for _, item := range intSlice {
		if item == 0 {
			continue
		}
		uniqueMap[item] = true
	}
	uniqueSlice := make([]int64, 0, len(uniqueMap))
	for key := range uniqueMap {
		uniqueSlice = append(uniqueSlice, key)
	}
	return intSlice
}

func UniqueId() string {
	guid := xid.New()
	return guid.String()
}

func DateNow(typeFormat int) string {
	timeNow := time.Now()
	dateNow := time.Now().UTC()
	if typeFormat == 1 { //date only
		return dateNow.Format("2006-01-02")
	}
	if typeFormat == 2 { //time only
		return timeNow.Format("15:43:5")
	}
	if typeFormat == 3 { //datetime
		return dateNow.String()
	}
	return ""
}

func GormTime(timeParam time.Time) datatypes.Time {
	timeOnly := timeParam.Format("15:04:05")
	splitted := strings.Split(timeOnly, ":")
	return datatypes.NewTime(StringInt(splitted[0]), StringInt(splitted[1]), StringInt(splitted[2]), 0)
}

func DeletedAt() gorm.DeletedAt {
	return gorm.DeletedAt{
		Time:  time.Now(),
		Valid: true,
	}
}

func BlankString(s string) bool {
	// return true if whitespace
	// - The string cannot be empty.
	// - The string cannot contain only spaces.
	// - The string cannot start with a space.
	if s == "" || strings.TrimSpace(s) == "" || strings.HasPrefix(s, " ") {
		return true
	}
	return false
}

func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

// paginate helper
func GetParamPagination(c echo.Context) Pagination {
	// Get the query parameters from the request.
	query := c.Request().URL.Query()

	// Get the "limit", "page", "sort", and "asc" query parameters.
	// If the parameter is not present, the default value is returned.
	limit, _ := strconv.Atoi(query.Get("limit"))
	page, _ := strconv.Atoi(query.Get("page"))
	sort := query.Get("sort")
	asc, _ := strconv.Atoi(query.Get("asc"))

	// If the limit is not set or is set to 0, use a default value of 15.
	// If the limit is greater than 100, use a maximum value of 100.
	if limit == 0 {
		limit = 15
	}
	if limit > 100 {
		limit = 100
	}

	// If the page is not set or is set to 1, use a default value of 0.
	// Otherwise, decrement the page number by 1.
	if page <= 1 {
		page = 0
	} else {
		page--
	}

	// Set the sort order to "desc" by default.
	// If the "asc" query parameter is set to 1, set the sort order to "asc".
	ascFinal := "desc"
	if asc == 1 {
		ascFinal = "asc"
	}

	// If the "sort" query parameter is set, format it as `"field_name" order`.
	if sort != "" {
		sort = ToSnakeCase(sort)
		sort = `"` + sort + `" ` + ascFinal
	}

	// Return the pagination parameters as a Pagination struct.
	return Pagination{
		Limit:  limit,
		Page:   page,
		Sort:   sort,
		Search: c.QueryParam("search"),
		Field:  ToSnakeCase(c.QueryParam("field")),
	}
}

func Paginate(pagination Pagination, qry *gorm.DB, total int64) (*gorm.DB, H) {
	offset := (pagination.Page) * pagination.Limit
	qryData := qry.Limit(pagination.Limit).Offset(offset)
	if pagination.Sort == "\"\" asc" {
		return qryData, PaginateInfo(pagination, total)
	}
	return qryData.Order(pagination.Sort), PaginateInfo(pagination, total)
}

func PaginateInfo(paging Pagination, totalData int64) H {
	// Calculate the total number of pages.
	totalPages := math.Ceil(float64(totalData) / float64(paging.Limit))

	// Calculate the next and previous page numbers.
	nextPage := paging.Page + 1
	if nextPage >= int(totalPages) {
		nextPage = 0
	}
	previousPage := paging.Page - 1
	if previousPage < 1 {
		previousPage = 0
	}

	// Increment the current page number.
	// If the current page is less than 1, set it to 1.
	paging.Page++
	if paging.Page < 1 {
		paging.Page = 1
	}

	// Return the pagination information as a map.
	return H{
		"nextPage":     nextPage,
		"previousPage": previousPage,
		"currentPage":  paging.Page,
		"totalPages":   totalPages,
		"totalData":    totalData,
	}
}

func ToSnakeCase(camel string) string {
	// Preallocate a slice of bytes with enough capacity to hold the final string.
	// This will avoid additional memory allocations and string copies when building the output string.
	buf := make([]byte, 0, len(camel)+5)

	// Iterate through the runes in the input string.
	for i := 0; i < len(camel); i++ {
		c := camel[i]
		if c >= 'A' && c <= 'Z' {
			// If the current rune is an uppercase letter, insert an underscore and convert it to lowercase.
			if len(buf) > 0 {
				buf = append(buf, '_')
			}
			buf = append(buf, c-'A'+'a')
		} else {
			// Otherwise, just append the current rune as is.
			buf = append(buf, c)
		}
	}
	return string(buf)
}

func MeValidate(c echo.Context) (map[string]interface{}, error) { //check if token is valid then parse the token to struct
	tokenString := c.Request().Header.Get("Authorization")
	tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
	claims := jwt.MapClaims{}
	_, errClaim := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if jwt.SigningMethodHS256 != token.Method {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		secret := ""
		return []byte(secret), nil
	})
	if errClaim != nil {
		println(errClaim)
	}
	ch := make(chan map[string]interface{})
	go flattenJSON(claims, "", ch)
	flattenedData := <-ch
	return flattenedData, nil
}

func Me(c echo.Context) (map[string]interface{}, error) { //parse jwt token
	tokenString := c.Request().Header.Get("Authorization")
	tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, errors.New("cannot parse token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("cannot parse token")
	}
	var claims map[string]interface{}
	sonic.Unmarshal(payload, &claims)
	ch := make(chan map[string]interface{})
	go flattenJSON(claims, "", ch)
	flattenedData := <-ch
	return flattenedData, nil
}

func flattenJSON(data map[string]interface{}, parentKey string, ch chan<- map[string]interface{}) {
	// make a map to keep track of the keys that have been added
	keys := make(map[string]bool)

	for key, value := range data {
		// construct the new key by concatenating the parent key and current key
		newKey := parentKey + key

		// check if the value is of type map[string]interface{}
		if _, ok := value.(map[string]interface{}); ok {
			// create a new channel for communicating with the nested function call
			ch := make(chan map[string]interface{})
			// recursively call the flattenJSON function for the nested map
			go flattenJSON(value.(map[string]interface{}), newKey+".", ch)
			flattenedData := <-ch

			// iterate through the flattened data
			for k, v := range flattenedData {
				// check if the key already exists in the data map
				if !KeyExists(data, k) {
					data[k] = v
					keys[k] = true
				}
			}
			// remove the nested map from the data map
			delete(data, key)
		}
	}
	ch <- data
}

func HasTrustee(trusteeValue interface{}, values []string) bool {
	arr := trusteeValue.([]interface{})
	var found = false
	for _, v := range arr {
		if v, ok := v.(string); ok {
			for _, value := range values {
				if v == value {
					found = true
					break
				}
			}
		}
	}
	if found {
		return found
	}
	return found
}

func Includes(haystack []interface{}, needle interface{}) bool {
	for _, sliceItem := range haystack {
		if sliceItem == needle {
			return true
		}
	}
	return false
}

/*func ToSnakeCaseV1(camel string) string {
	var buf bytes.Buffer
	for _, c := range camel {
		if 'A' <= c && c <= 'Z' {
			// just convert [A-Z] to _[a-z]
			if buf.Len() > 0 {
				buf.WriteRune('_')
			}
			buf.WriteRune(c - 'A' + 'a')
		} else {
			buf.WriteRune(c)
		}
	}
	return buf.String()
}*/
//end paginate helper
