# ARM / Synology build
CGO_ENABLED=0 GOARM=7 GOARCH=arm GOOS=linux go build cmd/photos
