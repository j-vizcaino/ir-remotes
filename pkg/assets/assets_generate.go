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
		assetsDir := fmt.Sprintf("../../assets/%s", dir)
		err := vfsgen.Generate(http.Dir(assetsDir), vfsgen.Options{
			Filename: fmt.Sprintf("%s/assets_vfsdata.go", dir),
			PackageName:  dir,
			BuildTags:    "embedded",
			VariableName: "Assets",
		})
		if err != nil {
			log.Fatalln(err)
		}
	}
}
