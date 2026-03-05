package helpers

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
)

func GetBaseURL(appKey, envKey string) (string, error) {
	v := strings.TrimSpace(beego.AppConfig.String(appKey))
	if v == "" || strings.Contains(v, "${") {
		v = strings.TrimSpace(os.Getenv(envKey))
	}
	if v == "" {
		return "", fmt.Errorf("no está configurado %s ni %s", appKey, envKey)
	}
	if !strings.HasPrefix(v, "http://") && !strings.HasPrefix(v, "https://") {
		v = "http://" + v
	}
	return strings.TrimRight(strings.TrimSpace(v), "/"), nil
}

func JoinURL(base, path string) string {
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(path, "/")
}

func ValidateAbsoluteURL(u string) error {
	parsed, err := url.Parse(u)
	if err != nil {
		return fmt.Errorf("URL inválida: %v", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("URL inválida (sin scheme/host): %s", u)
	}
	return nil
}

func UnwrapDataToMap(resp map[string]interface{}) map[string]interface{} {
	if resp == nil {
		return nil
	}
	if raw, ok := resp["Data"]; ok {
		switch d := raw.(type) {
		case []interface{}:
			if len(d) > 0 {
				if m, ok := d[0].(map[string]interface{}); ok {
					return m
				}
			}
		case map[string]interface{}:
			return d
		}
	}
	if _, ok := resp["Id"]; ok {
		return resp
	}
	return nil
}

func ExtractIdAtoi(resp map[string]interface{}) int {
	// retorna 0 si no puede convertir
	if resp == nil {
		return 0
	}
	if raw, ok := resp["Data"]; ok {
		switch d := raw.(type) {
		case map[string]interface{}:
			id, err := strconv.Atoi(fmt.Sprintf("%v", d["Id"]))
			if err == nil {
				return id
			}
		case []interface{}:
			if len(d) > 0 {
				if m, ok := d[0].(map[string]interface{}); ok {
					id, err := strconv.Atoi(fmt.Sprintf("%v", m["Id"]))
					if err == nil {
						return id
					}
				}
			}
		}
	}
	id, err := strconv.Atoi(fmt.Sprintf("%v", resp["Id"]))
	if err == nil {
		return id
	}
	return 0
}
