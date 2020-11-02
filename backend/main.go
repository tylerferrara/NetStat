package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	validator "gopkg.in/validator.v2"
)

// TODO: SQL ATTACK PREVENTION (aka: sanitize all inputs)
// 			404 errors
//			200 success

// Postgresql's timestamp format
const timeFormat = "2006-01-02T15:04:05Z07:00"

// sendResult sends a JSON payload to the given response writer
func sendResult(w http.ResponseWriter, v interface{}) {
	// enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET,HEAD,OPTIONS,POST,PUT")
	w.Header().Set("Access-Control-Allow-Headers", "Access-Control-Allow-Headers, Origin,Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers")
	reply, err := json.Marshal(v)
	if err != nil {
		log.Println(err)
	}
	fmt.Fprintf(w, "%s", reply)
}

// ROUTE: /login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	/* Validation
	- cridentials match
	- citizen is registered
	- citizen has not voted
	*/
	db := mountDB()
	defer db.Close()
	// define JSON structures
	type response struct {
		SSN string `validate:"max=6,regexp=^[0-9]*$"`
		DOB string `validate:"max=10,regexp=^(0[1-9]|1[012])-(0[1-9]|[12][0-9]|3[01])-(19|20)[0-9][0-9]$"`
	}
	type payload struct {
		SSN      string
		DOB      string
		Eligible bool
	}
	// read response data
	var pay = payload{"", "", false}
	var respData response
	respRaw, respErr := ioutil.ReadAll(r.Body)
	if respErr != nil {
		log.Println(respErr)
		sendResult(w, pay)
		return
	}
	json.Unmarshal(respRaw, &respData)
	// sanitize
	if errs := validator.Validate(respData); errs != nil {
		log.Println(errs)
		sendResult(w, pay)
		return
	}
	// create query
	query := "SELECT SSN, DOB, is_registered, has_voted FROM citizen WHERE ssn=$1 AND dob=$2;"
	row := db.QueryRow(query, respData.SSN, respData.DOB)
	// get query results
	var ssn, dob string
	var isRegistered, hasVoted bool
	err := row.Scan(&ssn, &dob, &isRegistered, &hasVoted)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Error when querying: %s\n", query)
			log.Println(err)
		}
	}
	// form response
	pay.SSN, pay.DOB = respData.SSN, respData.DOB
	if ssn == respData.SSN && dob == respData.DOB && isRegistered && !hasVoted {
		pay.Eligible = true
	}
	// send JSON
	sendResult(w, pay)
}

// ROUTE: /register
func registerHandler(w http.ResponseWriter, r *http.Request) {
	/* Validation
	- citizen exits
	- citizen has not registered
	- citizen has not voted (edge case)
	*/
	db := mountDB()
	defer db.Close()
	// define JSON structures
	type response struct {
		SSN string `validate:"max=6,regexp=^[0-9]*$"`
		DOB string `validate:"max=10,regexp=^(0[1-9]|1[012])-(0[1-9]|[12][0-9]|3[01])-(19|20)[0-9][0-9]$"`
	}
	type payload struct {
		SSN     string
		DOB     string
		Success bool
		Message string
	}
	var pay = payload{"", "", false, ""}
	// read response data
	var respData response
	respRaw, respErr := ioutil.ReadAll(r.Body)
	if respErr != nil {
		log.Println(respErr)
		pay.Message = "Malformed JSON"
		sendResult(w, pay)
		return
	}
	json.Unmarshal(respRaw, &respData)
	// sanitize
	if errs := validator.Validate(respData); errs != nil {
		log.Println(errs)
		pay.Message = "Invalid input format"
		sendResult(w, pay)
		return
	}
	pay.SSN, pay.DOB = respData.SSN, respData.DOB
	// does this citizen exist?
	query := "SELECT SSN, DOB, is_registered, has_voted FROM citizen WHERE ssn=$1 AND dob=$2;"
	row := db.QueryRow(query, respData.SSN, respData.DOB)
	var ssn, dob string
	var isRegistered, hasVoted bool
	err := row.Scan(&ssn, &dob, &isRegistered, &hasVoted)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Error when querying: %s\n", query)
			log.Println(err)
		}
	}
	if ssn != respData.SSN || dob != respData.DOB {
		pay.Message = "Not a citizen"
		sendResult(w, pay)
		return
	}
	if isRegistered {
		pay.Message = "Already registered"
		sendResult(w, pay)
		return
	}
	if hasVoted {
		// we should NEVER get here
		log.Println("[CRITICAL] citizen voted without registering!")
		log.Println(respData)
		pay.Message = "Already voted"
		sendResult(w, pay)
		return
	}

	query = "UPDATE citizen SET is_registered=TRUE WHERE ssn=$1 AND dob=$2;"
	_, execErr := db.Exec(query, ssn, dob)
	if execErr != nil {
		log.Println("Failed to register citizen!")
		log.Println(respData)
		log.Println(execErr)
		pay.Message = "Encountered an error trying to register"
		sendResult(w, pay)
		return
	}
	pay.Success = true
	pay.Message = "Citizen is now registered"

	sendResult(w, pay)
}

// ROUTE: /vote
func voteHandler(w http.ResponseWriter, r *http.Request) {
	/* Validation
	- candidate exists
	- citizen exits
	- citizen is registered
	- citizen has not voted
	- vote is cast within election time window
	*/
	db := mountDB()
	defer db.Close()
	// define JSON structures
	type response struct {
		SSN       string `validate:"max=6,regexp=^[0-9]*$"`
		DOB       string `validate:"max=10,regexp=^(0[1-9]|1[012])-(0[1-9]|[12][0-9]|3[01])-(19|20)[0-9][0-9]$"`
		Candidate string `validate:"max=20,regexp^[a-zA-Z]*$"`
	}
	type payload struct {
		SSN     string
		DOB     string
		Success bool
		Message string
	}
	// read response data
	var pay = payload{"", "", false, ""}
	var respData response
	respRaw, respErr := ioutil.ReadAll(r.Body)
	if respErr != nil {
		log.Println(respErr)
		pay.Message = "Malformed JSON"
		sendResult(w, pay)
		return
	}
	json.Unmarshal(respRaw, &respData)
	// sanitize
	if errs := validator.Validate(respData); errs != nil {
		log.Println(errs)
		pay.Message = "Invalid input format"
		sendResult(w, pay)
		return
	}
	pay.SSN, pay.DOB = respData.SSN, respData.DOB
	// get citizen
	query := "SELECT id, SSN, DOB, is_registered, has_voted FROM citizen WHERE ssn=$1 AND dob=$2;"
	row := db.QueryRow(query, respData.SSN, respData.DOB)
	var pkCitizen int
	var ssn, dob string
	var isRegistered, hasVoted bool
	err := row.Scan(&pkCitizen, &ssn, &dob, &isRegistered, &hasVoted)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Error when querying: %s\n", query)
			log.Println(err)
		}
	}
	// validate citizen
	if respData.SSN != ssn || respData.DOB != dob {
		pay.Message = "Not a citizen"
		sendResult(w, pay)
		return
	}
	if !isRegistered {
		pay.Message = "Not registered"
		sendResult(w, pay)
		return
	}
	if hasVoted {
		pay.Message = "Citizen already voted"
		sendResult(w, pay)
		return
	}

	query = "SELECT id, fk_election, name FROM candidate WHERE name=$1;"
	row = db.QueryRow(query, respData.Candidate)
	var candidate string
	var pkCandidate, fkElection int
	err = row.Scan(&pkCandidate, &fkElection, &candidate)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Error when querying: %s\n", query)
			log.Println(err)
		}
	}
	if candidate == "" || candidate != respData.Candidate {
		pay.Message = "Invalid candidate"
		sendResult(w, pay)
		return
	}

	query = "SELECT start_time, end_time FROM election WHERE id=$1;"
	row = db.QueryRow(query, fkElection)
	var startStr, endStr string
	err = row.Scan(&startStr, &endStr)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Error when querying: %s\n", query)
			log.Println(err)
		}
	}
	// centralize
	now := time.Now()
	var startTime, endTime time.Time
	startTime, err = time.Parse(timeFormat, startStr)
	if err != nil {
		fmt.Println(err)
	}
	endTime, err = time.Parse(timeFormat, endStr)
	if err != nil {
		fmt.Println(err)
	}

	if now.After(endTime) || now.Before(startTime) {
		pay.Message = "Vote is outside of election time window"
		sendResult(w, pay)
		return
	}

	// mark citizen as voted
	query = "UPDATE citizen SET has_voted=TRUE WHERE ssn=$1 AND dob=$2;"
	_, execErr := db.Exec(query, ssn, dob)
	if execErr != nil {
		log.Println(execErr)
		pay.Message = "Encountered an error trying to cast vote"
		sendResult(w, pay)
		return
	}

	// create new vote
	query = "INSERT INTO vote (fk_election, fk_citizen, fk_candidate, vote_time) VALUES ($1, $2, $3, current_timestamp);"
	_, resErr := db.Exec(query, fkElection, pkCitizen, pkCandidate)
	if resErr != nil {
		log.Println(resErr)
		pay.Message = "Encountered an error trying to insert vote"
		sendResult(w, pay)
		return
	}

	pay.Success = true
	pay.Message = "Vote counted"
	sendResult(w, pay)
}

// ROUTE: /results
func resultsHandler(w http.ResponseWriter, r *http.Request) {
	/* Validation
	- candidate exists
	- candidate belongs to an election
	- checks all votes are within election time window
	*/
	db := mountDB()
	defer db.Close()
	// define JSON structures
	type response struct {
		Candidate string `validate:"max=20,regexp=^[a-zA-Z]*$"`
	}
	type payload struct {
		Candidate string // add a success and message
		Votes     int
		Final     bool
	}
	// read response data
	var pay = payload{"", 0, false}
	var respData response
	respRaw, respErr := ioutil.ReadAll(r.Body)
	if respErr != nil {
		log.Println(respErr)
		sendResult(w, pay)
		return
	}
	json.Unmarshal(respRaw, &respData)
	// sanitize
	if errs := validator.Validate(respData); errs != nil {
		log.Println(errs)
		sendResult(w, pay)
		return
	}
	pay.Candidate = respData.Candidate
	// get the candidate and election primary keys
	query := "SELECT id, fk_election, name FROM candidate WHERE name=$1;"
	row := db.QueryRow(query, respData.Candidate)
	var pkCandidate, pkElection int
	var candidate string
	err := row.Scan(&pkCandidate, &pkElection, &candidate)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Error when querying: %s\n", query)
			log.Println(err)
		}
	}
	if candidate == "" || candidate != respData.Candidate {
		pay.Candidate = ""
		sendResult(w, pay)
		return
	}
	// check election end timestamp
	query = "SELECT end_time FROM election WHERE id=$1;"
	row = db.QueryRow(query, pkElection)
	var endStr string
	err = row.Scan(&endStr)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("No election found! Query: %s\n", query)
			log.Println(err)
		}
	}
	now := time.Now()
	endTime, tErr := time.Parse(timeFormat, endStr)
	if tErr != nil {
		log.Println(tErr)
	}
	if now.After(endTime) {
		pay.Final = true
	}
	// count votes
	query = "SELECT vote_time FROM vote WHERE fk_election=$1 AND fk_candidate=$2;"
	rows, rErr := db.Query(query, pkElection, pkCandidate)
	if rErr != nil {
		log.Println(rErr)
	}
	defer rows.Close()
	for rows.Next() {
		// validate the date
		var voteStr string
		if err := rows.Scan(&voteStr); err != nil {
			log.Println("[ERROR] Unable to scan vote!")
			pay.Votes = -1
			sendResult(w, pay)
			return
		}
		voteTime, tErr := time.Parse(timeFormat, voteStr)
		if tErr != nil {
			log.Println("[ERROR] Unable to read vote time!")
			log.Println(tErr)
			pay.Votes = -1
			sendResult(w, pay)
			return
		}
		if voteTime.Before(endTime) {
			pay.Votes = pay.Votes + 1
		}
	}
	sendResult(w, pay)
}

// mountDB retrieves an open connection to the SQL database
func mountDB() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		os.Getenv("PG_HOST"), os.Getenv("PG_PORT"), os.Getenv("PG_USER"), os.Getenv("PG_PASSWORD"), os.Getenv("PG_DB"))

	d, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	return d
}

func main() {
	// define routes
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/vote", voteHandler)
	http.HandleFunc("/results", resultsHandler)
	// start web server
	log.Fatal(http.ListenAndServe(":8080", nil))
}
