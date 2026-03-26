package helpers

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

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

func FirstRowFromResponse(raw interface{}) (map[string]interface{}, error) {
	if m, ok := raw.(map[string]interface{}); ok {
		if d, ok := m["Data"]; ok {
			switch dd := d.(type) {
			case []interface{}:
				if len(dd) == 0 {
					return nil, fmt.Errorf("respuesta sin datos")
				}
				row, ok := dd[0].(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("Data[0] no es objeto")
				}
				return row, nil
			case map[string]interface{}:
				return dd, nil
			default:
				return nil, fmt.Errorf("formato Data no soportado: %T", d)
			}
		}
		return m, nil
	}

	if arr, ok := raw.([]interface{}); ok {
		if len(arr) == 0 {
			return nil, fmt.Errorf("respuesta sin datos")
		}
		row, ok := arr[0].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("[0] no es objeto")
		}
		return row, nil
	}

	return nil, fmt.Errorf("formato de respuesta no soportado: %T", raw)
}
