package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/valyala/fastjson"
)

type ProjectConfig struct {
	Action          string `json:"action"`
	Branch          string `json:"branch"`
	BuildCommand    string `json:"buildCommand"`
	Cwd             string `json:"cwd"`
	DeployCommand   string `json:"deployCommand"`
	PreBuildCommand string `json:"preBuildCommand"`
	Trigger         string `json:"trigger"`
	SshKeyPath      string `json:"sshKeyPath"`
}

func throwError(err error) {
	if err != nil {
		panic(err)
	}
}

func getRootPath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	return exPath
}

func parseJSONString(value string) *fastjson.Value {
	var p fastjson.Parser
	jsonContent, err := p.Parse(value)

	throwError(err)

	return jsonContent
}

func parseJSONBody(body io.Reader) *fastjson.Value {
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	bodyStr := buf.String()

	return parseJSONString(bodyStr)
}

func bytesToString(value []byte) string {
	return string(value[:])
}

func buildApp(currentProject ProjectConfig) {
	cmd := exec.Command("/bin/sh", getRootPath()+"/fetchAndBuildAPP.sh", currentProject.Cwd, currentProject.Branch, currentProject.PreBuildCommand, currentProject.BuildCommand, currentProject.DeployCommand, currentProject.SshKeyPath)

	var stdbuffer bytes.Buffer
	writer := io.MultiWriter(os.Stdout, &stdbuffer)

	cmd.Stdout = writer
	cmd.Stderr = writer

	cmdErr := cmd.Run()
	throwError(cmdErr)

	log.Println(stdbuffer.String())
}

func fetchAndBuildAPP(targetBranch string, state string, eventType string) {
	content, err := ioutil.ReadFile(getRootPath() + "/config.json")
	throwError(err)

	var projects []ProjectConfig

	jsonErr := json.Unmarshal(content, &projects)
	throwError(jsonErr)

	for _, project := range projects {
		if project.Branch == targetBranch && project.Action == state && project.Trigger == eventType {
			buildApp(project)
		}
	}

}

func triggerBuild(v *fastjson.Value) {
	eventType := bytesToString(v.GetStringBytes("object_kind"))
	targetBranch := bytesToString(v.GetStringBytes("object_attributes", "target_branch"))
	state := bytesToString(v.GetStringBytes("object_attributes", "state"))

	fetchAndBuildAPP(targetBranch, state, eventType)
}

func hook(writer http.ResponseWriter, request *http.Request) {
	v := parseJSONBody(request.Body)
	go triggerBuild(v)

	writer.WriteHeader(200)
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/hook", hook).Methods("POST")
	log.Println("I am started")
	log.Fatal(http.ListenAndServe(":8000", router))
}
