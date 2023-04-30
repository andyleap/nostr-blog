package main

import (
	"context"
	"embed"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/andyleap/nostr/client"
	"github.com/andyleap/nostr/proto"
	"github.com/andyleap/nostr/proto/comm"
)

//go:embed templates
var assets embed.FS

var templates = template.Must(template.New("").Funcs(templateFuncs).ParseFS(assets, "templates/*.html"))

var templateFuncs = template.FuncMap{
	"parseTime": func(t int64) time.Time {
		return time.Unix(t, 0)
	},
}

func main() {
	var relayURL = os.Getenv("RELAY_URL")
	var pubkey = os.Getenv("PUB_KEY")
	if relayURL == "" || pubkey == "" {
		log.Fatal("RELAY_URL or PUB_KEY is empty")
	}

	r, err := client.Dial(context.Background(), relayURL)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		postSub, err := r.Subscribe(req.Context(), &comm.Filter{
			Kinds:   []int64{0, 1},
			Authors: []string{pubkey},
		})
		if err != nil {
			log.Println(err)
			return
		}
		posts := []*proto.Event{}
		done := false
		userInfo := map[string]interface{}{}
		for !done {
			select {
			case p, ok := <-postSub.Events():
				if !ok {
					done = true
					continue
				}
				if p.Kind == 0 {
					json.Unmarshal([]byte(p.Content), &userInfo)
					continue
				}
				posts = append(posts, p)
			case <-postSub.Backfilling():
				postSub.Close()
			}
		}

		//reverse order of posts
		for i := len(posts)/2 - 1; i >= 0; i-- {
			opp := len(posts) - 1 - i
			posts[i], posts[opp] = posts[opp], posts[i]
		}

		err = templates.ExecuteTemplate(rw, "index.html", map[string]interface{}{
			"user":  userInfo,
			"posts": posts,
		})
		if err != nil {
			log.Println(err)
		}
	})
	http.ListenAndServe(":8080", nil)
}

type PostWatcher struct {
	sub *client.Subscription
}
