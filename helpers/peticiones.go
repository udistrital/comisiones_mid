package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astaxie/beego"
)

func GetJsonTest(url string, target interface{}) (status int, err error) {
	r, err := http.Get(url)
	if err != nil {
		return r.StatusCode, err
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			beego.Error(err)
		}
	}()

	return r.StatusCode, json.NewDecoder(r.Body).Decode(target)
}

func PostJsonTest(url string, data interface{}, target interface{}) (status int, err error) {

	body := new(bytes.Buffer)
	fmt.Println(body)
	if data != nil {
		if err = json.NewEncoder(body).Encode(data); err != nil {
			return 0, err
		}
	}

	r, err := http.Post(url, "application/json", body)
	if err != nil {
		return 0, err
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			beego.Error(err)
		}
	}()
	fmt.Println(r.StatusCode)
	fmt.Println(json.NewDecoder(r.Body))
	return r.StatusCode, json.NewDecoder(r.Body).Decode(target)
}
