package main

import (
	"fmt"
	applicationbuilder "go_tcc_drm_wrapper/internal/application_builder"
	wrappertemplate "go_tcc_drm_wrapper/internal/wrapper_template"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

const INPUT_DIR = "application_input"
const OUTPUT_DIR = "application_output"

func main() {

	// Eu preciso criar um endpoint que vai conseguir encapsular
	// uma aplicação em um Wrapper e devolver o zip

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.String(200, "Hello World")
	})

	router.POST("/wrap", WrapperHandler)

	router.Run(":8080")
}

func WrapperHandler(c *gin.Context) {

	contractId := c.PostForm("contractId")
	chain := c.PostForm("chain")
	execPath := c.PostForm("execPath")

	if contractId == "" || chain == "" || execPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing parameters",
		})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing file",
		})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Error opening file",
		})
		return
	}

	defer file.Close()
	buf := make([]byte, 4)
	_, err = file.Read(buf)
	if err != nil && err != io.EOF {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error reading file",
		})
		return
	}

	// Volta para o começo do arquivo antes de qualquer leitura/salvamento
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error resetting file reading: %s", err.Error())
		return
	}

	if !isFileZip(buf) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File is not a valid zip",
		})
		return
	}

	// AQUI É ONDE EU PRECISO FAZER A IMPLEMENTAÇÃO
	// Informações que eu tenho aqui:
	// - contractId
	// - chain
	// - execPath
	// - file

	zipInputPath, err := saveTemporaryFile(fileHeader, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error saving file",
		})
		return
	}

	wrapperBuilder := wrappertemplate.NewWrapperTemplateBuilder()
	wrapperBuilder.SetContractId(contractId)
	wrapperBuilder.SetChain(chain)

	appBuilder := applicationbuilder.New()
	appBuilder.SetOutputPath(OUTPUT_DIR)
	appBuilder.SetExecName(execPath)
	appBuilder.SetZipPath(zipInputPath)
	appBuilder.SetWrapperBuilder(wrapperBuilder)

	err = appBuilder.BuildApplication()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error building application",
		})
		return
	}

	absOutputDir, err := filepath.Abs(OUTPUT_DIR)
	if err != nil {
		panic(err)
	}

	var execNameWithoutExtension string
	if strings.HasSuffix(execPath, ".exe") {
		execNameWithoutExtension = strings.TrimSuffix(execPath, ".exe")
	} else {
		execNameWithoutExtension = execPath
	}

	zipFilePath := absOutputDir + "\\" + execNameWithoutExtension + ".zip"
	println("ZIP FILE PATH: ", zipFilePath)

	if _, err := os.Stat(zipFilePath); os.IsNotExist(err) {
		c.String(http.StatusNotFound, "File not found")
		return
	}

	zipFile, err := os.Open(zipFilePath)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error opening file: %s", err.Error())
		return
	}
	defer zipFile.Close()

	zipFileInfo, err := zipFile.Stat()
	if err != nil {
		c.String(http.StatusInternalServerError, "Error getting file info: %s", err.Error())
		return
	}

	// Vou tentar retornar os bytes do arquivo.
	buffer := make([]byte, zipFileInfo.Size())
	_, err = zipFile.Read(buffer)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error reading file: %s", err.Error())
		return
	}

	println("Setting Headers...")

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename="+execNameWithoutExtension+".zip")
	c.Header("Content-Size", fmt.Sprintf("%d", len(buffer)))

	fmt.Printf("Sending %d bytes\n", len(buffer))

	c.Data(http.StatusOK, "application/zip", buffer)

	// err = os.RemoveAll(OUTPUT_DIR)
	// if err != nil {
	// 	println("Error removing directory: ", err.Error())
	// }

}

func isFileZip(data []byte) bool {
	return len(data) >= 4 &&
		data[0] == 'P' &&
		data[1] == 'K' &&
		(data[2] == 3 || data[2] == 5 || data[2] == 7) &&
		(data[3] == 4 || data[3] == 6 || data[3] == 8)
}

func saveTemporaryFile(file *multipart.FileHeader, c *gin.Context) (string, error) {

	filename := filepath.Base(file.Filename)

	dir := INPUT_DIR

	err := os.MkdirAll(dir, os.ModePerm)

	if err != nil {
		return "", err
	}

	savePath := filepath.Join(dir, filename)

	// Salva o arquivo
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.String(http.StatusInternalServerError, "Falha ao salvar arquivo: %s", err.Error())
		return "", err
	}

	// Obtém o caminho absoluto do arquivo no sistema
	absPath, err := filepath.Abs(savePath)

	if err != nil {
		c.String(http.StatusInternalServerError, "Erro ao obter caminho absoluto: %s", err.Error())
		return "", err
	}

	return absPath, nil
}
