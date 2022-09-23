package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	m "pseudo-wordle.com/model"
	util "pseudo-wordle.com/util"
	"rsc.io/quote"
)

// fallback if value are not present in configuration
var LEN_WORD int = 5
var TOTAL_GUESSES int = 5

const pathConfig string = "config/conf.json"
const pathDictionary string = "./data/WordDictionary.json"

var obj m.WordDatabase
var filteredDatabase []string
var indexWord int
var word string
var numberGuesses int = 0
var previousGuesses []m.GuessScores

type guessHandler struct{}
type dictionaryHandler struct{}

/**
* Middleware example
 */
func mdw(in http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r == nil || r.Body == nil {
			http.Error(w, "request is not valid", http.StatusBadRequest)
			return
		}
		// %q put quotes to the string
		log.Printf("received a %q request from the service at %q of %s", r.Method, r.RequestURI, strings.Split(r.RemoteAddr, ":")[0])

		defer r.Body.Close() // call close() everytime, similar to finally

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "error reading the body", http.StatusBadRequest)
			return
		}
		if r.Method == http.MethodPost {
			log.Printf("the body of the request was %q", string(body))
		}

		// resetting the pointer of the row of body so that we can read it again from the beginning
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		in.ServeHTTP(w, r)
	})
}

/**
* Handling guess requests.
 */
func (h *guessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r == nil || r.Body == nil {
		http.Error(w, "request is not valid", http.StatusBadRequest)
		return
	}
	defer r.Body.Close() // call close() everytime, similar to finally

	guessRequest := m.JsonGuessRequest{}

	body, err := io.ReadAll(r.Body)
	handlingError(err, "", w)

	err = json.Unmarshal(body, &guessRequest)
	handlingError(err, "", w)

	// what happen when an user call the endpoint /guess
	userGuess := strings.ToLower(guessRequest.Guess)

	// the user is asking a status, can be created another API /status
	if len(userGuess) != LEN_WORD {
		log.Printf("The word %s must be long %d and it was %d\n", userGuess, LEN_WORD, len(userGuess))
		// giving a status without adding anything
		json.NewEncoder(w).Encode(buildGuessResponse(0))
		return
	}
	previousGuesses = append(previousGuesses, m.GuessScores{Word: []rune(userGuess), Score: guessScore(userGuess)})

	for _, e := range previousGuesses {
		log.Printf("%v\n", e)
	}

	numberGuesses++
	if strings.Compare(userGuess, word) == 0 {
		log.Println("The user WON")
		json.NewEncoder(w).Encode(buildGuessResponse(1))
		reset(w, r)
	} else if TOTAL_GUESSES-numberGuesses == 0 {
		log.Println("No more guesses, word reseted!")
		json.NewEncoder(w).Encode(buildGuessResponse(-1))
		reset(w, r)
	} else {
		log.Printf("guesses remaining %d\n", TOTAL_GUESSES-numberGuesses)
		json.NewEncoder(w).Encode(buildGuessResponse(0))
	}

}

/**
* Handling dictionary requests.
 */
func (h *dictionaryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r == nil || r.Body == nil {
		http.Error(w, "request is not valid", http.StatusBadRequest)
		return
	}
	defer r.Body.Close() // call close() everytime, similar to finally

	dictionaryRequest := m.JsonDictionaryRequest{}

	body, err := io.ReadAll(r.Body)
	handlingError(err, "", w)

	err = json.Unmarshal(body, &dictionaryRequest)
	handlingError(err, "", w)

	// what happen when an user call the endpoint /dictionary

	json.NewEncoder(w).Encode(buildDictionaryResponse(dictionaryRequest.Starting, dictionaryRequest.Ending, dictionaryRequest.Filter))
}

func buildGuessResponse(verdict int) m.JsonGuessResponse {
	var arrayStringGuessScores []m.StringGuessScores

	for _, e := range previousGuesses {
		var stringGuesses m.StringGuessScores
		stringGuesses.Score = e.Score
		stringGuesses.Word = string(e.Word[:])

		arrayStringGuessScores = append(arrayStringGuessScores, stringGuesses)
	}
	log.Printf("building jsonGuessResponse %v", arrayStringGuessScores)

	return m.JsonGuessResponse{Guesses: arrayStringGuessScores, Verdict: verdict, GuessesRemaining: TOTAL_GUESSES - numberGuesses}
}

func buildInfoResponse() m.JsonInfoResponse {
	log.Printf("building jsonInfoResponse length word:%d total guess:%d", LEN_WORD, TOTAL_GUESSES)
	return m.JsonInfoResponse{LEN_WORD: LEN_WORD, TOTAL_GUESSES: TOTAL_GUESSES}
}

/**
* Building the dictionary response using the different options given by the request.
* @param starting integer that gives the start of the array.
* @param ending integer that gives the end of the array.
* @param filter string to filter by the prefix.
*
 */
func buildDictionaryResponse(starting int, ending int, filter string) m.JsonDictionaryResponse {
	log.Printf("building jsonDictionaryResponse %d %d", starting, ending)

	filteredDictionary := obj.Dictionary

	if len(filter) != 0 {
		filteredDictionary = util.Filter(filteredDictionary, func(s string) bool {
			return strings.HasPrefix(s, filter)
		})
	}

	if ending == 0 || ending > len(filteredDictionary) {
		ending = len(filteredDictionary)
	}
	if starting > len(filteredDictionary) {
		starting = len(filteredDictionary)
	}

	return m.JsonDictionaryResponse{Dictionary: filteredDictionary[starting:ending], Total: len(filteredDictionary), Starting: starting, Ending: ending, Filter: filter}
}

/**
* Building the reset response.
* Similar to the guess response but with the verdict to -1.
 */
func buildResetResponse(verdict int) m.JsonGuessResponse {
	return buildGuessResponse(verdict)
}

/**
* Generic function for handling errors.
 */
func handlingError(e error, msg string, channel http.ResponseWriter) {
	if e != nil {

		if len(msg) == 0 {
			http.Error(channel, e.Error(), http.StatusBadRequest)
		} else {
			http.Error(channel, msg, http.StatusBadRequest)
		}
		return
	}
}

/**
* Calculating the score of the word guessed.
 */
func guessScore(userGuess string) []int {
	var score = make([]int, LEN_WORD)

	for i, e := range userGuess {

		// if the character is present => score = 1
		if strings.ContainsRune(word, e) {
			score[i] = 1
		}
		// if the character is present and equals => score = 2
		if score[i] == 1 && word[i] == byte(e) {
			score[i] = 2
		}
	}

	return score
}

/**
* Setting a new word.
 */
func reset(w http.ResponseWriter, r *http.Request) {
	// setting new word
	indexWord = rand.Intn(len(filteredDatabase))
	numberGuesses = 0
	previousGuesses = previousGuesses[:0] // resetting the array of guesses
	word = filteredDatabase[indexWord]

	log.Printf("The word for this game is %q!\n", word)
}

/**
*	Setting up the options given the config file in pathConfig.
 */
func getConfiguration() {
	file, err := os.Open(pathConfig)

	if err != nil { //is there a file
		log.Fatal("error:", err)
		return
	}

	fi, err := file.Stat()
	if err != nil { //cannot retrieve stats
		log.Fatal("error:", err)
		return
	}

	if fi.Size() == 0 { // it is empty
		log.Printf("The file %q is empty, using default values", pathConfig)
		return
	}

	decoder := json.NewDecoder(file)
	configuration := m.Configuration{}
	err = decoder.Decode(&configuration)
	if err != nil { // problems reading json
		log.Fatal("error:", err)
	}

	if configuration.LEN_WORD != 0 {
		LEN_WORD = configuration.LEN_WORD
	}
	if configuration.TOTAL_GUESSES != 0 {
		TOTAL_GUESSES = configuration.TOTAL_GUESSES
	}

	defer file.Close() //always close it
}

/**
*	Setting up the Dictionary in pathDictionary used during the game.
 */
func getDictionary() {
	// reading the "Database"
	data, err := ioutil.ReadFile(pathDictionary)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
		return
	}

	// unmarshall it
	err = json.Unmarshal(data, &obj)
	if err != nil {
		log.Fatal("error:", err)
		return
	}

	// filtering the DATABASE given the length of the possible words
	filteredDatabase = util.Filter(obj.Dictionary, func(s string) bool {
		return len(s) == LEN_WORD
	})
}

/**
* Main function.
 */
func main() {
	log.Printf("The word for this game is %q!\n", word)

	// server mux
	mux := http.NewServeMux()
	var h http.Handler = &guessHandler{}    // <=> h := http.Handler(&guessHandler{})
	d := http.Handler(&dictionaryHandler{}) // <=> var d http.Handler = &dictionaryHandler{}

	mux.Handle("/guess", h)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(buildInfoResponse())
	})

	// reseting the word by user request
	mux.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("The word %q is resetted!\n", word)
		reset(w, r)
		json.NewEncoder(w).Encode(buildResetResponse(-1))
	})

	mux.Handle("/dictionary", d)

	log.Printf("There are %d words of %d character in the dictionary\n", len(filteredDatabase), LEN_WORD)

	// printing some inspirational quote when the server starts
	log.Println(quote.Go())

	log.Fatal(http.ListenAndServe(":8080", mdw(mux)))

}

/**
* Initialization.
 */
func init() {
	// seeding a random seed
	rand.Seed(time.Now().UnixNano())

	// configuration and database init
	getConfiguration()
	getDictionary()

	// initializing the word of the game
	indexWord = rand.Intn(len(filteredDatabase))
	word = filteredDatabase[indexWord]
	previousGuesses = previousGuesses[:0] // can use also make([]guessScores)
}
