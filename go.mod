module github.com/andrewmccall/geofire-go

go 1.13

require (
	cloud.google.com/go/firestore v1.5.0
	github.com/andrewmccall/geoutils v0.1.0
	google.golang.org/genproto v0.0.0-20210325141258-5636347f2b14
)

replace github.com/andrewmccall/geoutils => ../geoutils
