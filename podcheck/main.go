package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")

	//en go no hay try/catch, cada funcion retorna 2 cosas: resultado y un error
	//intenta construir el config
	//si algo sale mal, imprime el error y sale
	//si err es nil, todo bien sigue
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Printf("Error building kubeconfig: %v\n", err)
		os.Exit(1)
	}

	// con el config cargado, se crea el cliente, el clientset es el objeto que tiene
	// todos los metodos para hablar con la API de kubernetes - es con un kubectl
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error creating clientset: %v\n", err)
		os.Exit(1)
	}

	//variable
	namespace := ""
	//si me pasa mas de 2 argumentso y el primero es '-n' toma el siguiente como namespace
	if len(os.Args) > 2 && os.Args[1] == "-n" {
		namespace = os.Args[2]
	}
	
	minRestarts := 0
	for i, arg := range os.Args {
		if arg == "--restarts" && i+1 < len(os.Args) {
			n, err := strconv.Atoi(os.Args[i+1])
			if err != nil {
				fmt.Println("--restarts debe ser un número")
				os.Exit(1)
			}
			minRestarts = n
		}
	}

	//llamada a la API
	// CoreV1() — el grupo de la API de Kubernetes donde viven los pods
	// Pods(namespace) — quiero pods de este namespace, si está vacío trae todos
	// List(...) — listálos todos
	// context.TODO() — por ahora ignoralo, es para cancelar operaciones, lo vemos después
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing pods: %v\n", err)
		os.Exit(1)
	}

	onlyErrors := false
	for _, arg := range os.Args {
		if arg == "--only-errors" {
			onlyErrors = true
		}
	}

	//es una lista de pods
	//range itera sobre ella - el '_' descarta el indice (0,1,2) porq no lo necesito, solo quiero el pod
	// En Go si declarás una variable y no la usás, el compilador falla con error. El _ es la forma de decirle "sé que hay un valor acá pero no me importa".
	for _, pod := range pods.Items {
		status := string(pod.Status.Phase)
		var restarts int32
		// 		Un pod puede tener varios contenedores. Phase solo te da el estado general del pod,
		// 		pero no te dice si un contenedor específico está en CrashLoopBackOff.
		// 		Por eso iterás sobre ContainerStatuses — el estado de cada contenedor individualmente.
		// 		Si alguno está en estado Waiting con razón CrashLoopBackOff, sobreescribís el status.
		// 		Esto es exactamente lo que hace kubectl get pods cuando muestra CrashLoopBackOff en la
		// 		columna STATUS.
		for _, cs := range pod.Status.ContainerStatuses {
			restarts += cs.RestartCount
			if cs.State.Waiting != nil && cs.State.Waiting.Reason == "CrashLoopBackOff" {
				status = "CrashLoopBackOff ⚠️"
			}
		}

		marker := "✓"
		if status != "Running" && status != "Succeeded" {
			marker = "✗"
		}
		if onlyErrors && marker == "✓" {
			continue
		}
		
		if minRestarts > 0 && int(restarts) < minRestarts {
			continue
		}

		fmt.Printf("%s %-50s %-20s %-25s restarts: %d\n", marker, pod.Name, pod.Namespace, status, restarts)
	}
}

