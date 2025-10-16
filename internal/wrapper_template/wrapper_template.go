package wrappertemplate

import (
	"errors"
	"fmt"
)

type WrapperTemplateBuilder struct {
	mintId          string
	ownerAddress    string
	applicationPath string
	applicationName string
}

func NewWrapperTemplateBuilder() *WrapperTemplateBuilder {
	return &WrapperTemplateBuilder{}
}

func (w *WrapperTemplateBuilder) SetMintId(mintId string) {
	w.mintId = mintId
}

func (w *WrapperTemplateBuilder) SetOwnerAddress(ownerAddress string) {
	w.ownerAddress = ownerAddress
}

func (w *WrapperTemplateBuilder) SetApplicationPath(applicationPath string) {
	w.applicationPath = applicationPath
}

func (w *WrapperTemplateBuilder) SetApplicationName(applicationName string) {
	w.applicationName = applicationName
}

func (w *WrapperTemplateBuilder) BuildTemplate() (string, error) {
	if w.mintId == "" || w.ownerAddress == "" || w.applicationPath == "" {
		return "", errors.New("mintId or ownerAddress or applicationPath is empty")
	}

	template := fmt.Sprintf(`
	package main

	import (
		_ "embed"
		"fmt"
		"log"
		"os"
		"os/exec"
	)

	//go:embed %v
	var executavel []byte

	const MINT_ID = "%v"
	const OWNER_ADDRESS = "%v"
	const PASSWORD = "123456"
	const EXECUTABLE_NAME = "%v"

	func main() {
		log.Println("Iniciando execução do wrapper!")
		log.Print("Informe a senha para acessar a aplicação: ")

		var pass string

		fmt.Scan(&pass)

		if pass != PASSWORD {
			log.Fatal("Access to content Denied!")
		}

		workDir, err := os.Getwd()

		if err != nil {
			log.Fatal("Error getting working directory:", err)
		}

		workDir += "/"

		tmpFile, err := os.CreateTemp(workDir, EXECUTABLE_NAME)

		if err != nil {
			log.Fatal("Error creating temp file:", err)
		}

		defer os.Remove(tmpFile.Name())

		if _, err = tmpFile.Write(executavel); err != nil {
			log.Fatal("Error writing to temp file:", err)
		}

		if err := tmpFile.Sync(); err != nil {
			log.Fatal("Error syncing temp file:", err)
		}

		if err := tmpFile.Chmod(0755); err != nil {
			log.Fatal("Error chmoding temp file:", err)
		}

		if err := tmpFile.Close(); err != nil {
			log.Fatal("Error closing temp file:", err)
		}

		cmd := exec.Command(tmpFile.Name())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Fatal("Error running temp file:", err)
		}

	}

	`, w.applicationPath, w.mintId, w.ownerAddress, w.applicationName)

	return template, nil
}
