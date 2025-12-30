module github.com/matperez/group-photo-reaction-bot/admin

go 1.25.5

replace github.com/matperez/group-photo-reaction-bot/bot => ../bot

require (
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/lib/pq v1.10.9
	github.com/pressly/goose/v3 v3.26.0
	golang.org/x/crypto v0.46.0
)

require (
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
)
