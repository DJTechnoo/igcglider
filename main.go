package main

import (
	"encoding/json"
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/marni/goigc"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// general app constants
const root = "/paragliding"
const idArg = 4    // URL index for ID
const fieldArg = 5 // URL index for FIELD

// db constants
var dbURL = "mongodb://igcuser:igc4life@ds141783.mlab.com:41783/igc"
var dbName = "igc"
var dbCollection = "igcstruct"

// Global variables and structs
var startTime time.Time

// holds data for /igcinfo/api
type meta struct {
	Uptime  string `json:"uptime"`
	Info    string `json:"info"`
	Version string `json:"version"`
}

// holds data for /igcinfo/api/track/id
type igcFields struct {
	ID       bson.ObjectId `bson:"_id,omitempty" json:"-"`
	TrackID  int           `bson:"id" json:"-"`
	HDate    time.Time     `json:"H_date"`
	Pilot    string        `json:"pilot"`
	Glider   string        `json:"glider"`
	GliderID string        `json:"glider_id"`
	TrackLen float64       `json:"track_lenght"`
	TrackURL string        `json:"track_src_url"`
}

// the response type for POST /igcinfo/api/track
type resID struct {
	TrackID int `json:"id"`
}

// This function finds the amount of docs in db, uses it to return
//	an auto incremented ID
func getIncrementedID() (int, error) {
	var uniqueID int
	session, err := mgo.Dial(dbURL)
	if err != nil {
		return uniqueID, err
	}

	defer session.Close()

	uniqueID, err = session.DB(dbName).C(dbCollection).Count()

	if err != nil {
		return uniqueID, err
	}

	return uniqueID, err

}

// Takes a Unix time difference and returns string of ISO 8601
func calculateDuration(t time.Duration) string {
	startNewTime := time.Now()
	totalTime := int(startNewTime.Unix()) - int(startTime.Unix()) //int(t) / int(time.Second)

	remainderSeconds := totalTime % 60 // final seconds
	minutes := totalTime / 60
	remainderMinutes := minutes % 60 // final minutes
	hours := minutes / 60
	remainderHours := hours % 24 // final hours
	days := hours / 24
	remainderDays := days % 7 // final days
	months := days / 28
	remainderMonths := months % 12 // final months
	years := months / 12           // final years

	s := "P" + strconv.Itoa(years) + "Y" + strconv.Itoa(remainderMonths) + "M" + strconv.Itoa(remainderDays) + "D" + strconv.Itoa(remainderHours) + "H" + strconv.Itoa(remainderMinutes) + "M" + strconv.Itoa(remainderSeconds) + "S"
	return s
}

// Handles  igcinfo/api and outputs metadata in json
func metaHandler(w http.ResponseWriter, r *http.Request) {
	mt := meta{
		Uptime:  calculateDuration(time.Since(startTime)),
		Info:    "Service for Paragliding tracks.",
		Version: "v1"}

	m, err := json.MarshalIndent(&mt, "", "    ")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	fmt.Fprintf(w, string(m))
}

// After a POST, url is passed here to parse a track-object
func processURL(igcURL string, w http.ResponseWriter) (igcFields, error) {

	fields := igcFields{}
	track, err := igc.ParseLocation(igcURL)
	if err != nil {

		return fields, err
	}

	// Get unique ID
	var uniqueID int
	uniqueID, err = getIncrementedID()
	if err != nil {
		return fields, err
	}

	// Calculate total track distance
	totalDistance := 0.0
	for i := 0; i < len(track.Points)-1; i++ {
		totalDistance += track.Points[i].Distance(track.Points[i+1])
	}

	fields = igcFields{
		ID:       bson.NewObjectId(),
		HDate:    track.Date,
		TrackID:  uniqueID,
		Pilot:    track.Pilot,
		Glider:   track.GliderType,
		GliderID: track.GliderID,
		TrackLen: totalDistance,
		TrackURL: igcURL}

	// Response with ID as json and return the track
	response := resID{fields.TrackID}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	enc.Encode(&response)
	return fields, nil
}

// Add a track to the DB
func addToDB(fields igcFields) {

	session, err := mgo.Dial(dbURL)
	if err != nil {
		return
	}

	defer session.Close()

	err = session.DB(dbName).C(dbCollection).Insert(fields)

	if err != nil {
		return
	}
}

// List array of IDs in json
func displayIDs(w http.ResponseWriter) {
	session, err := mgo.Dial(dbURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	//session.SetMode(mgo.Monotonic, true)

	c := session.DB(dbName).C(dbCollection)

	item := igcFields{}

	find := c.Find(bson.M{})

	response := make([]int, 0)
	items := find.Iter()
	for items.Next(&item) {
		response = append(response, item.TrackID)
	}

	json.NewEncoder(w).Encode(&response)

}

// Check for POST and GET requests. POSTs URL
func inputHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		displayIDs(w)
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			return
		}

		icgURL := r.FormValue("url")
		fields, err := processURL(string(icgURL), w)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		addToDB(fields)

	}
}

//	In /igcinfo/api/track/ID we use ID to find a track i db
func getTrack(idOfTrack int) igcFields {
	session, err := mgo.Dial(dbURL)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	c := session.DB(dbName).C(dbCollection)
	response := igcFields{}

	c.Find(bson.M{"id": idOfTrack}).One(&response)

	return response

}

//	In /igcinfo/api/track/ID/FIELD we use ID to find a track i db
//	and FIELD to display that field
func getField(fields igcFields, field string, w http.ResponseWriter) {
	switch field {
	case "pilot":
		fmt.Fprintln(w, fields.Pilot)
	case "glider":
		fmt.Fprintln(w, fields.Glider)
	case "glider_id":
		fmt.Fprintln(w, fields.GliderID)
	case "track_length":
		fmt.Fprintln(w, fields.TrackLen)
	case "H_date":
		fmt.Fprintln(w, fields.HDate)
	case "track_src_url":
		fmt.Fprintln(w, fields.TrackURL)
	default:
		status := 404
		http.Error(w, http.StatusText(status), status)
	}

}

//	Handles the last two arguments for <ID> and <FIELD>
//
//
func argsHandler(w http.ResponseWriter, r *http.Request) {

	parts := strings.Split(r.URL.Path, "/") // array of url parts
	fields := igcFields{}

	if len(parts) > fieldArg+1 {
		status := 404
		http.Error(w, http.StatusText(status), status)
		return
	}

	if len(parts) > idArg {

		idOfTrack, _ := strconv.Atoi(parts[idArg])
		fields = getTrack(idOfTrack)

		if len(parts) < fieldArg+1 {
			enc := json.NewEncoder(w)
			enc.SetIndent("", "    ")
			enc.Encode(&fields)
		}
	}

	if len(parts) > fieldArg {

		field := string(parts[fieldArg])
		getField(fields, field, w)
	}
}

// Returns the amount of documents in the DB
func countHandler(w http.ResponseWriter, r *http.Request) {
	var docs int
	session, err := mgo.Dial(dbURL)
	if err != nil {
		return
	}

	defer session.Close()

	docs, err = session.DB(dbName).C(dbCollection).Count()

	if err != nil {
		return
	}

	fmt.Fprintln(w, docs)
}

// Deletes all documents in the DB collection
func deleteAll(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodDelete {
		session, err := mgo.Dial(dbURL)
		if err != nil {
			return
		}

		defer session.Close()

		_, err = session.DB(dbName).C(dbCollection).RemoveAll(bson.M{})
		if err != nil {
			return
		}
	}
}

// Main program
func main() {
	startTime = time.Now()
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	http.HandleFunc(root+"/api", metaHandler)
	http.HandleFunc(root+"/api/track", inputHandler)
	http.HandleFunc(root+"/api/track/", argsHandler)
	http.HandleFunc(root+"/admin/api/tracks_count", countHandler)
	http.HandleFunc(root+"/admin/api/tracks", deleteAll)
	http.ListenAndServe(":"+port, nil)
}
