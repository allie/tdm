package main

import (
	"log"
	"os"

	"github.com/allie/tdm/tdm"
	"github.com/allie/tdm/gui"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	consumerKey := os.Getenv("CONSUMER_KEY")
	consumerSecret := os.Getenv("CONSUMER_SECRET")
	accessToken := os.Getenv("ACCESS_TOKEN")
	accessTokenSecret := os.Getenv("ACCESS_TOKEN_SECRET")

	client, err := tdm.NewTdm(consumerKey, consumerSecret, accessToken, accessTokenSecret)
	if err != nil {
		log.Fatalf("Failed creating tdm client: %v", err)
	}

	client.Log()

	gui := gui.NewGui()
	gui.Init()
	gui.Loop()
}
