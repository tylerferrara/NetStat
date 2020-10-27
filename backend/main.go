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
)

// TODO: SQL ATTACK PREVENTION (aka: sanitize all inputs)

// Postgresql's timestamp format (for parsing)
const timeFormat = "2006-01-02T15:04:05Z07:00"

func sendResult(w http.ResponseWriter, v interface{}) {
	reply, err := json.Marshal(v)
	if err != nil {
		log.Println(err)
	}
	fmt.Fprintf(w, "%s", reply)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	db := mountDB()
	defer db.Close()
	// define JSON structures
	type response struct {
		SSN string
		DOB string
	}
	type payload struct {
		SSN      string
		DOB      string
		Eligible bool
	}
	// read response data
	var respData response
	respRaw, respErr := ioutil.ReadAll(r.Body)
	if respErr != nil {
		log.Println(respErr)
	}
	json.Unmarshal(respRaw, &respData)
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
	var pay = payload{respData.SSN, respData.DOB, false}
	if ssn == respData.SSN && dob == respData.DOB && isRegistered && !hasVoted {
		pay.Eligible = true
	}
	// send JSON
	sendResult(w, pay)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	db := mountDB()
	defer db.Close()
	// define JSON structures
	type response struct {
		SSN string
		DOB string
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
		sendResult(w, pay)
		return
	}
	json.Unmarshal(respRaw, &respData)
	pay.SSN = respData.SSN
	pay.DOB = respData.DOB
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

func voteHandler(w http.ResponseWriter, r *http.Request) {
	db := mountDB()
	defer db.Close()
	// define JSON structures
	type response struct {
		SSN       string
		DOB       string
		Candidate string
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
	}
	json.Unmarshal(respRaw, &respData)
	pay.SSN = respData.SSN
	pay.DOB = respData.DOB
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
	if candidate != respData.Candidate {
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

func resultsHandler(w http.ResponseWriter, r *http.Request) {
	db := mountDB()
	defer db.Close()
	// define JSON structures
	type response struct {
		Candidate string
	}
	type payload struct {
		Candidate string
		Votes     int
		Final     bool
	}
	// read response data
	var pay = payload{"", 0, false}

	var respData response
	respRaw, respErr := ioutil.ReadAll(r.Body)
	if respErr != nil {
		log.Println(respErr)
	}
	json.Unmarshal(respRaw, &respData)
	pay.Candidate = respData.Candidate
	// get the candidate and election primary keys
	query := "SELECT id, fk_election, name FROM candidate WHERE name=$1;"
	row := db.QueryRow(query, respData.Candidate)
	var pkCandidate, pkElection int
	var canidate string
	err := row.Scan(&pkCandidate, &pkElection, &canidate)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Error when querying: %s\n", query)
			log.Println(err)
		}
	}
	if canidate != respData.Candidate {
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
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/vote", voteHandler)
	http.HandleFunc("/results", resultsHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
