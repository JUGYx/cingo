package api

import (
    "net/http"
    "encoding/json"
    "fmt"
    "io"
    "net/url"
)

const API_URL = "https://cinemana.shabakaty.com/api/android"

func get(gurl string) (*http.Response, error) {
    client := &http.Client{
        CheckRedirect: nil,
    }

    req, err := http.NewRequest("GET", gurl, nil)
    if err != nil {
        panic(err)
    }

    req.Header.Add("User-Agent",
    "Mozilla/5.0 (X11; Linux x86_64; rv:124.0) Gecko/20100101 Firefox/124.0")

    req.Header.Add("Accept", "application/json, text/plain, */*")
    req.Header.Add("Accept-Language", "en-US,en;q=0.5")
    req.Header.Add("Sec-Fetch-Dest", "empty")
    req.Header.Add("Sec-Fetch-Mode", "no-cors")
    req.Header.Add("Sec-Fetch-Site", "same-origin")
    req.Header.Add("Connection", "keep-alive")

    resp, err := client.Do(req)

    if err == nil && resp.StatusCode > 299 {
        return nil, fmt.Errorf("Get request to `%s` returned status %d", resp.StatusCode)
    }

    return resp, err
}

func Query(query string, page_number int, genre string) (*[]interface{}, error) {
    query = url.QueryEscape(query)
    aurl := fmt.Sprintf("%s/AdvancedSearch?level=0&videoTitle=%s&staffTitle=%s&page=%d&type=%s&", API_URL, query, query, page_number, genre)

    resp, err := get(aurl)
    if err != nil {
        return nil, err
    }

    defer resp.Body.Close()

    var result []interface{}

    bytes, err := io.ReadAll(resp.Body)

    if err != nil {
        return nil, err
    }

    err = json.Unmarshal(bytes, &result)
    if err != nil {
        return nil, err
    }

    return &result, nil
}

func GetSeasons(id string) (*[]interface{}, error) {
    aurl := fmt.Sprintf("%s/videoSeason/id/%s", API_URL, id)

    resp, err := get(aurl)
    if err != nil {
        return nil, err
    }

    defer resp.Body.Close()

    var result []interface{}

    bytes, err := io.ReadAll(resp.Body)

    if err != nil {
        return nil, err
    }

    err = json.Unmarshal(bytes, &result)
    if err != nil {
        return nil, err
    }

    return &result, nil
}

func GetMedia(id string) (*[]interface{}, *map[string]interface{}, error) {
    vid_aurl := fmt.Sprintf("%s/transcoddedFiles/id/%s", API_URL, id)
    sub_aurl := fmt.Sprintf("%s/allVideoInfo/id/%s", API_URL, id)

    vid_resp, err := get(vid_aurl)
    if err != nil {
        return nil, nil, err
    }
    defer vid_resp.Body.Close()

    sub_resp, err := get(sub_aurl)
    if err != nil {
        return nil, nil, err
    }
    defer sub_resp.Body.Close()

    var vid_result []interface{}

    vid_bytes, err := io.ReadAll(vid_resp.Body)

    if err != nil {
        return nil, nil, err
    }

    err = json.Unmarshal(vid_bytes, &vid_result)
    if err != nil {
        return nil, nil, err
    }

    var sub_result map[string]interface{}

    sub_bytes, err := io.ReadAll(sub_resp.Body)

    if err != nil {
        return &vid_result, nil, err
    }

    err = json.Unmarshal(sub_bytes, &sub_result)
    if err != nil {
        return &vid_result, nil, err
    }

    return &vid_result, &sub_result, nil
}
