package melody

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func (b *Builder) WithQueryParams(model interface{}, qp map[string]string) *Builder {
	var page, perPage int
	var err error

	b.parseStruct(model)

	for key, value := range qp {
		switch key {
		case "page":
			page, err = strconv.Atoi(value)
			if err != nil {
				panic(err)
			}
			break
		case "per_page":
			perPage, err = strconv.Atoi(value)
			if err != nil {
				panic(err)
			}
			break
		default:
			err = b.parseQueryParam(key, value)
			if err != nil {
				panic(err)
			}
		}
	}

	if perPage != 0 {
		b.Limit(perPage)
		if page != 0 {
			b.Offset(perPage * (page - 1))
		}
	}

	return b
}

func (b *Builder) parseQueryParam(key string, value string) error {
	value = strings.ReplaceAll(value, "%20", " ")
	rgx := regexp.MustCompile(`^(.*)%5B(.*)%5D$`)

	json := key
	cond := "equal"

	indexes := rgx.FindAllStringSubmatch(key, -1)
	if len(indexes) == 1 && len(indexes[0]) == 3 {
		json = indexes[0][1]
		cond = indexes[0][2]
	}

	field := b.ctx.JsonToDB[json]
	//if field == "" {
	//
	//}

	switch strings.ToLower(cond) {
	case "equal": // =23
		b.Where(field, "=", value)
		break
	case "not-equal": // [not-equal]=23
		b.Where(field, "!=", value)
		break
	case "like": // [like]=23
		b.Where(field, "LIKE", "%"+value+"%")
		break
	case "not-like": // [not-like]=23
		b.Where(field, "NOT LIKE", "%"+value+"%")
		break
	case "in", "not-in": // [in]=1,2 or [not-in]=1,2
		array := strings.Split(value, ",")

		s := make([]interface{}, len(array))
		for i, v := range array {
			s[i] = v
		}

		operator := "IN"
		if strings.ToLower(cond) == "not-in" {
			operator = "NOT IN"
		}

		b.Where(field, operator, s...)
		break
	}

	return nil
}

func (b *Builder) parseStruct(st interface{}) {
	jsonToDb := make(map[string]string)

	parseStruct(st, jsonToDb, "")

	b.ctx.JsonToDB = jsonToDb
}

func parseStruct(st interface{}, data map[string]string, sub string) {
	t := reflect.TypeOf(st)
	val := reflect.ValueOf(st)

	for i := 0; i < t.NumField(); i++ {
		s := t.Field(i)
		json := strings.ReplaceAll(s.Tag.Get("json"), ",omitempty", "")
		sql := s.Tag.Get("db")

		if s.Type.Kind().String() == "struct" {
			parseStruct(val.Field(i).Interface(), data, sub+json+".")
		}

		if json != "-" && json != "" && sql != "" {
			data[sub+json] = sql
		}
	}
}
