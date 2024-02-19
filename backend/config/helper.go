package config

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/3WDeveloper-GM/library_app/backend/internal/validator"
)

func (app *App) ReadString(qs url.Values, key, deafultValue string) string {

	s := qs.Get(key)

	if s == "" {
		return deafultValue
	}

	return s
}

func (app *App) ReadCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)

	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

func (app *App) ReadInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {

	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return i
}

func (app *App) ToUpper(str string) string {
	new_str := []byte(str)
	for i := 0; i < len(str); i++ {
		if str[i] >= 'a' && str[i] <= 'z' {
			chr := uint8(rune(str[i]) - 'a' + 'A')
			new_str[i] = chr
		}
	}
	return string(new_str)
}

func (app *App) FindDiffOneInTwo(list1, list2 []string) []string {
	set := make(map[string]struct{})
	for _, item := range list1 {
		set[item] = struct{}{}
	}

	var difference []string

	for _, item := range list2 {
		if _, ok := set[item]; !ok {
			difference = append(difference, item)
		}
	}
	return difference
}
