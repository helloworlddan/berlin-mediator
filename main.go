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
	"github.com/helloworlddan/berlin-mediator/culture"
	"golang.org/x/text/language"

	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"
)

const Version = "0.0.1"

func main() {
	projectID := flag.String("g", "", "specify GCP project id.")
	targetLanguage := flag.String("l", "de", "target language.")
	targetRegion := flag.String("r", "DE", "target region.")

	flag.Parse()

	if *projectID == "" {
		fmt.Fprintf(os.Stderr, "no GCP project ID specified, supply one with '-g'\n")
		os.Exit(-1)
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
	fmt.Printf(
		"--------------CONFIG-----------------\nTarget language & region: %s %s \n\n",
		*targetLanguage,
		*targetRegion,
	)

	fmt.Printf("--------------INPUT-----------------\n%s\n\n", input)

	prompt := "[API CALL] translate.DetectLanguage...."
	output, err := detectLanguage(input)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	sourceLanguage := output
	dumpChainLink(prompt, output, sourceLanguage)

	delta, err := culture.Delta(sourceLanguage, *targetRegion)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	fmt.Printf("--------------DELTA-----------------\n%#v\n\n", delta)

	intensities := culture.StyleToTextIntensity(delta)

	if (delta != culture.Style{}) {
		prompt = fmt.Sprintf("Rephrase this text but make it %s", strings.Join(intensities, ", "))
		output, err = predictText(prompt, input, *projectID, 256)
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
		dumpChainLink(prompt, output)
		input = output
	}

	prompt = "[APICALL] translate.Translate...."
	output, err = translateLanguage(input, *targetLanguage, *targetLanguage)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	dumpChainLink(prompt, output)

}

func dumpChainLink(prompt string, output string, context ...string) {
	fmt.Printf("--------------PROMPT-----------------\n%s\n\n", prompt)
	fmt.Printf("--------------OUTPUT-----------------\n%s\n\n", output)
	if len(context) > 0 {
		fmt.Printf("--------------CONTEXT----------------\n%s\n\n", context)
	}
}

func translateLanguage(input string, lang string, region string) (string, error) {
	ctx := context.Background()
	client, err := translate.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()
	langBase := language.MustParseBase(lang)
	langRegion := language.MustParseRegion(region)

	langTag, err := language.Compose(langBase, langRegion)
	if err != nil {
		return "", err
	}

	translationList, err := client.Translate(ctx, []string{input}, langTag, nil)

	translation := translationList[0]

	return translation.Text, nil
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
