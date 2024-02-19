package internal

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"

	"github.com/3WDeveloper-GM/library_app/backend/internal/validator"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

func (f *Filters) ValidateFilters(v *validator.Validator) bool {

	var section = "page"
	var minPage = 0
	var maxPage = 1000000
	var fieldGreaterThanMsg = "%s field must be larger than %d"
	var fieldLessThanMsg = "%s field must be less than %d"

	v.Check(f.Page > minPage, section, fmt.Sprintf(fieldGreaterThanMsg, section, minPage))
	v.Check(f.Page <= maxPage, section, fmt.Sprintf(fieldLessThanMsg, section, maxPage))

	section = "page_size"
	var minPageSize = 0
	var maxPageSize = 30

	v.Check(f.Page > minPage, section, fmt.Sprintf(fieldGreaterThanMsg, section, minPageSize))
	v.Check(f.Page <= maxPage, section, fmt.Sprintf(fieldLessThanMsg, section, maxPageSize))

	section = "sort"
	v.Check(v.In(f.Sort, f.SortSafeList), section, fmt.Sprintf("invalid %s parameter", section))

	return v.Valid()
}

func DiffArrays(oldArr, newArr []string) ([]string, []string, []string) {
	common := make([]string, 0)
	exclusiveOld := make([]string, 0)
	exclusiveNew := make([]string, 0)

	// Create a map to store elements from the old array
	oldMap := make(map[string]bool)
	for _, v := range oldArr {
		oldMap[v] = true
	}

	// Iterate over elements in the new array
	for _, v := range newArr {
		if _, ok := oldMap[v]; ok {
			// Common element found
			common = append(common, v)
			// Remove the element from the oldMap to identify exclusive elements later
			delete(oldMap, v)
		} else {
			// Element exclusive to the new array
			exclusiveNew = append(exclusiveNew, v)
		}
	}

	// Elements remaining in oldMap are exclusive to the old array
	for k := range oldMap {
		exclusiveOld = append(exclusiveOld, k)
	}

	return common, exclusiveOld, exclusiveNew
}

func HashArrays(arr1, arr2 []string) ([]string, []string) {
	hashArr1 := make([]string, len(arr1))
	hashArr2 := make([]string, len(arr2))

	for i, v := range arr1 {
		hashArr1[i] = hashString(v)
	}

	for i, v := range arr2 {
		hashArr2[i] = hashString(v)
	}

	return hashArr1, hashArr2
}

func hashString(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	hashBytes := h.Sum(nil)
	return hex.EncodeToString(hashBytes)
}
