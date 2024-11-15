package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s [-list|-l|-ls] [-file filename] [-format text|json]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -list|-l|-ls\n\tList all files\n")
		fmt.Fprintf(os.Stderr, "  -file string\n\tSpecific file to retrieve (if not specified, lists all files)\n")
		fmt.Fprintf(os.Stderr, "  -format string\n\tOutput format: json or text (default \"text\")\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s                    # Show this help\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -list              # List all files\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -file=example.txt  # Get contents of example.txt\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -format=json       # List all files in JSON format\n", os.Args[0])
	}

	// Add multipls flags for list
	list := flag.Bool("list", false, "List all files")
	listShort := flag.Bool("l", false, "List all files")
	listLs := flag.Bool("ls", false, "List all files")

	format := flag.String("format", "text", "Output format: (json or text)")
	file := flag.String("file", "", "Specific file to retrieve (optional)")
	flag.Parse()

	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(0)
	}

	// Check is the list flag is set
	isList := *list || *listShort || *listLs

	baseUrl := "http://localhost:3000/files"
	url := baseUrl

	// If the file is specified and we're not listing, get the file
	if *file != "" && !isList {
		url = fmt.Sprintf("%s/%s", baseUrl, *file)
	}
	url = fmt.Sprintf("%s?format=%s", url, *format)

	resp, err := http.Get(url)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error making request %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error: Server returned status %d\n", resp.StatusCode)
		io.Copy(os.Stderr, resp.Body)
		os.Exit(1)
	}

	io.Copy(os.Stdout, resp.Body)
	fmt.Println() // Add a newline at the end
}
