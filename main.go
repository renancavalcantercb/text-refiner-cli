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
	flag.Parse()

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
		Model: "gpt-4",
		Messages: []Message{
			{
				Role:    "user",
				Content: "Improve this text: " + inputText,
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
		fmt.Println("Improved text:")
		fmt.Println(openAIResp.Choices[0].Message.Content)
	} else {
		log.Fatal("No response from OpenAI API or choices are empty")
	}
}