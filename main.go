package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func writeLog(file *os.File, lines chan string) {

	for line := range lines {

		file.WriteString(line)

		file.Sync()
	}

}

func main() {

	address := flag.String("address", "localhost:6379", "redis address")
	password := flag.String("password", "", "redis password")
	filename := flag.String("filename", "output.csv", "output file name")

	flag.Parse()

	if _, err := os.Stat(*filename); err == nil {
		panic("File already exists!")
	}

	file, err := os.OpenFile(*filename,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatal(err)
	}

	lines := make(chan string)

	go writeLog(file, lines)

	_, err = file.WriteString("usedMemory,usedMemoryHuman,timestamp")

	if err != nil {
		log.Fatal(err)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     *address,
		Password: *password,
		DB:       0,
	})

	ctx := context.Background()

	for {

		info := client.Info(ctx, "memory").Val()

		regexUsedMemory := regexp.MustCompile(`used_memory:(.*?)\r\n`)
		regexUsedMemoryHuman := regexp.MustCompile(`used_memory_human:(.*?)\r\n`)

		usedMemory := regexUsedMemory.FindStringSubmatch(info)
		usedMemoryHuman := regexUsedMemoryHuman.FindStringSubmatch(info)
		currentTime := time.Now().Unix()

		newLine := fmt.Sprintf("\n%s,%s,%s", usedMemory[1], usedMemoryHuman[1], strconv.Itoa(int(currentTime)))

		lines <- newLine

		time.Sleep(time.Second)
	}

}
