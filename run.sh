#!/bin/bash

echo "Starting server on 127.0.0.1:8080..."
gnome-terminal -- bash -c "go run lab3.go server 127.0.0.1:8080" &

sleep 2

echo "Starting 3 clients..."

gnome-terminal -- bash -c "go run lab3.go client 127.0.0.1:8080"

gnome-terminal -- bash -c "go run lab3.go client 127.0.0.1:8080"

gnome-terminal -- bash -c "go run lab3.go client 127.0.0.1:8080"

echo "All clients started."
