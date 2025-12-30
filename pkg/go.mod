module github.com/matperez/group-photo-reaction-bot/pkg

go 1.25.5

require (
	github.com/matperez/group-photo-reaction-bot/bot v0.0.0
	github.com/mattn/go-sqlite3 v1.14.32
)

replace github.com/matperez/group-photo-reaction-bot/bot => ../bot

