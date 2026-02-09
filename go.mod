module omnibase

go 1.22

require github.com/google/adk-go v0.0.0
require github.com/jmoiron/sqlx v0.0.0
require github.com/qdrant/go-client v0.0.0

replace github.com/google/adk-go => ./third_party/adk-go
replace github.com/jmoiron/sqlx => ./third_party/sqlx
replace github.com/qdrant/go-client => ./third_party/qdrant-go-client
