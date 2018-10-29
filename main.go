package main

import (
	"log"
	"net/http"
	"os"
	"fmt"
	"github.com/marni/goigc"
	"encoding/json"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"time"
	"strings"
	"strconv"

)

const root = "/igcinfo"
const idArg = 4            // URL index for ID
const fieldArg = 5         // URL index for FIELD

var	Url = "mongodb://igcuser:igc4life@ds141783.mlab.com:41783/igc"
var	Name = "igc"
var Collection = "igcstruct"
var startTime time.Time


type Meta struct {
	Uptime  string `json:"uptime"`
	Info    string `json:"info"`
	Version string `json:"version"`
}



type Fields struct {
	Id bson.ObjectId 	`bson:"_id,omitempty" json:"-"`
	TrackID int			`bson:"id" json:"-"`
	HDate    time.Time `json:"H_date"`
	Pilot    string    `json:"pilot"`
	Glider   string    `json:"glider"`
	GliderID string    `json:"glider_id"`
	TrackLen float64   `json:"track_lenght"`
	TrackURL string		`json:"track_src_url"`
}

type ID struct {
	TrackID int			`json:"id"`
}


func getIncrementedID()(int, error){
	var uniqueID int
	session, err := mgo.Dial(Url)
	if err != nil {
		return uniqueID, err
	}
	
	defer session.Close()
	
	uniqueID, err = session.DB(Name).C(Collection).Count()
	
	if err != nil {
		return uniqueID, err
	}
	
	return uniqueID, err

}


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





func metaHandler(w http.ResponseWriter, r * http.Request){
	meta := Meta{
	Uptime:  calculateDuration(time.Since(startTime)),
	Info:    "Service for Paragliding tracks.",
	Version: "v1"}

	m, err := json.MarshalIndent(&meta, "", "    ")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	fmt.Fprintf(w, string(m))
}




func processURL(igcURL string, w http.ResponseWriter)(Fields, error){

	fields := Fields{}
	track, err := igc.ParseLocation(igcURL)
	if err != nil {

		return fields, err
	}
	
	var uniqueID int
	uniqueID, err = getIncrementedID()
	if err != nil {
		return fields, err
	}
	
	totalDistance := 0.0
	for i := 0; i < len(track.Points)-1; i++ {
    	totalDistance += track.Points[i].Distance(track.Points[i+1])
	}
	
	fields = Fields{
				Id: bson.NewObjectId(),
				HDate: track.Date,
				TrackID: uniqueID,
				Pilot: track.Pilot,
				Glider: track.GliderType,
				GliderID: track.GliderID,
				TrackLen: totalDistance,
				TrackURL: igcURL}
				
	response := ID{fields.TrackID}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	enc.Encode(&response)
	return fields, err
}


func addToDB(fields Fields){

	session, err := mgo.Dial(Url)
	if err != nil {
		return
	}
	
	defer session.Close()
	
	err = session.DB(Name).C(Collection).Insert(fields)
	
	if err != nil {
		return
	}
}


func displayIDs(w http.ResponseWriter){
	session, err := mgo.Dial(Url)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	
	//session.SetMode(mgo.Monotonic, true)
	
	c := session.DB(Name).C(Collection)
	
	item := Fields{}
	
	find := c.Find(bson.M{})
	
	response := make([]int, 0)
	items := find.Iter()
	for items.Next(&item) {
		response = append(response, item.TrackID)
	}
	
	json.NewEncoder(w).Encode(&response)

}



func inputHandler(w http.ResponseWriter, r * http.Request){
	switch r.Method {
		case http.MethodGet:						
			displayIDs(w)
		case http.MethodPost:					
			if err := r.ParseForm(); err != nil {
				return
			}

			icgURL := r.FormValue("url")	
			fields, _ := processURL(string(icgURL), w)
			addToDB(fields);
					
	}
}


func getTrack(idOfTrack int)(Fields){
	session, err := mgo.Dial(Url)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	
	c := session.DB(Name).C(Collection)
	response := Fields{}
	
	c.Find(bson.M{"id": idOfTrack}).One(&response)
	
	
	return response
	
}


func getField(fields Fields, field string, w http.ResponseWriter){
	switch field {
		case "pilot":			fmt.Fprintln(w, fields.Pilot)
		case "glider":			fmt.Fprintln(w, fields.Glider)
		case "glider_id":		fmt.Fprintln(w, fields.GliderID)
		case "track_length":	fmt.Fprintln(w, fields.TrackLen)
		case "H_date":			fmt.Fprintln(w, fields.HDate)
		case "track_src_url":	fmt.Fprintln(w, fields.TrackURL)
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
	fields := Fields{}

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


func countHandler(w http.ResponseWriter, r * http.Request){
	var docs int
	session, err := mgo.Dial(Url)
	if err != nil {
		return 
	}
	
	defer session.Close()
	
	docs, err = session.DB(Name).C(Collection).Count()
	
	if err != nil {
		return 
	}
	
	fmt.Fprintln(w, docs)
}


func deleteAll(w http.ResponseWriter, r * http.Request){

	
	if r.Method == http.MethodDelete {
		session, err := mgo.Dial(Url)
		if err != nil {
			return 
		}

		defer session.Close()

		_, err = session.DB(Name).C(Collection).RemoveAll(bson.M{})
		if err != nil {
			return 
		}
	}
}



func main() {
	startTime = time.Now()
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	http.HandleFunc(root + "/api", metaHandler)
	http.HandleFunc(root + "/api/track", inputHandler)
	http.HandleFunc(root + "/api/track/", argsHandler)
	http.HandleFunc(root + "/admin/api/tracks_count", countHandler)
	http.HandleFunc(root + "/admin/api/tracks", deleteAll)
	http.ListenAndServe(":" +port, nil);
}
