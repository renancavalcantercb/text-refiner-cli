package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	OpenAIModel        string `json:"openai_model"`
	OpenAIAPIEndpoint  string `json:"openai_api_endpoint"`
	RequestTimeoutSecs int    `json:"request_timeout_seconds"`
	DefaultLanguage    string `json:"default_language"`
	OpenAIAPIKey       string
}

type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

var (
	httpClient         = &http.Client{}
	supportedLanguages = map[string]string{
		"en":    "en-us",
		"pt":    "pt-br",
		"en-us": "en-us",
		"pt-br": "pt-br",
	}
)

func main() {
	config, err := loadAppConfig()
	if err != nil {
		log.Fatal(err)
	}

	text, lang, copyToClipboard := parseFlags()

	if lang == "" {
		lang = config.DefaultLanguage
	}

	validatedLang, err := validateLanguage(lang)
	if err != nil {
		log.Fatal(err)
	}

	httpClient.Timeout = time.Duration(config.RequestTimeoutSecs) * time.Second

	inputText := getInputText(text)

	response, err := callOpenAI(config.OpenAIAPIKey, OpenAIRequest{
		Model: config.OpenAIModel,
		Messages: []Message{
			{
				Role:    "user",
				Content: fmt.Sprintf("Improve this text in %s and only return the improved text: %s", validatedLang, inputText),
			},
		},
	}, config.OpenAIAPIEndpoint)
	if err != nil {
		log.Fatal(err)
	}

	improvedText := response.Choices[0].Message.Content
	fmt.Println("Improved text:")
	fmt.Println(improvedText)

	if copyToClipboard {
		if err := clipboard.WriteAll(improvedText); err != nil {
			log.Fatalf("Error copying to clipboard: %v", err)
		}
		fmt.Println("Improved text copied to clipboard!")
	}
}

func loadAppConfig() (*AppConfig, error) {
	file, err := os.Open("config.json")
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %w", err)
	}
	defer file.Close()

	var config AppConfig
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("error decoding config file: %w", err)
	}

	if err := godotenv.Load(".env"); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}
	config.OpenAIAPIKey = os.Getenv("OPENAI_API_KEY")
	if config.OpenAIAPIKey == "" {
		return nil, errors.New("OPENAI_API_KEY not set in .env file")
	}

	return &config, nil
}

func parseFlags() (text, lang string, copy bool) {
	flag.StringVar(&text, "text", "", "Text to be improved")
	flag.StringVar(&lang, "lang", "", "Language for improvement (en-us or pt-br)")
	flag.StringVar(&lang, "l", "", "Alias for -lang")
	flag.BoolVar(&copy, "copy", false, "Copy improved text to clipboard")
	flag.BoolVar(&copy, "c", false, "Alias for -copy")

	flag.Parse()
	return
}

func getInputText(flagText string) string {
	if flagText != "" {
		return flagText
	}

	fmt.Println("Enter text to be improved (press Enter twice or Ctrl+D to finish):")
	scanner := bufio.NewScanner(os.Stdin)
	var sb strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		sb.WriteString(line + "\n")
	}
	return strings.TrimSpace(sb.String())
}

func validateLanguage(lang string) (string, error) {
	if validated, ok := supportedLanguages[strings.ToLower(lang)]; ok {
		return validated, nil
	}
	return "", fmt.Errorf("invalid language option: %s", lang)
}

func callOpenAI(apiKey string, request OpenAIRequest, endpoint string) (OpenAIResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return OpenAIResponse{}, fmt.Errorf("error serializing request: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return OpenAIResponse{}, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return OpenAIResponse{}, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return OpenAIResponse{}, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return OpenAIResponse{}, fmt.Errorf("HTTP status not OK (%s): %s", resp.Status, string(body))
	}

	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return OpenAIResponse{}, fmt.Errorf("error deserializing response: %w", err)
	}

	if openAIResp.Error.Message != "" {
		return OpenAIResponse{}, fmt.Errorf("OpenAI API error: %s", openAIResp.Error.Message)
	}

	return openAIResp, nil
}
