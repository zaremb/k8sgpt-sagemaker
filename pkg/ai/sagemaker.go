/*
Copyright 2023 The K8sGPT Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ai

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"encoding/json"
	"log"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/cache"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sagemakerruntime"
)

type SageMakerAIClient struct {
	client   string
	language string
	model    string
}

type Generation struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (c *SageMakerAIClient) Configure(config IAIConfig, language string) error {
	token := config.GetPassword()
	c.language = language
	c.client = fmt.Sprintf("I am a SageMaker client with the token %s ", token)
	c.model = config.GetModel()
	return nil
}

func (c *SageMakerAIClient) GetCompletion(ctx context.Context, prompt string, promptTmpl string) (string, error) {
	// Create a completion request

	if len(promptTmpl) == 0 {
		promptTmpl = PromptMap["default"]
	}

	// Create a new AWS session
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String("eu-west-1")},
		SharedConfigState: session.SharedConfigEnable,
	}))
	// Create a new SageMaker runtime client
	svc := sagemakerruntime.New(sess)

	data := map[string]interface{}{
		"inputs": []interface{}{
			[]interface{}{
				map[string]interface{}{
					"role":    "system",
					"content": "DEFAULT_PROMPT",
				},
				map[string]interface{}{
					"role":    "user",
					"content": fmt.Sprintf(promptTmpl, c.language, prompt),
				},
			},
		},
		"parameters": map[string]interface{}{
			"max_new_tokens": 256,
			"top_p":         0.9,
			"temperature":   0.6,
		},
	}
	// Convert data to []byte
	bytesData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error:", err)
		log.Fatal(err)
		return "", err
	}

	// Define the endpoint name
	endpointName := "endpoint-Wg5SzX93VOZF"

	// Create an input object
	input := &sagemakerruntime.InvokeEndpointInput{
		Body:           bytesData,
		EndpointName:   aws.String(endpointName),
		ContentType:    aws.String("application/json"), // Set the content type as per your model's requirements
		Accept:         aws.String("application/json"), // Set the accept type as per your model's requirements
		CustomAttributes: aws.String("accept_eula=true"),
	}

	// Call the InvokeEndpoint function
	result, err := svc.InvokeEndpoint(input)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	// Define a slice of Generations
	var generations []struct {
		Generation Generation `json:"generation"`
	}

	err = json.Unmarshal([]byte(string(result.Body)), &generations)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	// Access the content
	content := generations[0].Generation.Content
	return content, nil
}

func (a *SageMakerAIClient) Parse(ctx context.Context, prompt []string, cache cache.ICache, promptTmpl string) (string, error) {
	// parse the text with the AI backend
	inputKey := strings.Join(prompt, " ")
	// Check for cached data
	sEnc := base64.StdEncoding.EncodeToString([]byte(inputKey))
	cacheKey := util.GetCacheKey(a.GetName(), a.language, sEnc)

	response, err := a.GetCompletion(ctx, inputKey, promptTmpl)
	if err != nil {
		color.Red("error getting completion: %v", err)
		return "", err
	}

	err = cache.Store(cacheKey, base64.StdEncoding.EncodeToString([]byte(response)))

	if err != nil {
		color.Red("error storing value to cache: %v", err)
		return "", nil
	}

	return response, nil
}

func (a *SageMakerAIClient) GetName() string {
	return "amazonsagemaker"
}
