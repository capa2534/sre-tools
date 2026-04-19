package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"strings"
)


// Cuando Loki responde, manda un JSON. Este struct le dice a Go cómo mapear ese JSON 
// a variables que podés usar. Las comillas backtick como `json:"status"` son tags — 
// le dicen a Go "este campo del struct corresponde a esta llave del JSON".
// map[string]string es un diccionario clave-valor — así es como llegan los 
// labels del stream: {"namespace": "monitoring", "pod": "loki-stack-0"}.
// [][]string es una lista de listas — cada log entry es [timestamp, mensaje].
type LokiResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Stream map[string]string `json:"stream"`
			Values [][]string        `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

func main() {
	lokiURL := "http://localhost:3100"
	namespace := ""
	filter := ""
	for i, arg := range os.Args {
    	if arg == "-filter" && i+1 < len(os.Args) {
        filter = os.Args[i+1]
    	}	
	}
	if len(os.Args) > 2 && os.Args[1] == "-n" {
		namespace = os.Args[2]
	}

	if namespace == "" {
		fmt.Println("Uso: lokitail -n <namespace>")
		os.Exit(1)
	}

	sinceStr := "30m"
	
	for i, arg := range os.Args {
		if arg == "-since" && i+1 < len(os.Args){
			sinceStr = os.Args[i+1]
		}
	} 

	duration, err := time.ParseDuration(sinceStr)
	if err != nil {
		fmt.Printf("Error en -since: %v\n", err)
		os.Exit(1)
	}
	since := time.Now().Add(-duration).UnixNano()
	query := fmt.Sprintf(`{namespace="%s"}`, namespace)
	

	url := fmt.Sprintf("%s/loki/api/v1/query_range?query=%s&start=%d&limit=50",
		lokiURL, query, since)
	
// http.Get hace el request HTTP. El defer resp.Body.Close() es importante — defer 
// ejecuta esa línea cuando la función termina, no ahí mismo. Es la forma idiomática 
// de Go de asegurarse que siempre cerrás la conexión aunque el programa falle más 
// adelante.
// io.ReadAll lee todo el body como bytes, y luego json.Unmarshal convierte esos bytes
// al struct.
		
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error consultando Loki: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error leyendo respuesta: %v\n", err)
		os.Exit(1)
	}

	var lokiResp LokiResponse
	if err := json.Unmarshal(body, &lokiResp); err != nil {
		fmt.Printf("Error parseando JSON: %v\n", err)
		os.Exit(1)
	}

	for _, stream := range lokiResp.Data.Result {
    pod := stream.Stream["pod"]
    for _, value := range stream.Values {
        line := value[1]
        if filter == "" || strings.Contains(line, filter) {
            fmt.Printf("[%s] %s\n", pod, line)
        	}
    	}
	}
}
