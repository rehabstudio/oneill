package config

import "reflect"

const (
	defaultDefinitionsURI         string = "file:///etc/oneill/definitions"
	defaultDockerApiEndpoint      string = "unix:///var/run/docker.sock"
	defaultNginxConfigDirectory   string = "/etc/nginx/sites-enabled"
	defaultNginxHtpasswdDirectory string = "/etc/nginx/htpasswd"
	defaultServingDomain          string = "example.com"
	defaultLogLevel               int    = 2
)

// isZero tests that a given value is the zero value for its type, this is
// used to decide when to set default config values
//
// borrowed from go-yaml :)
func isZero(i interface{}) bool {
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.String:
		return len(v.String()) == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Slice:
		return v.Len() == 0
	case reflect.Map:
		return v.Len() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Struct:
		vt := v.Type()
		for i := v.NumField() - 1; i >= 0; i-- {
			if vt.Field(i).PkgPath != "" {
				continue // Private field
			}
			if !isZero(v.Field(i)) {
				return false
			}
		}
		return true
	}
	return false
}
