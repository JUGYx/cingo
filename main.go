package main

import (
    "fmt"
    "github.com/JUGYx/cingo/api"
)

func main() {
    query_result, err := api.Query("jojo", 0, "series")
    if err != nil {
        fmt.Println(err);
        return
    }
    
    first := (*query_result)[0].(map[string]interface{})

    s_result, err := api.GetSeasons(first["nb"].(string))
    if err != nil {
        fmt.Println(err);
        return
    }

    // fmt.Println(first)
   
    result := (*s_result)[0].(map[string]interface{})
    id := result["nb"].(string)

    vid, sub, err := api.GetMedia(id)
    if err != nil {
        fmt.Println(err);
        return
    }

    fmt.Println(vid)
    fmt.Println(sub)
}
