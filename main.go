package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	groupieData       GroupieData
	dateRelationsData map[int]any
)

type GroupieData struct {
	Artists []struct {
		ID      int      `json:"id"`
		Image   string   `json:"image"`
		Name    string   `json:"name"`
		Members []string `json:"members"`
		Cdate   int      `json:"creationDate"`
		Falbum  string   `json:"firstAlbum"`
	}
	Location struct {
		Index []struct {
			ID           int                 `json:"id"`
			Datarelation map[string][]string `json:"datesLocations"`
		} `json:"index"`
	}
}

type Details struct {
	Image     string
	Name      string
	Cdate     int
	Falbum    string
	Relations map[string][]string
}


func FetchData() {
	urls := map[string]any{
		"https://groupietrackers.herokuapp.com/api/artists":  &groupieData.Artists,
		"https://groupietrackers.herokuapp.com/api/relation": &groupieData.Location,
	}

	for url, dectag := range urls {
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal("Error fetching data:", err)
		}
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(dectag); err != nil {
			log.Fatal("Error decoding data:", err)
		}
		
	}
	for i := range groupieData.Artists {
		groupieData.Artists[i].Name = strings.TrimSpace(strings.ToUpper(groupieData.Artists[i].Name))
	}

	dateRelationsData = make(map[int]any)

	for _, RelApi := range groupieData.Location.Index {
		for _, ArtApi := range groupieData.Artists {
			if RelApi.ID == ArtApi.ID {
				dateRelationsData[RelApi.ID] = Details{
					Image:     ArtApi.Image,
					Name:      ArtApi.Name,
					Cdate:     ArtApi.Cdate,
					Falbum:    ArtApi.Falbum,
					Relations: RelApi.Datarelation,
				}
				break
			}
		}
	}

}

func Hundler(w http.ResponseWriter, r *http.Request) {
	TmplStatus, _ := template.ParseFiles("template/status.html")
	if TmplStatus == nil { http.Error(w, "500 Internal Server Error", 500); return }

	path := r.URL.Path
	switch {

	case path == "/" && r.Method == "GET":
		tmpl, _ := template.ParseFiles("template/home.html")
		if tmpl == nil { w.WriteHeader(500) ;TmplStatus.Execute(w, "500 Internal Server Error"); return }

		err := tmpl.Execute(w, groupieData.Artists)
		if err != nil { w.WriteHeader(500); TmplStatus.Execute(w, "500 Internal Server Error"); return}

	case strings.HasPrefix(r.URL.Path, "/locations/") && r.Method == "GET":

		id := r.URL.Path[len("/locations/"):]
		if len(id) == 0 { w.WriteHeader(404); TmplStatus.Execute(w, "404 Not Found"); return }

		artistID, err := strconv.Atoi(id)
		if err != nil || artistID <= 0 || artistID > 52 { w.WriteHeader(400); TmplStatus.Execute(w, "400 Status Bad Request"); return }

		tmpl, err := template.ParseFiles("template/details.html")
		if tmpl == nil { w.WriteHeader(500); TmplStatus.Execute(w, "500 Internal Server Error") ;return }

		err = tmpl.Execute(w, dateRelationsData[artistID])
		if err != nil { w.WriteHeader(500); TmplStatus.Execute(w, "500 Internal Server Error"); return }

	default:
		w.WriteHeader(404); TmplStatus.Execute(w, "404 Not Found"); return

	}
}

func main() {
	args := os.Args[1:]
	if len(args) != 0 { return }

	FetchData()
	http.HandleFunc("/", Hundler)

	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
