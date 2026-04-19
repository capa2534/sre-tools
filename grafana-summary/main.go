package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type AlertRule struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	For   string `json:"for"`
}

func main() {
	grafanaURL := "http://localhost:3000"
	token := ""

	for i, arg := range os.Args {
		if arg == "--token" && i+1 < len(os.Args) {
			token = os.Args[i+1]
		}
	}

	if token == "" {
		fmt.Println("Uso: grafana-summary --token <api-token>")
		os.Exit(1)
	}

	req, err := http.NewRequest("GET", grafanaURL+"/api/v1/provisioning/alert-rules", nil)
	if err != nil {
		fmt.Printf("Error creando request: %v\n", err)
		os.Exit(1)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error consultando Grafana: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error leyendo respuesta: %v\n", err)
		os.Exit(1)
	}

	var alerts []AlertRule
	if err := json.Unmarshal(body, &alerts); err != nil {
		fmt.Printf("Error parseando JSON: %v\n", err)
		os.Exit(1)
	}

	if len(alerts) == 0 {
		fmt.Println("No hay alertas configuradas.")
		os.Exit(0)
	}

	fmt.Printf("Alertas configuradas: %d\n\n", len(alerts))
	for _, alert := range alerts {
		fmt.Printf("⚠️  %s (for: %s)\n", alert.Title, alert.For)
	}
}
