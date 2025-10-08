package applicationbuilder

import (
	"archive/zip"
	"errors"
	"fmt"
	wrappertemplate "go_tcc_drm_wrapper/internal/wrapper_template"
	"io"
	"os"
	"os/exec"
)

type ApplicationBuilder struct {
	zipPath        string
	execName       string
	outputPath     string
	wrapperBuilder *wrappertemplate.WrapperTemplateBuilder
}

func New() *ApplicationBuilder {
	return &ApplicationBuilder{}
}

func (a *ApplicationBuilder) SetZipPath(zipPath string) {
	a.zipPath = zipPath
}

func (a *ApplicationBuilder) SetExecName(execName string) {
	a.execName = execName
}

func (a *ApplicationBuilder) SetOutputPath(outputPath string) {
	a.outputPath = outputPath
}

func (a *ApplicationBuilder) SetWrapperBuilder(wrapperBuilder *wrappertemplate.WrapperTemplateBuilder) {
	a.wrapperBuilder = wrapperBuilder
}

func (a *ApplicationBuilder) unzip() error {
	if a.zipPath == "" || a.execName == "" || a.outputPath == "" {
		return errors.New("zipPath, execName or outputPath is empty")
	}

	zr, err := zip.OpenReader(a.zipPath)

	if err != nil {
		return errors.New("error opening zip file: " + err.Error())
	}
	defer zr.Close()

	outputDirPath := a.outputPath + "/" + a.execName + "_folder"

	err = os.Mkdir(outputDirPath, 0755)

	if err != nil {
		return errors.New("error creating directory: " + err.Error())
	}

	for _, f := range zr.File {
		targetPath := outputDirPath + "/" + f.Name

		outFile, err := os.Create(targetPath)

		if err != nil {
			return errors.New("error creating file: " + err.Error())
		}
		defer outFile.Close()

		inFile, err := f.Open()

		if err != nil {
			return errors.New("error opening file: " + err.Error())
		}
		defer inFile.Close()

		_, err = io.Copy(outFile, inFile)

		if err != nil {
			return errors.New("error copying file: " + err.Error())
		}

	}

	return nil
}

func (a *ApplicationBuilder) BuildApplication() error {

	if a.zipPath == "" || a.execName == "" || a.outputPath == "" || a.wrapperBuilder == nil {
		return errors.New("zipPath, execName, outputPath or wrapperTemplate is empty")
	}

	// Aqui eu preciso fazer algumas coisas
	// 1 - Acessar os arquivos / pastas dentro do zip
	// 2 - Pegar o executável
	// 3 - Gerar o Executável com o Wrapper
	// 4 - Criar uma pasta no caminho informado
	// 5 - Salvar o novo executável com o Wrapper dentro dessa pasta
	// 6 - Copiar os arquivos / pastas restantes dentro do zip para a pasta criada

	fmt.Println("Unzipping...")

	err := a.unzip()

	if err != nil {
		return err
	}

	fmt.Println("Unzipped!")

	fmt.Println("Building wrapper...")

	folderPath := a.outputPath + "/" + a.execName + "_folder"
	a.wrapperBuilder.SetApplicationPath(a.execName)
	a.wrapperBuilder.SetApplicationName(a.execName)

	wrapper, err := a.wrapperBuilder.BuildTemplate()

	if err != nil {
		return err
	}

	fmt.Println("Wrapper built!")

	fmt.Println("Saving wrapper...")

	wrapperFile, err := os.Create(folderPath + "/" + "wrapper.go")

	if err != nil {
		return errors.New("error creating wrappers file: " + err.Error())
	}

	defer wrapperFile.Close()

	_, err = wrapperFile.WriteString(wrapper)

	if err != nil {
		return errors.New("error writing wrappers file: " + err.Error())
	}

	fmt.Println("Wrapper saved!")

	cmd := exec.Command("go", "build", "-o", folderPath+"/"+a.execName+"_wrapped", folderPath+"/wrapper.go")

	if err := cmd.Run(); err != nil {
		return errors.New("error building wrapper: " + err.Error())
	}

	fmt.Println("Application Wrapped!")

	// Agora é apagar o arquivo original

	os.Remove(a.outputPath + "/" + a.execName + "_folder/" + a.execName)

	return nil
}
