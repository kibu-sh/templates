package backend

// generate any database model code
//go:generate sqlc compile -f ./database/sqlc.yaml

// analyze each module and generate system plumbing code
//go:generate go run github.com/kibu-sh/kibu/internal/toolchain/kibugenv2/cmd/kibugenv2 ./...

// execute a second pass analysis that gathers all providers for a wire super set
//go:generate go run github.com/kibu-sh/kibu/internal/toolchain/kibuwire/cmd/kibuwire -out gen/ ./...

// finally, run wire to build the initialization code
//go:generate wire ./...
