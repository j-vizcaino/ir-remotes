// +build !embedded

package config

import "net/http"

var Assets http.FileSystem = http.Dir(".")
