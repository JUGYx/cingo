package main

import (
    "errors"
    "fmt"
    "os"
    "strings"
    "slices"
    "strconv"
    "os/exec"
    "github.com/JUGYx/cingo/api"
    "github.com/charmbracelet/huh"
)

func checkErr(err error) {
    if err != nil {
        if err == huh.ErrUserAborted {
            fmt.Println("Bye...")
            os.Exit(0)
        } else {
            fmt.Fprintf(os.Stderr, "[Error] %s\n", err)
            os.Exit(1)
        }
    }
}

func prepareList(result []interface{}) map[string]string {
    list := make(map[string]string, len(result))

    for _, item := range result {
        itemM := item.(map[string]interface{})
        key := fmt.Sprintf("%s | %s | %s", itemM["en_title"].(string),
        itemM["mDate"].(string), itemM["stars"].(string))
        list[key] = itemM["nb"].(string)
    }

    return list
}

func pickSeason(seasonsView []string) string { 
    var chosenSeason string

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
            Options(huh.NewOptions(seasonsView...)...).
            Title("Pick a season").
            Value(&chosenSeason),
        ),
    )

    err := form.Run()
    checkErr(err)

    return chosenSeason
}

func pickReso(id string) (string, string, string, error) {
    //TODO: Ask for subtitles if not present
    vids, sub, err := api.GetMedia(id)
    checkErr(err)

    arSubtitles := sub["arTranslationFilePath"].(string)
    enSubtitles := sub["enTranslationFilePath"].(string)

    // For some reason cinemana provides `defaultImages/loading.gif`
    // when no subtitles are present.
    // Therefore to be sure I check *File instead of *Path:
    if sub["arTranslationFile"] == "" {
        arSubtitles = ""
    }
    if sub["enTranslationFile"] == "" {
        enSubtitles = ""
    }

    resolutionView := make([]string, len(vids)+1)
    resolutionView[0] = "<<<"
    for i, r := range vids {
        rM := r.(map[string]interface{})
        resolutionView[i+1] = rM["resolution"].(string)
    }

    var chosenRes string

    resoForm := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
            Options(huh.NewOptions(resolutionView...)...).
            Title("Pick a resolution").
            Value(&chosenRes),
        ),
    )

    err = resoForm.Run()
    checkErr(err)

    if chosenRes == "<<<" {
        return "", "", "", errors.New("<<<")
    }

    var vidUrl string

    for _, r := range vids {
        rM := r.(map[string]interface{})
        if chosenRes == rM["resolution"].(string) {
            vidUrl = rM["videoUrl"].(string)
        }
    }

    return arSubtitles, enSubtitles, vidUrl, nil
}

func play(ar string, en string, vid string) {
    var para []string  

    if ar != "" {
        para = append(para, "--sub-file="+ar)
    }
    if en != "" {
        para = append(para, "--sub-file="+en)
    }
    para = append(para, vid)

    mpvCmd := exec.Command("mpv", para...)
    _, err := mpvCmd.Output()
    if err != nil {
        fmt.Fprintf(os.Stderr, "[Error] Make sure mpv is installed: %s\n", err)
        os.Exit(-1)
    }
}

func queryForm() (string, string) {
    var mediaType string
    var query string

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
            Options(huh.NewOptions("Movies", "Shows")...).
            Title("Pick a genre").
            Value(&mediaType),
        ),

        huh.NewGroup(
            huh.NewInput().
            Value(&query).
            Title("Media title?").
            Validate(func(s string) error {
                s = strings.Trim(s, " ")

                if s == "" {
                    return errors.New("Provide a title name!.")
                }

                return nil
            }).
            Description("Examples: one piece, jojo, mad god"),
        ),
    )

    err := form.Run()
    checkErr(err)

    query = strings.Trim(query, " ")

    return mediaType, query
}

func main() {
    noteForm := huh.NewForm(
        huh.NewGroup(huh.NewNote().
        Title("Cinemana").
        Description("* Client requires mpv to work (mpv.io)\n* Ctrl-c to quit")),
    )

    err := noteForm.Run()
    checkErr(err)

    queryFor: for {
        mediaType, query := queryForm()

        pageNumber := 0

        // No, I won't ask the user whether they want movies or series and use that
        if mediaType == "Movies" {
            mediaType = api.MOVIES
        } else {
            mediaType = api.SERIES
        }

        var chosenMedia string
        var mediaID string
        var listView []string

        for {
            result, err := api.Query(query, pageNumber, mediaType)
            checkErr(err)

            list := prepareList(result)

            listView = make([]string, len(list)+2)
            i := 0

            listView[i] = "<<<"

            i++

            for k := range list {
                listView[i] = k
                i++
            }
            listView[i] = "..."

            listForm := huh.NewForm(
                huh.NewGroup(
                    huh.NewSelect[string]().
                    Options(huh.NewOptions(listView...)...).
                    Title("Pick your media").
                    Validate(func(s string) error {
                        if s == "<<<" && pageNumber < 1 {
                            return errors.New("No pages further back")
                        }

                        return nil
                    }).
                    Value(&chosenMedia),
                ),
            )

            err = listForm.Run()
            checkErr(err)

            if chosenMedia == "..." {
                pageNumber++
                continue
            }

            if chosenMedia == "<<<" {
                pageNumber--
                continue
            }

            mediaID = list[chosenMedia]

            break
        }

        if mediaType == api.MOVIES {
            arSubtitles, enSubtitles, vidUrl, err := pickReso(mediaID)

            if err != nil {
                continue queryFor
            }

            play(arSubtitles, enSubtitles, vidUrl)
        } else {
            seasonFor: for {
                result, err := api.GetSeasons(mediaID)
                checkErr(err)

                seasons := make(map[string][]string)

                for _, item := range result {
                    itemM := item.(map[string]interface{})
                    seasons[itemM["season"].(string)] = append(seasons[itemM["season"].(string)],
                    itemM["nb"].(string))
                }

                seasonsView := make([]string, len(seasons)+1)
                i := 0

                seasonsView[i] = "<<<"

                i++

                for k := range seasons {
                    seasonsView[i] = k
                    i++
                }
                slices.Sort(seasonsView[1:])
                chosenSeason := pickSeason(seasonsView)

                if chosenSeason == "<<<" {
                    continue queryFor
                }

                var chosenEpisode string

                for {
                    episodeView := make([]string, len(seasons[chosenSeason])+1)
                    episodeView[0] = "<<<"
                    for i := range seasons[chosenSeason] {
                        episodeView[i+1] = strconv.Itoa(i+1)
                    }

                    form := huh.NewForm(
                        huh.NewGroup(
                            huh.NewSelect[string]().
                            Options(huh.NewOptions(episodeView...)...).
                            Title("Pick an episode").
                            Value(&chosenEpisode),
                        ),
                    )

                    err := form.Run()
                    checkErr(err)

                    if chosenEpisode == "<<<" {
                        continue seasonFor
                    }

                    chosenEpisode, err := strconv.Atoi(chosenEpisode)
                    if err != nil {
                        panic(err)
                    }

                    arSubtitles, enSubtitles, vidUrl, err := pickReso(seasons[chosenSeason][chosenEpisode-1])

                    if err != nil {
                        continue
                    }

                    play(arSubtitles, enSubtitles, vidUrl)
                }

                break
            }
        }
    }
}

