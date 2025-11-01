module irangate-webpanel

go 1.24.0

require (
	github.com/amiridev-org/irangate-ov v0.0.0
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/gorilla/handlers v1.5.2
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/websocket v1.5.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	golang.org/x/crypto v0.43.0
)

replace github.com/amiridev-org/irangate-ov => ../../

require (
	github.com/fatih/color v1.18.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/net v0.45.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
)
