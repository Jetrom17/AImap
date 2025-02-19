package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/mattn/go-isatty"
)

// Variáveis de cores ANSI
var (
	RED, GREEN, YELLOW, BLUE, NC string
)

func init() {
	if supportsColor() {
		RED = "\033[1;31m"
		GREEN = "\033[1;32m"
		YELLOW = "\033[1;33m"
		BLUE = "\033[1;34m"
		NC = "\033[0m"
	} else {
		RED, GREEN, YELLOW, BLUE, NC = "", "", "", "", ""
	}
}

func supportsColor() bool {
	if runtime.GOOS == "windows" {
		return false // Em Windows, requer ativação manual de cores ANSI
	}
	return isatty.IsTerminal(os.Stdout.Fd())
}

type RequestPayload struct {
	Contents []Content `json:"contents"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

type Response struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf(RED+"Uso: %s <argumentos do NMAP>\n"+NC, os.Args[0])
		os.Exit(1)
	}

	cmd := exec.Command("nmap", os.Args[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(RED + "Erro ao executar o Nmap: " + err.Error() + NC)
		os.Exit(1)
	}
	nmapOutput := string(output)

	fmt.Println(BLUE + "--- Resultado do Nmap ---" + NC)
	fmt.Println(nmapOutput)

	prompt := "Explique o seguinte resultado do NMAP sem enrolação para o profissional pentest/OSINT:\n" + nmapOutput
	payload := RequestPayload{
		Contents: []Content{{Parts: []Part{{Text: prompt}}}},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(RED + "Erro ao criar JSON: " + err.Error() + NC)
		os.Exit(1)
	}

	apiKey := "API" // https://aistudio.google.com/apikey
	apiURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=" + apiKey

	fmt.Println(YELLOW + "Consultando a IA..." + NC)
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println(RED + "Erro ao conectar à API: " + err.Error() + NC)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var response Response
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		fmt.Println(RED + "Erro ao processar resposta da API: " + err.Error() + NC)
		os.Exit(1)
	}

	if len(response.Candidates) > 0 && len(response.Candidates[0].Content.Parts) > 0 {
		fmt.Println(GREEN + "--- Explicação gerada pela IA ---" + NC)
		fmt.Println(RED + response.Candidates[0].Content.Parts[0].Text + NC)
	} else {
		fmt.Println(RED + "Erro: resposta da IA vazia." + NC)
	}
}
