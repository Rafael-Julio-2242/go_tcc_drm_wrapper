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
	if len(os.Args) != 5 {
		fmt.Println("Usage: go run main.go <mintId> <ownerAddress> <applicationPath> <execName>")
		os.Exit(1)
	}

	mintId := os.Args[1]
	ownerAddress := os.Args[2]
	applicationPath := os.Args[3]
	execName := os.Args[4]

	wrapperBuilder := wrappertemplate.NewWrapperTemplateBuilder()
	wrapperBuilder.SetMintId(mintId)
	wrapperBuilder.SetOwnerAddress(ownerAddress)

	appBuilder := applicationbuilder.New()
	appBuilder.SetExecName(execName)
	appBuilder.SetOutputPath("/home/rafa/Documentos/GitHub/go_tcc_drm_wrapper/output")
	appBuilder.SetZipPath(applicationPath)
	appBuilder.SetWrapperBuilder(wrapperBuilder)

	err := appBuilder.BuildApplication()

	if err != nil {
		fmt.Println("Error building application: ", err)
		os.Exit(1)
	}

}
