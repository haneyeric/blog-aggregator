package main

import (
	"github.com/haneyeric/blog-aggregator/internal/config"
	"github.com/haneyeric/blog-aggregator/internal/database"
)

type state struct {
	db   *database.Queries
	cfg *config.Config
}
