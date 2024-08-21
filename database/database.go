package database

import (
	"database/sql"
	"sync"
)

type DB struct {
	mux *sync.RWMutex
	DB  *sql.DB
}
