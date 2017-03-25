package himawari

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

func SearchCategories(large, middle string) ([]Category, error) {
	query := ""
	if len(large) > 0 {
		query += "large_category=" + url.QueryEscape(large)
		if len(middle) > 0 {
			query += "&"
		}
	}
	if len(middle) > 0 {
		query += "middle_category=" + url.QueryEscape(middle)
	}
	r, err := http.Get("http://172.31.125.100:8080/v1/categories/?" + query)
	// fmt.Println("http://172.31.125.100:8080/v1/categories/?" + query)
	if err != nil {
		fmt.Printf("http error: %#v %s\r\n", r, err.Error())
		return nil, err
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("io error: %#v\r\n", r)
		return nil, err
	}

	var categories []Category
	json.Unmarshal(body, &categories)

	return categories, nil
}
