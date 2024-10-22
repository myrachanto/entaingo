package main

import (
	"log"

	"github.com/myrachanto/entaingo/src/routes"
)

func init() {
	log.SetPrefix("entaingo :==> ")
}

// @title Entaingo  API Documention
// @version 1.0
// @description This is a entaingo API Documention server.

// @contact.name API Support
// @contact.url https://www.chantosweb.com
// @contact.email myrachanto1@gmail.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
func main() {
	log.Println("server started..........")
	routes.ApiServer()
}
