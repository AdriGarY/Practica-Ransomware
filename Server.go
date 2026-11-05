package main

import (
	"fmt"
	"net/http"
)

func loggerMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Imprime en la terminal el formato: GET /saludo
		fmt.Printf("%s %s", r.Method, r.URL.Path)
		
		// Continúa con la ejecución normal de la ruta
		next(w, r)
	}
}

func main() {
	http.HandleFunc("/clave", loggerMiddleware(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ByClaude")
	}))
	var puerto = ":9090"
	fmt.Printf("Servidor escuchando en http://localhost%s/clave\n", puerto)

	http.ListenAndServe(puerto, nil)
}

