package lookup

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/docker/libcompose/utils"
	"github.com/rancher/rancher-catalog-service/model"
	"github.com/rancher/rancher-compose-executor/config"
	rUtils "github.com/rancher/rancher-compose-executor/utils"
	"gopkg.in/yaml.v2"
)

type questionWrapper struct {
	Questions []model.Question `yaml:"questions,omitempty"`
}

type QuestionLookup struct {
	parent    config.EnvironmentLookup
	questions map[string]model.Question
	variables map[string]string
}

func NewQuestionLookup(file string, parent config.EnvironmentLookup) (*QuestionLookup, error) {
	ret := &QuestionLookup{
		parent:    parent,
		variables: map[string]string{},
		questions: map[string]model.Question{},
	}

	if err := ret.parse(file); err != nil {
		return nil, err
	}

	return ret, nil
}

func (q *QuestionLookup) parse(file string) error {
	contents, err := ioutil.ReadFile(file)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	q.questions, err = ParseQuestions(contents)
	if err != nil {
		return err
	}

	return nil
}

func ParseQuestions(contents []byte) (map[string]model.Question, error) {
	catalogConfig, err := ParseCatalogConfig(contents)
	if err != nil {
		return nil, err
	}

	questions := map[string]model.Question{}
	for _, question := range catalogConfig.Questions {
		questions[question.Variable] = question
	}

	return questions, nil
}

func ParseCatalogConfig(contents []byte) (*model.RancherCompose, error) {
	rawConfig, err := config.CreateRawConfig(contents)
	if err != nil {
		return nil, err
	}
	var rawCatalogConfig interface{}

	if rawConfig.Version == "2" && rawConfig.Services[".catalog"] != nil {
		rawCatalogConfig = rawConfig.Services[".catalog"]
	}

	var data map[string]interface{}
	if err := yaml.Unmarshal(contents, &data); err != nil {
		return nil, err
	}

	if data["catalog"] != nil {
		rawCatalogConfig = data["catalog"]
	} else if data[".catalog"] != nil {
		rawCatalogConfig = data[".catalog"]
	}

	if rawCatalogConfig != nil {
		var catalogConfig model.RancherCompose
		if err := utils.Convert(rawCatalogConfig, &catalogConfig); err != nil {
			return nil, err
		}

		return &catalogConfig, nil
	}

	return &model.RancherCompose{}, nil
}

func (f *QuestionLookup) Lookup(key string, config *config.ServiceConfig) []string {
	if v, ok := f.variables[key]; ok {
		return []string{fmt.Sprintf("%s=%s", key, v)}
	}

	if f.parent == nil {
		return nil
	}

	return f.parent.Lookup(key, config)
}

func ask(question model.Question) string {
	if len(question.Description) > 0 {
		fmt.Println(question.Description)
	}
	fmt.Printf("%s %s[%s]: ", question.Label, question.Variable, question.Default)

	answer, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return ""
	}

	answer = strings.TrimSpace(answer)
	if answer == "" {
		answer = question.Default
	}

	return answer
}

func (f *QuestionLookup) Variables() map[string]string {
	// TODO: precedence
	return rUtils.MapUnion(f.variables, f.parent.Variables())
}
