package model

type GuessScores struct {
	Word []rune
	//0 not present, 1 present but not correct position, 2 present and correct position
	Score []int
}

type StringGuessScores struct {
	Word string
	//0 not present, 1 present but not correct position, 2 present and correct position
	Score []int
}

type WordDatabase struct {
	Dictionary []string
}

type Configuration struct {
	LEN_WORD      int
	TOTAL_GUESSES int
}

type JsonGuessRequest struct {
	Guess string `json:"guess"`
}

type JsonDictionaryRequest struct {
	Starting  int    `json:"starting"`
	Ending    int    `json:"ending"`
	Character string `json:"character"`
	Filter    string `json:"filter"`
}

type JsonGuessResponse struct {
	Guesses          []StringGuessScores `json:"guesses"`
	Verdict          int                 `json:"verdict"`
	GuessesRemaining int                 `json:"guessesRemaining"`
}

type JsonInfoResponse struct {
	LEN_WORD      int `json:"length_word"`
	TOTAL_GUESSES int `json:"total_guesses"`
}

type JsonDictionaryResponse struct {
	Dictionary []string `json:"dictionary"`
	Total      int      `json:"total"`
	Starting   int      `json:"starting"`
	Ending     int      `json:"ending"`
	Filter     string   `json:"filtering"`
}
