package repository

import (
	"time"
)

type Users struct {
	Code      string
	Name      string
	Age       int
	CreatedAt time.Time
	UpdatedAt time.Time
}
