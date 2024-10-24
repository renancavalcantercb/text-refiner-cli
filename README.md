
# Text Refiner CLI

**Text Refiner CLI** is a command-line tool that refines text by improving grammar, clarity, and structure using the OpenAI GPT-4o-mini API.

## Features

- Refines and improves text input.
- Flexible input options: pass the text directly as a command-line argument or input it interactively.
- Powered by GPT-4o-mini from OpenAI.
- Easy to use, with API credentials securely stored in a `.env` file.

## Installation

### Prerequisites

- [Go](https://golang.org/doc/install) 1.18 or later installed on your machine.
- An [OpenAI API key](https://platform.openai.com/signup) with access to the GPT-4o-mini model.
- A `.env` file with your OpenAI API key.

### Cloning the repository

```bash
git clone https://github.com/your-username/text-refiner-cli.git
cd text-refiner-cli
```

### Installing dependencies

This project uses `godotenv` to manage environment variables. You can install it with the following command:

```bash
go get github.com/joho/godotenv
```

## Usage

### Step 1: Set up the `.env` file

Create a `.env` file in the project directory with the following content:

```bash
OPENAI_API_KEY=your_openai_api_key_here
```

Replace `your_openai_api_key_here` with your actual OpenAI API key.

### Step 2: Running the CLI

You can run the CLI in two different ways:

#### Option 1: Pass the text as an argument

```bash
go run main.go -text "This is the text you want to improve."
```

#### Option 2: Run interactively

Simply run the following command and input the text when prompted:

```bash
go run main.go
```

The program will prompt you to input the text you want to refine.

### Example Output

```bash
$ go run main.go -text "This is a test sentence that needs improvement."
Improved text:
"This is a test sentence that could be refined for clarity and precision."
```

## Error Handling

- If the API key is missing or incorrect, the program will return an error message.
- If there is an issue with the response from OpenAI, such as no text being returned, the program will notify you with an appropriate error message.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Feel free to fork this project and create pull requests with improvements or additional features. Contributions are welcome!
