package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "igtagsnarf"
	app.Usage = "download assets from instagram that match a tag"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "client",
			Value: "",
			Usage: "instagram client ID",
		},
		cli.StringFlag{
			Name:  "tag",
			Value: "",
			Usage: "tag to search by",
		},
	}

	app.Action = func(c *cli.Context) {
		tag, clientID := c.String("tag"), c.String("client")
		snarf(tag, clientID)
	}

	app.Run(os.Args)
}

// download a list of items that match a tag
func snarf(tag, clientID string) error {
	url := fmt.Sprintf("https://api.instagram.com/v1/tags/%s/media/recent?count=500&client_id=%s", tag, clientID)

	fmt.Println("GET %s", url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("error fetching list from instagram: ", err)
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		fmt.Println("error reading the response from instagram:", err)
		return err
	}

	s := struct {
		Assets []Asset `json:"data"`
	}{}

	if err := json.Unmarshal(body, &s); err != nil {
		fmt.Println("error unmarshaling response json:", err)
		fmt.Printf("%s", string(body))
		return err
	}

	work := make(chan Asset, len(s.Assets))
	done := make(chan error, 5)

	// launch five downloaders
	for i := 0; i < 5; i++ {
		go download(work, done, "files")
	}

	// send items to workers
	for _, a := range s.Assets {
		work <- a
	}
	close(work)

	for i := 0; i < 5; i++ {
		err := <-done
		if err != nil {
			// handle
		}
	}

	return nil
}

// trigger a download of an asset
func download(assets <-chan Asset, done chan<- error, out string) error {
	for a := range assets {
		path := fmt.Sprintf("./%s/%s-%s.jpg", out, a.User.Username, a.ID)

		imgUrl := a.Images["standard_resolution"].URL
		fmt.Println("downloading", imgUrl)
		resp, err := http.Get(imgUrl)
		if err != nil {
			fmt.Println("error: fetch img failed:", err)
			return err
		}

		body, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			fmt.Println("error: read response failed:", err)
			return err
		}

		if err := ioutil.WriteFile(path, body, 0644); err != nil {
			fmt.Println("error: could not write file:", err)
			return err
		}
	}

	done <- nil
	return nil
}
