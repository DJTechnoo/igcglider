### IGCFORLIFE

## About

This is a REST API for IGC files. 
By posting a URL to an ICG file, it can
extract the most important fields and parse them
into json.
The storage is persistent - They will be stored in a Mongo Database.

### Deployment
The URL to the app can be found here: `https://igcforlife.herokuapp.com`.

## Usage
### Track
Navigate to `/paragliding/api` to GET meta about app.
Navigate to `/paragliding/api/track` to GET all track IDs.
Navigate to `/paragliding/api/track/<id>` to GET meta about that track.
Navigate to `/paragliding/api/track/<id>/<field>` to GET field from that track.
At `/paragliding/api/track` use POST request with form `"url"` to add igc file.
Everything is output in json except the `<field>` request.

### Ticker
Navigate to `/paragliding/api/ticker` to GET latest added timestamp, and up to
five ids, first of which being the oldest, and the last being the latest on that page.
Navigate to `/paragliding/api/ticker/<timestamp>` to get up to five tracks that are
later than provided timestamp.




