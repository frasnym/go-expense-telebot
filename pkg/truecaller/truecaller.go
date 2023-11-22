package truecaller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/frasnym/go-expense-telebot/common"
	"github.com/frasnym/go-expense-telebot/common/logger"
	"github.com/frasnym/go-expense-telebot/config"
)

func GetPhoneNumberInformation(ctx context.Context, phoneNumber string) (*SearchResponse, error) {
	var err error
	defer func() {
		logger.LogService(ctx, "GetPhoneNumberInformation", err)
	}()

	url := fmt.Sprintf("https://asia-south1-truecaller-web.cloudfunctions.net/api/noneu/search/v1?q=%s&countryCode=id&type=41", common.RemovePrefix("62", phoneNumber))
	method := http.MethodGet

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		err = fmt.Errorf("err http.NewRequest: %w", err)
		return nil, err
	}

	req.Header.Add("authority", "asia-south1-truecaller-web.cloudfunctions.net")
	req.Header.Add("accept", "*/*")
	req.Header.Add("accept-language", "en-US,en;q=0.9,id;q=0.8")
	req.Header.Add("authorization", fmt.Sprintf("Bearer %s", config.GetConfig().TrueCallerToken))
	req.Header.Add("origin", "https://www.truecaller.com")
	req.Header.Add("referer", "https://www.truecaller.com/")
	req.Header.Add("sec-ch-ua", `"Chromium";v="118", "Google Chrome";v="118", "Not=A?Brand";v="99"`)
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-ch-ua-platform", `"macOS"`)
	req.Header.Add("sec-fetch-dest", "empty")
	req.Header.Add("sec-fetch-mode", "cors")
	req.Header.Add("sec-fetch-site", "cross-site")
	req.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("err client.Do: %w", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read and print the entire response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("err io.ReadAll: %w", err)
		return nil, err
	}
	rawResponse := string(responseBody)
	logger.Debug(ctx, rawResponse)

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("err resp.StatusCode: %d", resp.StatusCode)
		return nil, err
	}

	var result SearchResponse
	if err = json.Unmarshal(responseBody, &result); err != nil {
		err = fmt.Errorf("err json.Unmarshal: %w", err)
		return nil, err
	}
	result.Raw = string(responseBody)

	return &result, nil
}
