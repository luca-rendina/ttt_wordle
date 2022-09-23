package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	m "pseudo-wordle.com/model"
)

var server string = "http://localhost"
var port string = "8080"
var bodyType string = "application/json"
var userID string = ""

/**
* Welcome page Request.
 */
func welcomePage(userId string) {
	postBody, _ := json.Marshal(map[string]string{
		"userID": userId,
	})

	responseBody := bytes.NewBuffer(postBody)

	resp, err := http.Post(server+":"+port, bodyType, responseBody)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf(err.Error(), http.StatusBadRequest)
		return
	}

	sb := string(body)
	fmt.Printf(sb)
}

/**
* Guessing to the server.
 */
func guessing(guess string) {
	postBody, _ := json.Marshal(map[string]string{
		"guess": guess,
	})

	responseBody := bytes.NewBuffer(postBody)

	resp, err := http.Post(server+":"+port+"/guess", bodyType, responseBody)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var data m.JsonGuessResponse
	err = decoder.Decode(&data)
	if err != nil {
		fmt.Printf("%T\n%s\n%#v\n", err, err, err)
	}
	for _, guessScore := range data.Guesses {
		fmt.Printf("%d %d: %s %v\n", data.GuessesRemaining, data.Verdict, string(guessScore.Word), guessScore.Score)
	}

}

func main() {
	welcomePage(userID)

	for {

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		text := scanner.Text()

		guessing(text)
	}
}
