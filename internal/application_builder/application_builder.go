package applicationbuilder

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	wrappertemplate "go_tcc_drm_wrapper/internal/wrapper_template"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const METADATA_SIZE = 40 // Exemplo: 32 bytes (hash) + 8 bytes (uint64)

type Metadata struct {
	Hash             [32]byte
	SizeOfPureBinary uint64
}

type ApplicationBuilder struct {
	zipPath          string
	execName         string
	outputPath       string
	wrapperBuilder   *wrappertemplate.WrapperTemplateBuilder
	OutputFolderPath string
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

	println("Opening reader")

	zr, err := zip.OpenReader(a.zipPath)

	println("Reader opened")

	if err != nil {
		return errors.New("error opening zip file: " + err.Error())
	}
	defer zr.Close()

	absolutePath, err := filepath.Abs(a.outputPath)
	if err != nil {
		return errors.New("error getting absolute path: " + err.Error())
	}

	outputDirPath := absolutePath + "\\" + a.execName + "_folder"

	println("Removing directory")

	err = os.RemoveAll(outputDirPath)

	if err != nil {
		return errors.New("error removing directory: " + err.Error())
	}

	println("Creating directory")

	println("Output directory path: ", outputDirPath)

	err = os.Mkdir(outputDirPath, 0755)

	if err != nil {
		return errors.New("error creating directory: " + err.Error())
	}

	println("Unzipping files")

	for _, f := range zr.File {
		path := filepath.Join(outputDirPath, f.Name)

		if f.Name == a.execName {
			if strings.HasSuffix(f.Name, ".exe") {
				path = filepath.Join(outputDirPath, strings.TrimSuffix(f.Name, ".exe")+"_unwrapped.exe")
			} else {
				path = filepath.Join(outputDirPath, f.Name+"_unwrapped")
			}
		}

		println("Unzipping file: ", f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.Create(path)
		if err != nil {
			return errors.New("error creating file: " + err.Error())
		}

		inFile, err := f.Open()
		if err != nil {
			outFile.Close()
			return errors.New("error opening file from zip: " + err.Error())
		}

		println("File opened, copying...")
		_, err = io.Copy(outFile, inFile)

		inFile.Close()
		outFile.Close()

		if err != nil {
			return errors.New("error copying file: " + err.Error())
		}
	}

	println("Files unzipped")

	return nil
}

func (a *ApplicationBuilder) inputHash(binaryPath string) error {
	binaryFile, errOpeningFile := os.Open(binaryPath)
	if errOpeningFile != nil {
		return errors.New("error opening file to input hash validation: " + errOpeningFile.Error())
	}

	hasher := sha256.New()
	_, err := io.Copy(hasher, binaryFile)
	if err != nil {
		binaryFile.Close()
		return errors.New("error calculating hash: " + err.Error())
	}
	referenceHash := hasher.Sum(nil)

	fileInfo, errStatFile := binaryFile.Stat()
	if errStatFile != nil {
		binaryFile.Close()
		return errors.New("error getting file status: " + errStatFile.Error())
	}

	sizeOfBinary := fileInfo.Size()
	binaryFile.Close()

	var metadata Metadata
	copy(metadata.Hash[:], referenceHash)
	metadata.SizeOfPureBinary = uint64(sizeOfBinary)

	var metadataBytes bytes.Buffer
	err = binary.Write(&metadataBytes, binary.LittleEndian, metadata)
	if err != nil {
		return errors.New("error writing metadata to buffer: " + err.Error())
	}

	finalFile, err := os.OpenFile(binaryPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return errors.New("error opening file for appending metadata: " + err.Error())
	}
	defer finalFile.Close()

	_, err = finalFile.Write(metadataBytes.Bytes())
	if err != nil {
		return errors.New("error writing metadata to file: " + err.Error())
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

	var unwrappedApplicationName string

	if strings.HasSuffix(a.execName, ".exe") {
		unwrappedApplicationName = strings.TrimSuffix(a.execName, ".exe") + "_unwrapped.exe"
	} else {
		unwrappedApplicationName = a.execName + "_unwrapped"
	}

	// Pegar o caminho absoluto pra criar o arquivo
	absolutePath, err := filepath.Abs(a.outputPath)
	if err != nil {
		return errors.New("error getting absolute path: " + err.Error())
	}

	folderPath := absolutePath + "\\" + a.execName + "_folder"
	a.OutputFolderPath = folderPath

	println("folderPath: ", folderPath)

	a.wrapperBuilder.SetApplicationPath(unwrappedApplicationName)
	a.wrapperBuilder.SetApplicationName(a.execName)

	wrapper, err := a.wrapperBuilder.BuildTemplate()

	if err != nil {
		return err
	}

	fmt.Println("Wrapper built!")

	fmt.Println("Saving wrapper...")

	wrapperFile, err := os.Create(filepath.Join(folderPath, "wrapper.go"))
	if err != nil {
		return errors.New("error creating wrappers file: " + err.Error())
	}

	_, err = wrapperFile.WriteString(wrapper)
	if err != nil {
		wrapperFile.Close()
		return errors.New("error writing wrappers file: " + err.Error())
	}
	wrapperFile.Close()

	fmt.Println("Wrapper saved!")

	cmd := exec.Command("go", "build", "-o", filepath.Join(folderPath, a.execName), filepath.Join(folderPath, "wrapper.go"))

	if err := cmd.Run(); err != nil {
		return errors.New("error building wrapper: " + err.Error())
	}

	fmt.Println("Application Wrapped!")

	binaryPath := filepath.Join(folderPath, a.execName)
	errInputingHash := a.inputHash(binaryPath)
	if errInputingHash != nil {
		return errInputingHash
	}

	os.Remove(filepath.Join(folderPath, unwrappedApplicationName))
	os.Remove(filepath.Join(folderPath, "wrapper.go"))

	// Pegar o nome do arquivo sem o ".exe"

	var execNameWithoutExtension string
	if strings.HasSuffix(a.execName, ".exe") {
		execNameWithoutExtension = strings.TrimSuffix(a.execName, ".exe")
	} else {
		execNameWithoutExtension = a.execName
	}

	// Preciso compactar tudo de novo
	zipOutPath := filepath.Join(a.outputPath, execNameWithoutExtension+".zip")
	zipFile, err := os.Create(zipOutPath)
	if err != nil {
		return err
	}

	// 2. Cria o writer para o zip
	zipWriter := zip.NewWriter(zipFile)

	err = filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Evita incluir a própria pasta raiz no nome interno
		relPath, err := filepath.Rel(folderPath, path)
		if err != nil {
			return err
		}

		// Ignora o diretório raiz
		if relPath == "." {
			return nil
		}

		// Se for diretório, apenas garante a estrutura (sem conteúdo)
		if info.IsDir() {
			_, err = zipWriter.Create(relPath + "/")
			return err
		}

		// 4. Abre arquivo para adicionar ao zip
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// Cria entrada no zip com o caminho relativo
		entry, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		// Copia os bytes do arquivo para o zip
		_, err = io.Copy(entry, file)
		return err
	})

	if err != nil {
		zipWriter.Close()
		zipFile.Close()
		return err
	}

	err = zipWriter.Close()
	if err != nil {
		zipFile.Close()
		return err
	}

	return zipFile.Close()
}
