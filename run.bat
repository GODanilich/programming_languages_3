@echo off

echo Starting server on 127.0.0.1:8080...
start cmd /k "go run lab3.go server 127.0.0.1:8080"

timeout /t 2 > nul

echo Starting 3 clients...

start cmd /k "go run lab3.go client 127.0.0.1:8080"

start cmd /k "go run lab3.go client 127.0.0.1:8080"

start cmd /k "go run lab3.go client 127.0.0.1:8080"

echo All clients started.
