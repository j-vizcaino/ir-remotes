// +build ignore

package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/shurcooL/vfsgen"
)

func main() {
	for _, dir:= range []string{"ui", "config"}{
		err := vfsgen.Generate(http.Dir(dir), vfsgen.Options{
			Filename: fmt.Sprintf("../pkg/assets/%s/assets_vfsdata.go", dir),
			PackageName:  dir,
			BuildTags:    "embedded",
			VariableName: "Assets",
		})
		if err != nil {
			log.Fatalln(err)
		}
	}
}
