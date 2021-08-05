package helper

import (
	"devread/log"
	"net/http"
	"time"

	"github.com/cenkalti/backoff"
	"go.uber.org/zap"
)

const (
	errUnexpectedResponse = "unexpected response: %s"
)

func getRequest(pathURL string) (*http.Response, error) {
	log := log.WriteLog()

	req, _ := http.NewRequest("GET", pathURL, nil)
	client := &http.Client{}
	resp, err := client.Do(req)

	log.Sugar().Info(zap.String("Get ", pathURL), zap.Int("Status Code ", resp.StatusCode))
	if err != nil {
		log.Error("Phản hồi Không mong đợi ", zap.Int("Status Code", resp.StatusCode), zap.Error(err))
		return nil, err
	}
	return resp, nil
}

func GetRequestWithRetries(api string) (*http.Response, error) {
	var err error
	var resp *http.Response

	log := log.WriteLog()

	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = 5 * time.Minute

	for {
		resp, err = getRequest(api)
		if err == nil {
			break
		}
		d := bo.NextBackOff()
		if d == backoff.Stop {
			log.Debug("Hết thời gian thử lại")
		}
		log.Error("Request lỗi ", zap.Error(err))
		log.Sugar().Info("Thử lại trong ", d)
		time.Sleep(d)
	}

	// Tất cả các lần thử lại không thành công
	if err != nil {
		return nil, err
	}
	return resp, nil
}