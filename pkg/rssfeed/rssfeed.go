package rssfeed

import (
	"fmt"

	"github.com/mmcdole/gofeed"
	"github.com/mmcdole/gofeed/rss"
)

type MyCustomTranslator struct {
	defaultTranslator *gofeed.DefaultRSSTranslator
}

func GetEpisodes() {
	fp := gofeed.NewParser()
	fp.RSSTranslator = ShowRSSTranslator()
	feed, _ := fp.ParseURL("http://showrss.info/user/224423.rss?magnets=true&namespaces=true&name=null&quality=null&re=null")
	fmt.Println(feed.Items)

}

func ShowRSSTranslator() *MyCustomTranslator {
	t := &MyCustomTranslator{}

	// We create a DefaultRSSTranslator internally so we can wrap its Translate
	// call since we only want to modify the precedence for a single field.
	t.defaultTranslator = &gofeed.DefaultRSSTranslator{}
	return t
}

func (ct *MyCustomTranslator) Translate(feed interface{}) (*gofeed.Feed, error) {
	rss, found := feed.(*rss.Feed)
	if !found {
		return nil, fmt.Errorf("Feed did not match expected type of *rss.Feed")
	}

	f, err := ct.defaultTranslator.Translate(rss)
	if err != nil {
		return nil, err
	}

	// if rss.ITunesExt != nil && rss.ITunesExt.Author != "" {
	// 	f.Author = rss.ITunesExt.Author
	// } else {
	// 	f.Author = rss.ManagingEditor
	// }
	return f, err
}
