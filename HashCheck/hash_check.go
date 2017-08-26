package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

func hashIdentification(timeDesc time.Time, chatID int64, firstName, lastName string) string {
	fmt.Printf("time      = %s\n", timeDesc.Format(time.RFC3339))
	date_seed := (timeDesc.Unix() + 5 * 3600) / 86400
	fmt.Printf("date_seed = %d\n", date_seed)
	hash_data := []byte(fmt.Sprintf("%s %x %x %s %x %s %x", SECRET, chatID, len(firstName), firstName, len(lastName), lastName, date_seed))
	fmt.Printf("hash_data = %s\n", hash_data)
	hash_sum := sha1.Sum(hash_data)
	return base64.RawURLEncoding.EncodeToString(hash_sum[:6])
}

func main() {
	var err error
	if len(os.Args) != 5 {
		fmt.Printf("Usage: %s time chatID firstName lastName\n\n", os.Args[0])
		return
	}
	timeStr := os.Args[1]
	var timeDesc time.Time
	if timeStr != "" {
		timeDesc, err = time.Parse(time.RFC3339, timeStr)
		if err != nil {
			log.Fatalf("Require time format: %s\n", time.RFC3339)
		}
	} else {
		timeDesc = time.Now()
	}
	chatIDStr := os.Args[2]
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		log.Fatalln("Cannot parse chatID")
	}
	firstName := os.Args[3]
	lastName := os.Args[4]
	fmt.Printf("hash      = %s\n", hashIdentification(timeDesc, int64(chatID), firstName, lastName))
}

