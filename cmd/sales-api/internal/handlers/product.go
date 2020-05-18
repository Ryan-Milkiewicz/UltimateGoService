package handlers

import (
	"encoding/json"
	"garagesale/internal/product"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
)

// Product Service has handler for dealing with products
type Product struct {
	DB *sqlx.DB
}

//ListProducts gives all products as a list
func (p *Product) List(w http.ResponseWriter, r *http.Request) {

	list, err := product.List(p.DB)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("error querying db", err)
		return
	}

	data, err := json.Marshal(list)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("error marshalling", err)
		return
	}

	w.Header().Set("content-type", "application/json: charset=utf8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		log.Println("error writing", err)
	}
}
