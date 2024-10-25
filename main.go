package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/joho/godotenv"
)

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

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("API key not found in .env")
	}

	textPtr := flag.String("text", "", "Text to improve")
	langPtr := flag.String("lang", "en-us", "Language for the improvement (options: en-us, pt-br, aliases: en, pt)")
	copyPtr := flag.Bool("copy", false, "Copy the improved text to the clipboard")
	flag.Parse()

	langMap := map[string]string{
		"en":    "en-us",
		"pt":    "pt-br",
		"en-us": "en-us",
		"pt-br": "pt-br",
	}

	lang, ok := langMap[*langPtr]
	if !ok {
		log.Fatalf("Invalid language option: %s. Please use 'en-us', 'pt-br', 'en' or 'pt'.", *langPtr)
	}

	var inputText string

	if *textPtr != "" {
		inputText = *textPtr
	} else {
		fmt.Println("Please enter the text to improve (press Enter twice or Ctrl+D to finish):")

		scanner := bufio.NewScanner(os.Stdin)
		var sb strings.Builder

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				break
			}
			sb.WriteString(line + " ")
		}

		if err := scanner.Err(); err != nil {
			log.Fatalf("Error reading input: %v", err)
		}

		inputText = sb.String()
	}

	reqBody := OpenAIRequest{
		Model: "gpt-4o-mini",
		Messages: []Message{
			{
				Role:    "user",
				Content: fmt.Sprintf("Improve this text in %s and only return the improved text: %s", lang, inputText),
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatalf("Error serializing request: %v", err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making the request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading the response: %v", err)
	}

	var openAIResp OpenAIResponse
	err = json.Unmarshal(body, &openAIResp)
	if err != nil {
		log.Fatalf("Error deserializing the response: %v", err)
	}

	if openAIResp.Error.Message != "" {
		log.Fatalf("OpenAI API error: %v", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) > 0 {
		improvedText := openAIResp.Choices[0].Message.Content
		fmt.Println("Improved text:")
		fmt.Println(improvedText)

		if *copyPtr {
			err := clipboard.WriteAll(improvedText)
			if err != nil {
				log.Fatalf("Error copying text to clipboard: %v", err)
			}
			fmt.Println("Improved text copied to clipboard")
		}
	} else {
		log.Fatal("No response from OpenAI API or choices are empty")
	}
}
