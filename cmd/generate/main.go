package main

import (
	_ "embed"
	"fmt"
	applicationbuilder "go_tcc_drm_wrapper/internal/application_builder"
	wrappertemplate "go_tcc_drm_wrapper/internal/wrapper_template"
	"os"
)

// test app name test_game.x86_64

func main() {
	if len(os.Args) < 5 {
		fmt.Println("Usage: go run main.go <mintId> <applicationPath> <execName> <chain>")
		os.Exit(1)
	}

	mintId := os.Args[1]
	applicationPath := os.Args[2]
	execName := os.Args[3]
	chain := os.Args[4]

	wrapperBuilder := wrappertemplate.NewWrapperTemplateBuilder()
	wrapperBuilder.SetMintId(mintId)
	wrapperBuilder.SetChain(chain)

	appBuilder := applicationbuilder.New()
	appBuilder.SetExecName(execName)
	appBuilder.SetOutputPath("C:\\Users\\rafael\\Documents\\GitHub\\go_tcc_drm_wrapper\\output")
	appBuilder.SetZipPath(applicationPath)
	appBuilder.SetWrapperBuilder(wrapperBuilder)

	err := appBuilder.BuildApplication()

	if err != nil {
		fmt.Println("Error building application: ", err)
		os.Exit(1)
	}

}
