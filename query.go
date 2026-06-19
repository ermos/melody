package melody

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var timeType = reflect.TypeOf(time.Time{})

func (b *Builder) WithQueryParams(model interface{}, qp map[string]string) (*Builder, error) {
	var page, perPage int
	var err error

	b.parseStruct(model)

	for key, value := range qp {
		switch key {
		case "page":
			page, err = strconv.Atoi(value)
			if err != nil {
				return b, fmt.Errorf("invalid page %q: %w", value, err)
			}
		case "per_page":
			perPage, err = strconv.Atoi(value)
			if err != nil {
				return b, fmt.Errorf("invalid per_page %q: %w", value, err)
			}
		default:
			if err = b.parseQueryParam(key, value); err != nil {
				return b, err
			}
		}
	}

	if perPage != 0 {
		b.Limit(perPage)
		if page != 0 {
			b.Offset(perPage * (page - 1))
		}
	}

	return b, nil
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

	field, ok := b.ctx.JsonToDB[json]
	if !ok {
		return fmt.Errorf("unknown query param %q", json)
	}

	switch strings.ToLower(cond) {
	case "equal": // =23
		b.Where(field, "=", value)
	case "not-equal": // [not-equal]=23
		b.Where(field, "!=", value)
	case "like": // [like]=23
		b.Where(field, "LIKE", "%"+value+"%")
	case "not-like": // [not-like]=23
		b.Where(field, "NOT LIKE", "%"+value+"%")
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
	default:
		return fmt.Errorf("unknown condition %q", cond)
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

	for t != nil && t.Kind() == reflect.Ptr {
		if val.IsNil() {
			return
		}
		t = t.Elem()
		val = val.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		s := t.Field(i)
		if s.PkgPath != "" { // unexported field
			continue
		}

		json := strings.ReplaceAll(s.Tag.Get("json"), ",omitempty", "")
		sql := s.Tag.Get("db")

		// A field carrying a db tag is a column, even when its Go type is a
		// struct (e.g. time.Time): map it and do not recurse into it.
		if json != "-" && json != "" && sql != "" {
			data[sub+json] = sql
			continue
		}

		ft := s.Type
		for ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		if ft.Kind() == reflect.Struct && ft != timeType {
			prefix := sub
			if json != "" && json != "-" {
				prefix = sub + json + "."
			}
			parseStruct(val.Field(i).Interface(), data, prefix)
		}
	}
}
