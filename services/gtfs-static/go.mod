module github.com/teclara/railsix/gtfs-static

go 1.25.8

require (
	github.com/jamespfennell/gtfs v0.1.24
	github.com/teclara/railsix/shared v0.0.0
)

require (
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)

replace github.com/teclara/railsix/shared => ../shared
