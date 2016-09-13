package lookup

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/utils"
	"github.com/rancher/rancher-catalog-service/model"
	"github.com/rancher/rancher-compose/rancher"
	"gopkg.in/yaml.v2"
)

type questionWrapper struct {
	Questions []model.Question `yaml:"questions,omitempty"`
}

type QuestionLookup struct {
	parent    config.EnvironmentLookup
	questions map[string]model.Question
	variables map[string]string
	Context   *rancher.Context
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

	data := map[string]map[string]interface{}{}
	if err := yaml.Unmarshal(contents, &data); err != nil {
		return err
	}

	rawQuestions := data[".catalog"]
	if rawQuestions != nil {
		var wrapper questionWrapper
		if err := utils.Convert(rawQuestions, &wrapper); err != nil {
			return err
		}

		for _, question := range wrapper.Questions {
			q.questions[question.Variable] = question
		}
	}

	return nil
}

func join(key, v string) []string {
	return []string{fmt.Sprintf("%s=%s", key, v)}
}

func (f *QuestionLookup) Lookup(key, serviceName string, config *config.ServiceConfig) []string {
	if v, ok := f.variables[key]; ok {
		return join(key, v)
	}

	if f.Context != nil {
		stack, err := f.Context.LoadStack()
		if err == nil && stack != nil {
			if v, ok := stack.Environment[key]; ok {
				return join(key, fmt.Sprintf("%v", v))
			}
		}
	}

	if f.parent != nil {
		parentResult := f.parent.Lookup(key, serviceName, config)
		if len(parentResult) > 0 {
			return parentResult
		}
	}

	question, ok := f.questions[key]
	if !ok {
		return nil
	}

	answer := ask(question)
	if answer != "" {
		f.variables[key] = answer
		return join(key, answer)
	}

	return nil
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
