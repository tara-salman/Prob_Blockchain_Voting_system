# /bin/sh

go run main.go 0.peer.json 0.0 1 & sleep 1&
go run main.go 1.peer.json 0.1 1 & sleep 1&
go run main.go 2.peer.json 0.2 1& sleep 1&
go run main.go 3.peer.json 0.3 1&sleep 1&
go run main.go 4.peer.json 0.5 1&sleep 1&
go run main.go 5.peer.json 0.9 1&sleep 1&
go run main.go 6.peer.json 0.6 1&sleep 1&
go run main.go 7.peer.json 0.6 2&sleep 1&
go run main.go 8.peer.json 0.6 2&sleep 1&
go run main.go 9.peer.json 0.9 2

