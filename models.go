package main
import (
	"time"
)

type ApiResponse struct {
	Url string
	Status int
	ResponseRaw string
	Duration time.Duration
	Date time.Time
	HasError int
}

