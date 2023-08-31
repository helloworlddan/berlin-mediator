package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"cloud.google.com/go/translate"

	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"
)

const Version = "0.0.1"

func main() {
	projectID := flag.String("g", "", "specify GCP project id.")
	targetLanguage := flag.String("l", "", "target language.")
	targetCulture := flag.String("c", "", "target culture (country TLD).")

	flag.Parse()

	if *projectID == "" {
		fmt.Fprintf(os.Stderr, "no GCP project ID specified, supply one with '-g'\n")
		os.Exit(-1)
	}

	if *targetLanguage == "" {
		*targetLanguage = "English"
	}

	if *targetCulture == "" {
		*targetCulture = "Armenian"
	}

	inputFile := flag.Arg(0)
	var reader io.Reader
	reader = bufio.NewReader(os.Stdin)

	if inputFile != "-" {
		file, err := os.Open(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open file: %v\n", err)
			os.Exit(-1)
		}
		defer file.Close()
		reader = file
	}

	buffer := &strings.Builder{}
	_, err := io.Copy(buffer, reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read input: %v\n", err)
		os.Exit(-1)
	}

	// TODO OK let's chain stuff

	input := buffer.String()
	fmt.Printf("--------------INPUT-----------------\n%s\n", input)

	prompt := "Which language is this in just one word:"
	output, err := detectLanguage(input)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	sourceLanguage := output
	dumpChainLink(prompt, output, sourceLanguage)

	prompt = "From which culture is the speaker in this text, only use a single word:"
	output, err = predictText(prompt, input, *projectID, 1)
	sourceCulture := strings.TrimSpace(output)
	dumpChainLink(prompt, output, sourceCulture)

	prompt = fmt.Sprintf(
		"Translate this to %s and change the tone so it's more accessible to %s people",
		*targetLanguage,
		*targetCulture,
	)
	output, err = predictText(prompt, input, *projectID, 256)
	dumpChainLink(prompt, output)

	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func dumpChainLink(prompt string, output string, context ...string) {
	fmt.Printf("--------------PROMPT-----------------\n%s\n", prompt)
	fmt.Printf("--------------OUTPUT-----------------\n%s\n", output)
	fmt.Printf("--------------CONTEXT----------------\n%s\n", context)
}

func detectLanguage(input string) (string, error) {
	ctx := context.Background()
	client, err := translate.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	detectionList, err := client.DetectLanguage(ctx, []string{input})
	if err != nil {
		return "", err
	}

	fmt.Printf("%#v\n", detectionList)

	detection := detectionList[0]

	results := []string{}
	for _, lang := range detection {
		results = append(results, lang.Language.String())
	}

	return strings.Join(results, ","), nil
}

func predictText(prompt string, input string, project string, maxOut int) (string, error) {
	ctx := context.Background()
	client, err := aiplatform.NewPredictionClient(
		ctx,
		option.WithEndpoint("us-central1-aiplatform.googleapis.com:443"),
	)
	if err != nil {
		return "", err
	}
	defer client.Close()

	parameters, err := structpb.NewValue(map[string]interface{}{
		"temperature":     0.2,
		"maxOutputTokens": maxOut,
		"topK":            40,
		"topP":            0.95,
	})
	if err != nil {
		return "", err
	}

	instances, err := structpb.NewValue(map[string]interface{}{
		"prompt": fmt.Sprintf("%s: '%s'", prompt, input),
	})

	endpoint := fmt.Sprintf(
		"projects/%s/locations/us-central1/publishers/google/models/text-bison",
		project,
	)

	req := &aiplatformpb.PredictRequest{
		Endpoint:   endpoint,
		Instances:  []*structpb.Value{instances},
		Parameters: parameters,
	}
	resp, err := client.Predict(ctx, req)
	if err != nil {
		return "", err
	}

	predictionFields := resp.Predictions[0].GetStructValue().AsMap()
	return predictionFields["content"].(string), nil
}
