/// 2>/dev/null ; gorun "$0" "$@" ; exit $?

package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/joho/godotenv"
)

type Episodes struct {
	AirDate       string `json:"air_date"`
	EpisodeNumber int    `json:"episode_number"`
	Crew          []struct {
		Job                string      `json:"job"`
		Department         string      `json:"department"`
		CreditID           string      `json:"credit_id"`
		Adult              bool        `json:"adult"`
		Gender             int         `json:"gender"`
		ID                 int         `json:"id"`
		KnownForDepartment string      `json:"known_for_department"`
		Name               string      `json:"name"`
		OriginalName       string      `json:"original_name"`
		Popularity         float64     `json:"popularity"`
		ProfilePath        interface{} `json:"profile_path"`
	} `json:"crew"`
	GuestStars []struct {
		Character          string  `json:"character"`
		CreditID           string  `json:"credit_id"`
		Order              int     `json:"order"`
		Adult              bool    `json:"adult"`
		Gender             int     `json:"gender"`
		ID                 int     `json:"id"`
		KnownForDepartment string  `json:"known_for_department"`
		Name               string  `json:"name"`
		OriginalName       string  `json:"original_name"`
		Popularity         float64 `json:"popularity"`
		ProfilePath        string  `json:"profile_path"`
	} `json:"guest_stars"`
	ID             int     `json:"id"`
	Name           string  `json:"name"`
	Overview       string  `json:"overview"`
	ProductionCode string  `json:"production_code"`
	SeasonNumber   int     `json:"season_number"`
	StillPath      string  `json:"still_path"`
	VoteAverage    float64 `json:"vote_average"`
	VoteCount      int     `json:"vote_count"`
}

type TVShow struct {
	_ID          string     `json:"_id"`
	AirDate      string     `json:"air_date"`
	Episodes     []Episodes `json:"episodes"`
	Name         string     `json:"name"`
	Overview     string     `json:"overview"`
	ID           int        `json:"id"`
	PosterPath   string     `json:"poster_path"`
	SeasonNumber int        `json:"season_number"`
}

type EpisodeNames struct {
	Name            string `json:"name"`
	CurrentFilename string `json:"currentFilename"`
	NewFilename     string `json:"newFilename"`
}

var BuildVersion string = "0.1.0"

// get TMDB data for show
func showData(id string, season string) TVShow {
	// get the show data
	key := os.Getenv("TMDB_KEY")
	if key == "" {
		log.Fatalln("Please provide a TMDB API key")
	}

	url := "https://api.themoviedb.org/3/tv/" + id + "/season/" + season + "?api_key=" + key
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	// read the response
	rest, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// unmarshal the response
	var show TVShow
	err = json.Unmarshal(rest, &show)
	if err != nil {
		log.Fatal(err)
	}

	return show
}

// add zero to single digit season/episode numbers
func addZero(num int) string {
	if num < 10 {
		return "0" + strconv.Itoa(num)
	} else {
		return strconv.Itoa(num)
	}
}

// get formatted episode names
func episodeNames(data TVShow, showName string) []EpisodeNames {
	var episodeName []EpisodeNames

	for i := 0; i < len(data.Episodes); i++ {
		episode := data.Episodes[i]
		// clean names with parts
		partPat := regexp.MustCompile(`Part\s([0-9]+):\s`)
		name := partPat.ReplaceAllString(episode.Name, `Part $1 - `)
		// remove symbols and replace spaces with dashes
		symbolPat := regexp.MustCompile(`[,:]`)
		spacePat := regexp.MustCompile(`\s`)
		fmtPart := partPat.ReplaceAllString(episode.Name, `-Part_$1-`)
		fmtSymbols := symbolPat.ReplaceAllString(fmtPart, ``)
		fmtName := spacePat.ReplaceAllString(fmtSymbols, `_`)
		// generate new filename
		file := showName + "-S" + addZero(data.SeasonNumber) + "-E" + addZero(episode.EpisodeNumber) + "-"

		episodeName = append(episodeName, EpisodeNames{
			Name:            name,
			CurrentFilename: file + ".mkv",
			NewFilename:     file + fmtName + ".mkv",
		})
	}

	return episodeName
}

// rename file with episode name
func renameFile(episode EpisodeNames) {
	log.Println("[Rename]:" + episode.CurrentFilename + " -> " + episode.NewFilename)

	// rename the file
	err := os.Rename(episode.CurrentFilename, episode.NewFilename)
	if err != nil {
		log.Fatal(err)
	}
}

// add media title to filename
func mediaTitle(episode EpisodeNames) {
	log.Println("[mkvproedit]: title = " + episode.Name)

	// get current direcotry
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

    // add file metadata
	cmd := "mkvproedit" + path + "/" + episode.NewFilename + "-e info -s title=" + episode.Name
	err = exec.Command(cmd).Run()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// get arguments
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

    // parse arguments
	showNamePtr := flag.String("showName", "", "Show name")
	showIDPtr := flag.String("showID", "", "TMDB Show ID")
	seasonPtr := flag.String("season", "", "Show seaon")

	flag.Parse()

	// get the show data
	show := showData(*showIDPtr, *seasonPtr)

	switch true {
	case *showNamePtr == "":
		log.Fatalln("Please provide a show name")
	case *showIDPtr == "":
		log.Fatalln("Please provide a show ID")
	case *seasonPtr == "":
		log.Fatalln("Please provide a season number")
	default:
		newData := episodeNames(show, *showNamePtr)

		for i := 0; i < len(newData); i++ {
			renameFile(newData[i])
			mediaTitle(newData[i])
		}
	}
}
