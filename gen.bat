@echo off
protoc -I=proto --go_out=protocol proto/*.proto