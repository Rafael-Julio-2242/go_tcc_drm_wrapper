package wrappertemplate

import (
	"errors"
	"fmt"
)

type WrapperTemplateBuilder struct {
	contractId      string
	chain           string
	applicationPath string
	applicationName string
}

func NewWrapperTemplateBuilder() *WrapperTemplateBuilder {
	return &WrapperTemplateBuilder{}
}

func (w *WrapperTemplateBuilder) SetContractId(contractId string) {
	w.contractId = contractId
}

func (w *WrapperTemplateBuilder) SetChain(chain string) {
	w.chain = chain
}

func (w *WrapperTemplateBuilder) SetApplicationPath(applicationPath string) {
	w.applicationPath = applicationPath
}

func (w *WrapperTemplateBuilder) SetApplicationName(applicationName string) {
	w.applicationName = applicationName
}

func (w *WrapperTemplateBuilder) BuildTemplate() (string, error) {
	if w.contractId == "" || w.applicationPath == "" {
		return "", errors.New("contractId or applicationPath is empty")
	}

	functions := ""
	functions += "const METHOD_NAME = \"verify_access\"\n"
	functions += "const URL = \"http://localhost:55321/rpc\"\n"

	functions += "type RpcReqBody struct {\n"
	functions += "	Jsonrpc string   `json:\"jsonrpc\"`\n"
	functions += "	Method  string   `json:\"method\"`\n"
	functions += "	Params  []string `json:\"params\"`\n"
	functions += "	Id      string   `json:\"id\"`\n"
	functions += "}\n"

	functions += "type RpcError struct {\n"
	functions += "	Code    int      `json:\"code\"`\n"
	functions += "	Message string `json:\"message\"`\n"
	functions += "	Data    []string `json:\"data,omitempty\"`\n"
	functions += "}\n"

	functions += "type RpcResponseResult struct {\n"
	functions += "	Status     string `json:\"status\"`\n"
	functions += "	WalletUsed string `json:\"walletUsed\"`\n"
	functions += "	HasItem    bool   `json:\"hasItem\"`\n"
	functions += "}\n"

	functions += "type RpcResponse struct {\n"
	functions += "	Jsonrpc string    `json:\"jsonrpc\"`\n"
	functions += "	Result  RpcResponseResult `json:\"result,omitempty\"`\n"
	functions += "	Error   *RpcError `json:\"error,omitempty\"`\n"
	functions += "	Id      string    `json:\"id\"`\n"
	functions += "}\n"

	functions += "func generateId() string {\n"
	functions += "	possibleCaracters := []rune(\"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789/*-+=[]{};:|?!@#$%&*()\")\n\n"

	functions += "	var id string\n\n"

	functions += "	for len(id) < 12 {\n"
	functions += "		var charPos = rand.IntN(len(possibleCaracters) - 1)\n"
	functions += "		id += fmt.Sprintf(\"%c\", possibleCaracters[charPos])\n"
	functions += "\t}\n"

	functions += "\ttimestamp := fmt.Sprintf(\"%v\", time.Now().Unix())\n"

	functions += "\tid += \"_\" + timestamp\n"

	functions += "\treturn id\n"
	functions += "}\n"

	functions += "func generateBody() (*bytes.Reader, error) {\n"
	functions += "	requestId := generateId()\n\n"

	functions += "	reqBody := RpcReqBody{\n"
	functions += "		Jsonrpc: \"2.0\",\n"
	functions += "		Method:  METHOD_NAME,\n"
	functions += "		Params:  []string{CONTRACT_ID, CHAIN},\n"
	functions += "		Id:      requestId,\n"
	functions += "	}\n"

	functions += "	bodyBytes, err := json.Marshal(reqBody)\n"

	functions += "	if err != nil {\n"
	functions += "		log.Fatalf(\"ERROR ON GENERATING REQUEST BODY: %v\", err.Error())\n"
	functions += "	}\n"

	functions += "	return bytes.NewReader(bodyBytes), nil\n"
	functions += "}\n"

	functions += "func validator() bool {\n"

	functions += "	reqBody, err := generateBody()\n"

	functions += "	if err != nil {\n"
	functions += "		fmt.Println(\"ERROR ON VALIDATION: \", err.Error())\n"
	functions += "		return false\n"
	functions += "	}\n"

	functions += "	res, err := http.Post(URL, \"application/json\", reqBody)\n"

	functions += "	if err != nil {\n"
	functions += "		log.Fatalf(\"ERROR ON THE REQUEST: %v\", err.Error())\n"
	functions += "	}\n"

	functions += "	defer res.Body.Close()\n"

	functions += "	var rpcResp RpcResponse\n"
	functions += "	if err := json.NewDecoder(res.Body).Decode(&rpcResp); err != nil {\n"
	functions += "		log.Fatalf(\"ERROR ON DECODING RESPONSE: %v\", err.Error())\n"
	functions += "	}\n"

	functions += "	if rpcResp.Error != nil {\n"
	functions += "		fmt.Printf(\"Error on the incoming response from the CLIENT: code=%v, message=%v, data%v\", rpcResp.Error.Code, rpcResp.Error.Message, rpcResp.Error.Data)\n"
	functions += "		return false\n"
	functions += "	}\n"

	functions += "	var isValidAccess bool\n"

	functions += "	if rpcResp.Result.HasItem {\n"
	functions += "		isValidAccess = true\n"
	functions += "	} else {\n"
	functions += "		isValidAccess = false\n"
	functions += "	}\n"

	functions += "	return isValidAccess\n"
	functions += "}\n"

	functions += "func verifyIntegrity() bool {\n"

	functions += "selfPath, _ := os.Executable()\n"
	functions += "file, err := os.Open(selfPath)\n"
	functions += "if err != nil {\n"
	functions += "return false"
	functions += "}\n\n"
	functions += "defer file.Close()\n"
	functions += "fileInfo, _ := file.Stat()\n"
	functions += "totalSize := fileInfo.Size()\n"
	functions += "metadataOffset := totalSize - int64(METADATA_SIZE)\n"
	functions += "metadataBytes := make([]byte, METADATA_SIZE)\n"
	functions += "file.ReadAt(metadataBytes, metadataOffset)\n"
	functions += "var metadata Metadata\n"
	functions += "buf := bytes.NewReader(metadataBytes)\n"
	functions += "binary.Read(buf, binary.LittleEndian, &metadata)\n"
	functions += "hasher := sha256.New()\n"
	functions += "limitReader := io.LimitReader(file, int64(metadata.SizeOfPureBinary))\n"
	functions += "io.Copy(hasher, limitReader)\n"
	functions += "calculatedHash := hasher.Sum(nil)\n"
	functions += "expectedHash := metadata.Hash[:]\n"
	functions += "if bytes.Equal(calculatedHash, expectedHash) {\n"
	functions += "return true\n"
	functions += "} else {\n"
	functions += "return false\n"
	functions += "}\n"
	functions += "}\n"
	functions += "func main() {\n"
	functions += "\tlog.Println(\"Iniciando execução do wrapper!\")\n"
	functions += "hasIntegrity := verifyIntegrity()\n"
	functions += "if !hasIntegrity {\n"
	functions += "log.Fatal(\"APPLICATION INTEGRITY COMPROMISED!\")\n"
	functions += "}\n"
	functions += "\tcanAccess := validator()\n"

	functions += "\tif !canAccess {\n"
	functions += "\t	log.Fatal(\"Unauthorized access\")\n"
	functions += "\t}\n"

	functions += "\tworkDir, err := os.Getwd()\n"

	functions += "\tif err != nil {\n"
	functions += "\t	log.Fatal(\"Error getting working directory:\", err)\n"
	functions += "\t}\n"

	functions += "\tworkDir += \"/\"\n"

	functions += "\ttmpFile, err := os.CreateTemp(workDir, EXECUTABLE_NAME)\n"

	functions += "\tif err != nil {\n"
	functions += "\t	log.Fatal(\"Error creating temp file:\", err)\n"
	functions += "\t}\n"

	functions += "\tdefer os.Remove(tmpFile.Name())\n"

	functions += "\tif _, err = tmpFile.Write(executavel); err != nil {\n"
	functions += "\t	log.Fatal(\"Error writing to temp file:\", err)\n"
	functions += "\t}\n"

	functions += "\tif err := tmpFile.Sync(); err != nil {\n"
	functions += "\t	log.Fatal(\"Error syncing temp file:\", err)\n"
	functions += "\t}\n"

	functions += "\tif err := tmpFile.Chmod(0755); err != nil {\n"
	functions += "\t	log.Fatal(\"Error chmoding temp file:\", err)\n"
	functions += "\t}\n"

	functions += "\tif err := tmpFile.Close(); err != nil {\n"
	functions += "\t	log.Fatal(\"Error closing temp file:\", err)\n"
	functions += "\t}\n"

	functions += "\tcmd := exec.Command(tmpFile.Name())\n"
	functions += "\tcmd.Stdout = os.Stdout\n"
	functions += "\tcmd.Stderr = os.Stderr\n"

	functions += "\tif err := cmd.Run(); err != nil {\n"
	functions += "\t	log.Fatal(\"Error running temp file:\", err)\n"
	functions += "\t}\n"

	functions += "}\n"

	template := fmt.Sprintf(`
	package main

	import (
		"bytes"
		"crypto/sha256"
		_ "embed"
		"encoding/binary"
		"encoding/json"
		"fmt"
		"io"
		"log"
		"math/rand/v2"
		"net/http"
		"os"
		"os/exec"
		"time"
	)

	//go:embed %v
	var executavel []byte

	const CONTRACT_ID = "%v"
	const CHAIN = "%v"
	const EXECUTABLE_NAME = "%v"

	const METADATA_SIZE = 40 // Exemplo: 32 bytes (hash) + 8 bytes (uint64)

	type Metadata struct {
		Hash             [32]byte
		SizeOfPureBinary uint64
	}

		%s

	`, w.applicationPath, w.contractId, w.chain, w.applicationName, functions)

	return template, nil
}
