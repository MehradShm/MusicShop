package main

import (
	"log"
	"net/http"
	"os"

	handlers "D:\\Dev_Projects\\MusicShop_git\\MusicShop\\Backend\\internal\\http"
	repo "D:\\Dev_Projects\\MusicShop_git\\MusicShop\\Backend\\internal\\repository"
)

func main() {
	port := env("PORT", "8080")
	dsn := env("DATABASE_URL", "")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	r, err := repo.NewPostgresRepo(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	mux := http.NewServeMux()
	s := &handlers.Server{Repo: r}
	s.Routes(mux) // or s.routes(mux) depending on export

	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
