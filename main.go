package main

import (
	"log"
	"net/http"
	"os"
	"fmt"
	"github.com/marni/goigc"
	"encoding/json"
	//"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"time"
	//"strconv"


	//"github.com/gin-gonic/gin"
	//_ "github.com/heroku/x/hmetrics/onload"
)

const root = "/igcinfo"


type Meta struct {
	Uptime  string `json:"uptime"`
	Info    string `json:"info"`
	Version string `json:"version"`
}



type Fields struct {
	Id bson.ObjectId 	`bson:"_id,omitempty"`
	HDate    time.Time `json:"H_date"`
	Pilot    string    `json:"pilot"`
	Glider   string    `json:"glider"`
	GliderID string    `json:"glider_id"`
	TrackLen float64   `json:"track_lenght"`
}





func metaHandler(w http.ResponseWriter, r * http.Request){
	meta := Meta{
	Uptime:  "time",//calculateDuration(time.Since(startTime)),
	Info:    "Service for IGC tracks.",
	Version: "v1"}

	m, err := json.MarshalIndent(&meta, "", "    ")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	fmt.Fprintf(w, string(m))
}




func processURL(igcURL string, w http.ResponseWriter){

	track, err := igc.ParseLocation(igcURL)
	if err != nil {
		status := 400
		http.Error(w, http.StatusText(status), status)
		return
	}
	
	fields := Fields{
				HDate: track.Date,
				Pilot: track.Pilot,
				Glider: track.GliderType,
				GliderID: track.GliderID,
				TrackLen: 0}
				
	fmt.Fprintln(w, fields)
}






func inputHandler(w http.ResponseWriter, r * http.Request){
	switch r.Method {
		case http.MethodGet:						
			//json.NewEncoder(w).Encode()
		case http.MethodPost:					
			if err := r.ParseForm(); err != nil {
				return
			}

			icgURL := r.FormValue("icg")	
			processURL(string(icgURL), w)
					
	}
}


func argsHandler(w http.ResponseWriter, r * http.Request){
	fmt.Fprintln(w, "ARGS")
}



func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	http.HandleFunc(root + "/api", metaHandler)
	http.HandleFunc(root + "/api/igc", inputHandler)
	http.HandleFunc(root + "/api/igc/", argsHandler)
	http.ListenAndServe(":" +port, nil);
}
