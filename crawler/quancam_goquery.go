package crawler

import (
	"context"
	"devread/custom_error"
	"devread/repository"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"devread/helper"
	"devread/model"
	"log"
	"runtime"
	"strings"
)

const urlBase = "https://quan-cam.com"

func checkError(err error) {
	if err != nil {
		log.Println(err)
	}
}

func getOnePage(pathURL string) ([]model.Post, error) {
	response, err := helper.HttpClient.GetRequestWithRetries(pathURL)
	checkError(err)
	defer response.Body.Close()
	doc, err := goquery.NewDocumentFromReader(response.Body)
	checkError(err)
	posts := make([]model.Post, 0)
	doc.Find("div[class=post]").Each(func(i int, s *goquery.Selection) {
		var quancamPost model.Post
		quancamPost.Name = s.Find("h3.post__title > a").Text()
		link, _ := s.Find("h3.post__title > a").Attr("href")
		quancamPost.Link = urlBase + link
		quancamPost.Tag = strings.ToLower(strings.Replace(
			strings.Replace(
				s.Find("span.tagging > a").Text(), "\n", "", -1), "#", " ", -1))
		posts = append(posts, quancamPost)
	})
	return posts, nil
}

func QuancamPostV2(postRepo repository.PostRepo) {
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))
	group, ctx := errgroup.WithContext(context.Background())

	for page := 1; page <= 5; page++ {
		pathURL := fmt.Sprintf("%s/posts?page=%d", urlBase,page)
		err := sem.Acquire(ctx, 1)
		if err != nil {
			fmt.Printf("Acquire err = %+v\n", err)
			continue
		}
		group.Go(func() error {
			defer sem.Release(1)

			//do work
			posts, err := getOnePage(pathURL)
			checkError(err)

			queue := helper.NewJobQueue(runtime.NumCPU())
			queue.Start()
			defer queue.Stop()
			for _, post := range posts {
				queue.Submit(&QuancamProcessV2{
					post:     post,
					postRepo: postRepo,
				})
			}

			return nil
		})
	}
	if err := group.Wait(); err != nil {
		fmt.Printf("g.Wait() err = %+v\n", err)
	}
}

type QuancamProcessV2 struct {
	post     model.Post
	postRepo repository.PostRepo
}

func (process *QuancamProcessV2) Process() {
	// select post by name
	cacheRepo, err := process.postRepo.SelectPostByName(context.Background(), process.post.Name)
	if err == custom_error.PostNotFound {
		// insert post to database
		fmt.Println("Add: ", process.post.Name)
		_, err = process.postRepo.SavePost(context.Background(), process.post)
		checkError(err)
		return
	}

	// update post
	if process.post.Name != cacheRepo.Name ||
		process.post.Link != cacheRepo.Link {
		fmt.Println("Updated: ", process.post.Name)
		_, err = process.postRepo.UpdatePost(context.Background(), process.post)
		checkError(err)
	}
}
