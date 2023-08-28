module chirpy

go 1.21.0

require github.com/go-chi/chi/v5 v5.0.10

require internal/db v1.0.0

require golang.org/x/crypto v0.12.0 // indirect

replace internal/db => ./internal/db
