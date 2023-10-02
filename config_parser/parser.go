package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

type routeError struct {
	Key         string
	Description string
}

func (e *routeError) Error() string {
	return fmt.Sprintf("error in route `%s`: %s", e.Key, e.Description)
}

type ResponseConfig struct {
	Code int `yaml:"code"`
	Body *struct {
		ContentType string `yaml:"content_type,omitempty"`
		Echo        bool   `yaml:"echo,omitempty"`
		Value       any    `yaml:"value,omitempty"`
	} `yaml:"body"`
}

type RouteConfig map[string]ResponseConfig

// ConfigFile parsed config
type ConfigFile map[string]RouteConfig

// ParseConfig Parse config file from CONFIG_FILE env param ti ConfigFile struct
func ParseConfig(fileName string) *ConfigFile {
	res, err := parse(fileName)
	if err != nil {
		panic(err)
	}
	return res
}

const (
	GET  = http.MethodGet
	POST = http.MethodPost
	DEL  = http.MethodDelete
	PUT  = http.MethodPut
	DEF  = "DEFAULT"
)

var supportedMethods = []string{GET, POST, DEL, PUT, DEF}

const (
	TEXT = "text/plain"
	JSON = "application/json"
)

var supportedTypes = []string{TEXT, JSON}

func parse(fileName string) (*ConfigFile, error) {
	var (
		err          error
		file         *os.File
		data         []byte
		encodedValue []byte
	)

	file, err = os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err = io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	result := ConfigFile{}
	if err = yaml.Unmarshal(data, result); err != nil {
		return nil, fmt.Errorf("wrong schema")
	}

	// Clear unfilled route
	for k, routes := range result {
		if len(routes) == 0 {
			delete(result, k)
			continue
		}

		if !strings.HasPrefix(k, "/") {
			return nil, &routeError{
				Key:         k,
				Description: "route key must start with `/`",
			}
		}

		for method, route := range routes {
			uppercaseMethod := strings.ToUpper(method)
			if !slices.Contains(supportedMethods, uppercaseMethod) {
				return nil, &routeError{
					Key:         k,
					Description: fmt.Sprintf("unsupported method `%s`", method),
				}
			}

			if route.Body == nil || route.Code == 0 {
				delete(result, k)
				break
			}

			delete(routes, method)
			routes[uppercaseMethod] = route

			if route.Body.Echo {
				continue
			}

			if !slices.Contains(supportedTypes, route.Body.ContentType) {
				return nil, &routeError{
					Key: k,
					Description: fmt.Sprintf(
						"unsupported Content Type `%s`", route.Body.ContentType,
					),
				}
			}

			switch route.Body.ContentType {
			case TEXT:
				val, ok := route.Body.Value.(string)
				if !ok {
					return nil, &routeError{
						Key:         k,
						Description: "value on `text/plain` must be string",
					}
				}
				(*route.Body).Value = []byte(val)

			case JSON:
				val, ok := route.Body.Value.(string)
				if ok {
					d := new(interface{})
					err = json.Unmarshal([]byte(val), d)
					if err != nil {
						return nil, &routeError{
							Key:         k,
							Description: err.Error(),
						}
					}
					(*route.Body).Value = []byte(val)

				} else {
					encodedValue, err = json.Marshal(route.Body.Value)
					if err != nil {
						return nil, &routeError{
							Key:         k,
							Description: err.Error(),
						}
					}
					(*route.Body).Value = encodedValue
				}
			}

		}
	}

	return &result, nil
}
