package hist

import "time"

type Response struct {
	ID    int64  `json:"id"`
	IP    string `json:"ip"`
	Token string `json:"token"`
}

type Request struct {
	ID             int64     `json:"id"`
	IP             string    `json:"ip"`
	NumberRequests int64     `json:"number_requests"`
	Valid          bool      `json:"valid"`
	ExpiredAt      time.Time `json:"expired_at"`
}
