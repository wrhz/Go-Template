package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/PuerkitoBio/goquery"
	"github.com/mlctrez/wasmexec"
)

func handleElement(w *bufio.Writer, r io.Reader, tagName string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
    if err != nil {
        return "", err
    }

    element := doc.Find(tagName).First()

	text := element.Text()

	element.ReplaceWithHtml("<script src=\"wasm_exec.js\"></script>\n<script src=\"gt-runtime.js\"></script>")

	err = goquery.Render(w, doc.Selection)
    if err != nil {
        fmt.Printf("Error rendering HTML: %v\n", err)
    }

	w.Flush()

    return text, nil
}

func gt(wasmExecJSFilePath string, filePath string, cachePath string, htmlPath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Cannot open file: %v\n", err)
		return
	}

	htmlFile, err := os.Create(htmlPath)
	if err != nil {
		fmt.Printf("Cannot create HTML file: %v\n", err)
		return
	}

	defer file.Close()

	w := bufio.NewWriter(htmlFile)

	defer htmlFile.Close()

	data, err := handleElement(w, file, "gt-main")
	if err != nil {
		fmt.Printf("Error parsing HTML: %v\n", err)
			return
	}

	file, err = os.Create(cachePath)
	if err != nil {
		fmt.Printf("Cannot create cache file: %v\n", err)
		return
	}

	writer := bufio.NewWriter(file)

	_, err = writer.WriteString(data)
	if err != nil {
		fmt.Printf("Error writing to output file: %v\n", err)
		return
	}

	err = writer.Flush()
	if err != nil {
		fmt.Printf("Error flushing output file: %v\n", err)
		return
	}

	defer file.Close()

	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		fmt.Println("GOPATH is not set")
		return
	}

	content, err := wasmexec.Current()

	if err != nil {
		fmt.Printf("Error getting wasm_exec.js content: %v\n", err)
		return
	}

	wasmExecJSFile, err := os.Create(wasmExecJSFilePath)
	if err != nil {
		fmt.Printf("Cannot create wasm_exec.js file: %v\n", err)
		return
	}

	defer wasmExecJSFile.Close()

	_, err = wasmExecJSFile.Write(content)

	if err != nil {
		fmt.Printf("Error writing wasm_exec.js content: %v\n", err)
		return
	}

	outputPath := "./view/main.wasm"

	cmd := exec.Command("go", "build", "-o", outputPath, cachePath)
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error building Go code: %v\n", err)
		return
	}
}