package repository

import (
	"time"
)

type Products struct {
	Asin      string
	Name      string
	Maker     string
	Price     int
	Reason    string
	Url       string
	Status    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
