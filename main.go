package main

import (
	"log"
	"os"

	"github.com/allie/tdm/tdm"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	consumerKey := os.Getenv("CONSUMER_KEY")
	consumerSecret := os.Getenv("CONSUMER_SECRET")
	accessToken := os.Getenv("ACCESS_TOKEN")
	accessTokenSecret := os.Getenv("ACCESS_TOKEN_SECRET")

	client := tdm.NewTdm(consumerKey, consumerSecret, accessToken, accessTokenSecret)

	client.OpenStream()
	defer client.CloseStream()

	dms, err := client.GetDmStream()
	if err != nil {
		log.Fatalf("Failed getting stream: %v", err)
	}

	for dm := range dms {
		log.Printf("\nNew DM: %v\n", dm)
	}

	log.Printf("%v", dms)
}
