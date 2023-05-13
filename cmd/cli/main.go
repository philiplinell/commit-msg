package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env"
	"github.com/philiplinell/commit-msg/internal/build"
	"github.com/philiplinell/commit-msg/internal/commitassist"
	"github.com/philiplinell/commit-msg/internal/openai"
	"github.com/urfave/cli"
)

type config struct {
	APIKey string `env:"OPENAI_API_KEY"`
}

//nolint:gochecknoglobals
var (
	costFlag    bool
	timeoutFlag string
	filename    string
)

func main() {
	buildInfo, err := build.GetInfo()
	if err != nil {
		log.Fatal(err)
	}

	version := buildInfo.String()

	app := &cli.App{
		Name:        "commit-msg",
		Usage:       "A CLI tool to suggest commit messages",
		Description: "commit-msg will suggest a commit message from filenames and the lines changed.",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "cost",
				Usage:       "if the cost should be printed to stdout",
				Destination: &costFlag,
			},
			&cli.StringFlag{
				Name:        "timeout",
				Usage:       "the timeout for the request to OpenAI",
				Value:       "5s",
				Destination: &timeoutFlag,
			},
			&cli.StringFlag{
				Name:        "file",
				Usage:       "the file where the changes are. Usually this will be $COMMIT_MSG_FILE set in prepare-commit-msg hook",
				Destination: &filename,
				Required:    true,
			},
		},
		Action:  cliAction,
		Version: version,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func cliAction(_ *cli.Context) error {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	httpClient := http.DefaultClient

	openAiClient := openai.NewClient(httpClient, cfg.APIKey)
	commitClient := commitassist.New(openAiClient)

	timeout, err := time.ParseDuration(timeoutFlag)
	if err != nil {
		log.Fatalf("could not parse timeout duration: %s", err)
	}

	requestContext, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var response commitassist.GetTypeResponse

	gitDiff, err := readFile()
	if err != nil {
		//nolint:gocritic
		log.Fatalf("could not read file %q: %s", filename, err)
	}

	response, err = commitClient.GetCommitMessage(requestContext, gitDiff)

	if err != nil {
		handleError(err)
	}

	fmt.Println(response.Message)

	if costFlag {
		fmt.Printf("Cost %.2f cent\n", response.Cost)
	}

	return nil
}

func handleError(err error) {
	switch e := err.(type) {
	case commitassist.UnsureError:
		fmt.Println(e)
		os.Exit(3)
	case commitassist.UnexpectedStateError:
		fmt.Println("Unexpected number of messages returned")
		os.Exit(2)
	default:
		if errors.Is(e, context.DeadlineExceeded) {
			fmt.Println("Request timed out.")
			fmt.Printf("See API status at %q\n", "https://status.openai.com/")
			fmt.Println("or try again with a longer timeout (see --timeout flag).")
			os.Exit(4)
		}
		fmt.Printf("Unknown error %v\n", e)
		os.Exit(5)
	}
}

// readFile will read the file contents, ignoring lines starting with #
// (comments) and return a string.
func readFile() (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("open file %q: %w", filename, err)
	}
	defer file.Close()

	fileScanner := bufio.NewScanner(file)

	sb := strings.Builder{}

	for fileScanner.Scan() {
		currentLine := fileScanner.Text()
		if strings.HasPrefix(currentLine, "#") {
			continue
		}
		sb.WriteString(currentLine)
	}

	return sb.String(), nil
}
