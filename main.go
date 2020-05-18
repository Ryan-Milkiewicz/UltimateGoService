package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"garagesale/schema"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {

	// ============================================================================
	// App Starting

	log.Printf("main : Started")
	defer log.Println("main : Compleated")

	// ============================================================================
	// Setup Dependencies

	db, err := openDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	flag.Parse()
	switch flag.Arg(0) {
	case "migrate":
		if err := schema.Migrate(db); err != nil {
			log.Fatal("applying migrations", err)
		}
		log.Println("Migrations complete")
		return
	case "seed":
		if err := schema.Seed(db); err != nil {
			log.Fatal("applying seed data", err)
		}
		log.Println("Seed data inserted")
		return

	}

	// ============================================================================
	// Start Api Service

	ps := ProductService{db: db}

	api := http.Server{
		Addr:         "localhost:8000",
		Handler:      http.HandlerFunc(ps.List),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we dont collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for request
	go func() {
		log.Printf("main : API listening on %s", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatalf("error: listening and serving: %s", err)

	case <-shutdown:
		log.Println("main : Start shutdown")

		const timeout = 5 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		err := api.Shutdown(ctx)
		if err != nil {
			log.Printf("main : Graceful shutdown did not compleate in %v : %v", timeout, err)
			err = api.Close()
		}

		if err != nil {
			log.Fatalf("main : could not stop server gracefully : %v", err)
		}
	}
}

func openDB() (*sqlx.DB, error) {
	q := url.Values{}
	q.Set("sslmode", "disable")
	q.Set("timezone", "utc")

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword("postgres", "postgres"),
		Host:     "localhost",
		Path:     "postgres",
		RawQuery: q.Encode(),
	}

	return sqlx.Open("postgres", u.String())
}

//Product is something we sell
type Product struct {
	ID          string    `db:"product_id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Cost        int       `db:"cost" json:"cost"`
	Quantity    int       `db:"quantity" json:"quantity"`
	DateCreated time.Time `db:"date_created" json:"date_created"`
	DateUpdated time.Time `db:"date_updated" json:"date_updated"`
}

// Product Service has handler for dealing with products
type ProductService struct {
	db *sqlx.DB
}

//ListProducts gives all products as a list
func (p *ProductService) List(w http.ResponseWriter, r *http.Request) {

	list := []Product{}

	const q = `SELECT product_id, name, cost, quantity, date_updated, date_created FROM products`

	if err := p.db.Select(&list, q); err != nil {
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
