package zhelper

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo/v4"
	"github.com/rs/xid"
	"github.com/thedevsaddam/gojsonq/v2"
	"github.com/zulfirman/zhelper/model"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func Rs(c echo.Context, Ct map[string]interface{}) error {
	var Return model.Response
	if KeyExists(Ct, "code") {
		Return.Code = Ct["code"].(int)
	} else {
		Return.Code = 200
	}
	if KeyExists(Ct, "status") {
		if Ct["status"] == "" {
			Return.Status = http.StatusText(Return.Code)
		} else {
			Return.Status = fmt.Sprintf("%v", Ct["status"])
		}
	} else {
		Return.Status = http.StatusText(Return.Code)
	}

	if KeyExists(Ct, "content") {
		Return.Content = Ct["content"]
	} else {
		Return.Content = nil
	}
	if KeyExists(Ct, "other") {
		Return.Others = Ct["other"]
	} else {
		Return.Others = nil
	}
	Return.Path = Substr(c.Request().RequestURI, 150)
	return c.JSON(Return.Code, Return)
}

func RsSuccess(c echo.Context) error {
	return Rs(c, H{
		"content": "success",
	})
}

func RsError(c echo.Context, code int, message interface{}) error {
	return Rs(c, H{
		"status":  code,
		"message": message,
	})
}

func KeyExists(decoded map[string]interface{}, key string) bool {
	val, ok := decoded[key]
	return ok && val != nil
}

// H is a shortcut for map[string]interface{}
type H map[string]interface{}

func ReadyBodyJson(c echo.Context, json map[string]interface{}) map[string]interface{} {
	if err := c.Bind(&json); err != nil {
		return nil
	}
	return json
}

// MarshalXML allows type H to be used with xml.Marshal.
func (h H) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{
		Space: "",
		Local: "map",
	}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for key, value := range h {
		elem := xml.StartElement{
			Name: xml.Name{Space: "", Local: key},
			Attr: []xml.Attr{},
		}
		if err := e.EncodeElement(value, elem); err != nil {
			return err
		}
	}

	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

func GetReq(Url string, token string) (*resty.Response, error) {
	client := resty.New()
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
	client := resty.New()
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
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil
	}
	return result
}

func MarshalBinary(i interface{}) (data []byte) { //bytes to json string
	marshal, err := json.Marshal(i)
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
	toJson, err := json.Marshal(obj)
	if err != nil {
		return obj, err
	}
	if len(ignoreFields) == 0 {
		return obj, nil
	}
	toMap := map[string]interface{}{}
	json.Unmarshal(toJson, &toMap)
	for _, field := range ignoreFields {
		delete(toMap, field)
	}
	return toMap, nil
}
func ErrorRoute(code int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		/*Rs(w, r, H{
			"status": code,
			"content": H{
				"Error": IntString(code) + " Not Found",
			},
		})*/
	}
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
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			if item == "" {
				continue
			}
			list = append(list, item)
		}
	}
	return list
}
func ArrUniqueInt(intSlice []int) []int {
	allKeys := make(map[int]bool)
	list := []int{}
	for _, item := range intSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
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
func BlankString(stringText string) bool {
	if stringText == "" {
		return true
	}
	count := 0
	for _, v := range stringText {
		if v == ' ' {
			count++
		} else {
			break
		}
	}
	if count > 0 {
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
func GetParamPagination(c echo.Context) model.Pagination {
	// Initializing default
	limit := 15
	page := 1
	//sort := `"DateAdded" asc`
	sort := ""
	asc := 1
	query := c.Request().URL.Query()
	for key, value := range query {
		queryValue := value[len(value)-1]
		switch key {
		case "limit":
			limit, _ = strconv.Atoi(queryValue)
			break
		case "page":
			page, _ = strconv.Atoi(queryValue)
			break
		case "sort":
			sort = queryValue
			break
		case "asc":
			asc, _ = strconv.Atoi(queryValue)
			break
		}
	}
	if page <= 1 { //page 0 or 1 means start from beginning
		page = 0
	}
	if page > 1 {
		page = page - 1
	}
	ascFinal := "desc"
	if asc == 1 {
		ascFinal = "asc"
	}
	if sort != "" {
		sort = ToSnakeCase(sort)
		sort = `"` + sort + `" ` + ascFinal
	}
	if limit == 0 {
		limit = 15
	}
	if limit > 100 {
		limit = 100
	}
	return model.Pagination{
		Limit:  limit,
		Page:   page,
		Sort:   sort,
		Search: c.QueryParam("search"),
	}
}

func Paginate(c echo.Context, qry *gorm.DB, total int64) (*gorm.DB, H) {
	pagination := GetParamPagination(c)
	offset := (pagination.Page) * pagination.Limit
	qryData := qry.Limit(pagination.Limit).Offset(offset)
	if pagination.Sort == "\"\" asc" {
		return qryData, PaginateInfo(pagination, total)
	}
	return qryData.Order(pagination.Sort), PaginateInfo(pagination, total)
}

func PaginateInfo(paging model.Pagination, totalData int64) H {
	totalPages := math.Ceil(float64(totalData) / float64(paging.Limit))
	paging.Page = paging.Page + 1
	nextPage := paging.Page + 1
	if paging.Page < 1 {
		paging.Page = 1
	}
	if paging.Page >= int(totalPages) {
		nextPage = 0
	}
	previousPage := paging.Page - 1
	if previousPage < 1 {
		previousPage = 0
	}
	paginationInfo := H{
		"nextPage":     nextPage,
		"previousPage": previousPage,
		"currentPage":  paging.Page,
		"totalPages":   totalPages,
		"totalData":    totalData,
	}
	return paginationInfo
}
func ToSnakeCase(camel string) string {
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
}

//end paginate helper
