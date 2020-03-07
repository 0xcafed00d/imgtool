#!/bin/bash

mkdir imgtool_builds -p

cp README.md imgtool_builds
env GOOS=linux GOARCH=amd64 go build -o imgtool_builds/linux_64bit/imgtool *.go
env GOOS=linux GOARCH=arm go build -o imgtool_builds/linux_arm/imgtool *.go
env GOOS=darwin GOARCH=amd64 go build -o imgtool_builds/mac_64bit/imgtool *.go
env GOOS=windows GOARCH=386 go build -o imgtool_builds/win/imgtool.exe *.go

rm imgtool_builds.zip
zip -r imgtool_builds.zip imgtool_builds 