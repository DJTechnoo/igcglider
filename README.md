### IGCFORLIFE

## About

This is a REST API for IGC files. 
By posting a URL to an ICG file, it can
extract the most important fields and parse them
into json.
The storage is persistent - They will be stored in a Mongo Database.

### Deployment
The URL to the app can be found here: `https://igcforlife.herokuapp.com`.

### Usage
Navigate to `/igcinfo/api` to GET meta about app.
Navigate to `/igcinfo/api/track` to GET all track IDs.
Navigate to `/igcinfo/api/track/<id>` to GET meta about that track.
Navigate to `/igcinfo/api/track/<id>/<field>` to GET field from that track.
At `/igcinfo/api/track` use POST request with form `"url"` to add igc file.
Everything is output in json except the `<field>` request.




