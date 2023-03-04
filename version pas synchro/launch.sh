#!/bin/bash
go build server-tcp.go
gnome-terminal -e 'bash -c "go run server-tcp.go"'

cd jeu
go build -race
gnome-terminal -e 'bash -c "./course"'
sleep 2
gnome-terminal -e 'bash -c "./course"'
sleep 2
gnome-terminal -e 'bash -c "./course"'
sleep 2
gnome-terminal -e 'bash -c "./course"'
sleep 2


