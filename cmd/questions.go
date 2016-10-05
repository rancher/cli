package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rancher/go-rancher/catalog"
)

func askQuestions(answers map[string]interface{}, templateVersion catalog.TemplateVersion) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	for _, q := range templateVersion.Questions {
		question := catalog.Question{}
		err := mapstructure.Decode(q, &question)
		if err != nil {
			return nil, err
		}

		if answer, ok := answers[question.Variable]; ok {
			result[question.Variable] = answer
		} else {
			result[question.Variable] = askQuestion(question)
		}
	}
	return result, nil
}

func askQuestion(question catalog.Question) interface{} {
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
