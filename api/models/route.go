package models

import (
	"errors"
	"net/http"
	"net/url"
	"path"
	"strings"

	apiErrors "github.com/go-openapi/errors"
)

const (
	defaultRouteTimeout = 30 // seconds
)

var (
	ErrInvalidPayload      = errors.New("Invalid payload")
	ErrRoutesAlreadyExists = errors.New("Route already exists")
	ErrRoutesCreate        = errors.New("Could not create route")
	ErrRoutesGet           = errors.New("Could not get route from datastore")
	ErrRoutesList          = errors.New("Could not list routes from datastore")
	ErrRoutesMissingNew    = errors.New("Missing new route")
	ErrRoutesNotFound      = errors.New("Route not found")
	ErrRoutesPathImmutable = errors.New("Could not update route - path is immutable")
	ErrRoutesRemoving      = errors.New("Could not remove route from datastore")
	ErrRoutesUpdate        = errors.New("Could not update route")
)

type Routes []*Route

type Route struct {
	AppName        string      `json:"app_name"`
	Path           string      `json:"path"`
	Image          string      `json:"image"`
	Memory         uint64      `json:"memory"`
	Headers        http.Header `json:"headers"`
	Type           string      `json:"type"`
	Format         string      `json:"format"`
	MaxConcurrency int         `json:"max_concurrency"`
	Timeout        int32       `json:"timeout"`
	Config         `json:"config"`
}

var (
	ErrRoutesValidationFoundDynamicURL = errors.New("Dynamic URL is not allowed")
	ErrRoutesValidationInvalidPath     = errors.New("Invalid Path format")
	ErrRoutesValidationInvalidType     = errors.New("Invalid route Type")
	ErrRoutesValidationInvalidFormat   = errors.New("Invalid route Format")
	ErrRoutesValidationMissingAppName  = errors.New("Missing route AppName")
	ErrRoutesValidationMissingImage    = errors.New("Missing route Image")
	ErrRoutesValidationMissingName     = errors.New("Missing route Name")
	ErrRoutesValidationMissingPath     = errors.New("Missing route Path")
	ErrRoutesValidationMissingType     = errors.New("Missing route Type")
	ErrRoutesValidationPathMalformed   = errors.New("Path malformed")
	ErrRoutesValidationNegativeTimeout = errors.New("Negative timeout")
)

func (r *Route) Validate() error {
	var res []error

	if r.Memory == 0 {
		r.Memory = 128
	}

	if r.AppName == "" {
		res = append(res, ErrRoutesValidationMissingAppName)
	}

	if r.Path == "" {
		res = append(res, ErrRoutesValidationMissingPath)
	}

	u, err := url.Parse(r.Path)
	if err != nil {
		res = append(res, ErrRoutesValidationPathMalformed)
	}

	if strings.Contains(u.Path, ":") {
		res = append(res, ErrRoutesValidationFoundDynamicURL)
	}

	if !path.IsAbs(u.Path) {
		res = append(res, ErrRoutesValidationInvalidPath)
	}

	if r.Type == TypeNone {
		r.Type = TypeSync
	}

	if r.Type != TypeAsync && r.Type != TypeSync {
		res = append(res, ErrRoutesValidationInvalidType)
	}

	if r.Format == "" {
		r.Format = FormatDefault
	}

	if r.Format != FormatDefault && r.Format != FormatHTTP {
		res = append(res, ErrRoutesValidationInvalidFormat)
	}

	if r.MaxConcurrency == 0 {
		r.MaxConcurrency = 1
	}

	if r.Headers == nil {
		r.Headers = http.Header{}
	}

	if r.Config == nil {
		r.Config = map[string]string{}
	}

	if r.Timeout == 0 {
		r.Timeout = defaultRouteTimeout
	} else if r.Timeout < 0 {
		res = append(res, ErrRoutesValidationNegativeTimeout)
	}

	if len(res) > 0 {
		return apiErrors.CompositeValidationError(res...)
	}

	return nil
}

func (r *Route) Clone() *Route {
	var clone Route
	clone.AppName = r.AppName
	clone.Path = r.Path
	clone.Update(r)
	return &clone
}

// Update updates fields in r with non-zero field values from new.
// 0-length slice Header values, and empty-string Config values trigger removal of map entry.
func (r *Route) Update(new *Route) {
	if new.Image != "" {
		r.Image = new.Image
	}
	if new.Memory != 0 {
		r.Memory = new.Memory
	}
	if new.Type != "" {
		r.Type = new.Type
	}
	if new.Timeout != 0 {
		r.Timeout = new.Timeout
	}
	if new.Format != "" {
		r.Format = new.Format
	}
	if new.MaxConcurrency != 0 {
		r.MaxConcurrency = new.MaxConcurrency
	}
	if new.Headers != nil {
		if r.Headers == nil {
			r.Headers = make(http.Header)
		}
		for k, v := range new.Headers {
			if len(v) == 0 {
				r.Headers.Del(k)
			} else {
				for _, val := range v {
					r.Headers.Add(k, val)
				}
			}
		}
	}
	if new.Config != nil {
		if r.Config == nil {
			r.Config = make(Config)
		}
		for k, v := range new.Config {
			if v == "" {
				delete(r.Config, k)
			} else {
				r.Config[k] = v
			}
		}
	}
}

type RouteFilter struct {
	Path    string
	AppName string
	Image   string
}
