/// 2>/dev/null ; gorun "$0" "$@" ; exit $?

package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
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

var BuildVersion string = "1.0.0"

// get TMDB data for show
func showData(id string, season string) TVShow {
	// get TMDB API key
	envPath := os.Getenv("GOPATH") + "/.env.showrunner"
	err := godotenv.Load(envPath)
	if err != nil {
		println(color.RedString("[Error (godotenv.Load)]:"))
		log.Fatal(err)
	}

	// get the show data
	key := os.Getenv("TMDB_KEY")
	if key == "" {
		println(color.RedString("[Error (os.Getenv)]:"))
		log.Fatalln("Please provide a TMDB API key")
	}

	url := "https://api.themoviedb.org/3/tv/" + id + "/season/" + season + "?api_key=" + key
	resp, err := http.Get(url)
	if err != nil {
		println(color.RedString("[Error (http.Get)]:"))
		log.Fatal(err)
	}

	// read the response
	rest, err := io.ReadAll(resp.Body)
	if err != nil {
		println(color.RedString("[Error (io.ReadAll)]:"))
		log.Fatal(err)
	}

	// unmarshal the response
	var show TVShow
	err = json.Unmarshal(rest, &show)
	if err != nil {
		println(color.RedString("[Error (json.Unmarshal)]:"))
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
		// setup match patterns
		partPat := regexp.MustCompile(`Part\s([0-9]+):\s`)
		spacePat := regexp.MustCompile(`\s+`)
		dashPat := regexp.MustCompile(`\s-\s`)
		// clean names with parts
		name := partPat.ReplaceAllString(episode.Name, `Part $1 - `)
		// remove symbols and replace spaces with dashes
		fmtPart := partPat.ReplaceAllString(episode.Name, `-Part_$1-`)
		fmtSymbols := regexp.MustCompile(`[,:!@#$%^&*()_+{}|\[\]~;’‘'”“"<>?/]`).ReplaceAllString(fmtPart, ``)
		fmtName4 := regexp.MustCompile(`\.{3}`).ReplaceAllString(fmtSymbols, ``)
		fmtName3 := regexp.MustCompile(`\s--\s`).ReplaceAllString(fmtName4, `-`)
		fmtName2 := regexp.MustCompile(`\.\s`).ReplaceAllString(fmtName3, `_`)
		fmtName1 := dashPat.ReplaceAllString(fmtName2, `-`)
		fmtName := spacePat.ReplaceAllString(fmtName1, `_`)
		// generate new filename
		cleanName := spacePat.ReplaceAllString(showName, `_`)
		cleanerName := dashPat.ReplaceAllString(cleanName, `-`)
		file := cleanerName + "-S" + addZero(data.SeasonNumber) + "E" + addZero(episode.EpisodeNumber)

		episodeName = append(episodeName, EpisodeNames{
			Name:            name,
			CurrentFilename: file + ".mkv",
			NewFilename:     file + "-" + fmtName + ".mkv",
		})
	}

	return episodeName
}

// rename file with episode name
func renameFile(episode EpisodeNames) {
	log.Println(color.YellowString("[Rename]: ") + episode.CurrentFilename + " -> " + episode.NewFilename)

	// rename the file
	err := os.Rename(episode.CurrentFilename, episode.NewFilename)
	if err != nil {
		println(color.RedString("[Error (os.Rename)]:"))
		log.Fatal(err)
	}
}

// add media title to filename
func setMediaTitle(episode EpisodeNames) {
	log.Println(color.YellowString("[mkvpropedit]: ") + "title = " + episode.Name)

	// get current direcotry
	path, err := os.Getwd()
	if err != nil {
		println(color.RedString("[Error (os.Getwd)]:"))
		log.Fatal(err)
	}

	// add file metadata
	filepath := path + "/" + episode.CurrentFilename
	title := "title=" + "\"" + episode.Name + "\""

	cmd := exec.Command("mkvpropedit", filepath, "-e", "info", "-s", title)

	out, err := cmd.CombinedOutput()
	if err != nil {
		println(color.RedString("[Error (cmd.CombinedOutput)]:"))
		log.Fatal(err)
	}

	log.Println(color.YellowString("[mkvpropedit]: "), string(out))
}

func addMetadata(showName string, season string, showID string) {
	// get the show data
	show := showData(showID, season)
	newData := episodeNames(show, showName)

	// add metadata to files
	for i := 0; i < len(newData); i++ {
		setMediaTitle(newData[i])
		renameFile(newData[i])
	}
}

func main() {
	var showName string
	var season string
	var showID string

	// versioning
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "print app version",
	}

	// help
	cli.AppHelpTemplate = `NAME:
	{{.Name}} - {{.Usage}}

VERSION:
	{{.Version}}

USAGE:
	{{.HelpName}} [optional options]

OPTIONS:
{{range .VisibleFlags}}	{{.}}{{ "\n" }}{{end}}	
	`
	cli.HelpFlag = &cli.BoolFlag{
		Name:    "help",
		Aliases: []string{"h"},
		Usage:   "show help",
	}

	// execute app
	app := &cli.App{
		Name:    "showrunner",
		Usage:   "Add episode metadata to shows from TMDB",
		Version: BuildVersion,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "showName",
				Aliases:     []string{"n"},
				Usage:       "show name (spaces are allowed)",
				Destination: &showName,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "season",
				Aliases:     []string{"s"},
				Usage:       "show seaon",
				Destination: &season,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "showID",
				Aliases:     []string{"i"},
				Usage:       "TMDB show id",
				Destination: &showID,
				Required:    true,
			},
		},
		Action: func(*cli.Context) error {
			addMetadata(showName, season, showID)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
