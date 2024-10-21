package main

import (
	"log"

	"github.com/myrachanto/entaingo/src/routes"
)

func init() {
	log.SetPrefix("entaingo :==> ")
}
func main() {
	log.Println("server started..........")
	routes.ApiServer()
}
