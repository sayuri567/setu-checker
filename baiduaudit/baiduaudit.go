package baiduaudit

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/url"
	"time"

	"github.com/sayuri567/tool/util/curl"
	"github.com/sayuri567/tool/util/fileutil"
	"github.com/sirupsen/logrus"
)

const (
	aiImgCensor = "https://aip.baidubce.com/rest/2.0/solution/v1/img_censor/v2/user_defined"
	tokenUtl    = "https://openapi.baidu.com/oauth/2.0/token"

	tokenFile = "./access_token"
)

type Client struct {
	ak          string
	sk          string
	accessToken string
	expireAt    time.Time
}

func GetClient(ak, sk string) *Client {
	return &Client{ak: ak, sk: sk}
}

type CheckImageResp struct {
	LogId          int64                  `json:"log_id"`
	ErrorCode      int64                  `json:"error_code"`
	ErrorMsg       string                 `json:"error_msg"`
	Conclusion     string                 `json:"conclusion"`
	ConclusionType int                    `json:"conclusionType"`
	Data           CheckImageDataItemList `json:"data"`
	RawData        map[string]interface{} `json:"rawData"`
}

type CheckImageDataItem struct {
	Type           int                  `json:"type"`
	SubType        int                  `json:"subType"`
	Conclusion     string               `json:"conclusion"`
	ConclusionType int                  `json:"conclusionType"`
	Probability    float64              `json:"probability"`
	Msg            string               `json:"msg"`
	Codes          []string             `json:"codes"`
	DatasetName    string               `json:"datasetName"`
	Stars          []CheckImageStarItem `json:"stars"`
	Hits           []CheckImageHitsItem `json:"hits"`
}

type CheckImageStarItem struct {
	Probability float64 `json:"probability"`
	Name        string  `json:"name"`
	DatasetName string  `json:"datasetName"`
}

type CheckImageHitsItem struct {
	Probability float64  `json:"probability"`
	DatasetName string   `json:"datasetName"`
	Words       []string `json:"words"`
}

type CheckImageDataItemList []CheckImageDataItem

func (l CheckImageDataItemList) Len() int {
	return len(l)
}

func (l CheckImageDataItemList) Less(i, j int) bool {
	return l[i].Probability < l[j].Probability
}

func (l CheckImageDataItemList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (this *Client) CheckImages(filepath string) (*CheckImageResp, error) {
	buff, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	base64String := base64.StdEncoding.EncodeToString(buff)
	params := url.Values{
		"image": []string{base64String},
	}

	accessToken, err := this.GetAccessToken()
	if err != nil {
		return nil, err
	}

	data := &CheckImageResp{}
	var resp *curl.HttpResponse
	for i := 0; i < 3; i++ {
		if i > 0 {
			time.Sleep(time.Second * time.Duration(i*3))
		}
		resp, err = curl.Post(aiImgCensor+"?access_token="+accessToken, params, nil)
		if err != nil {
			continue
		}
		err = json.Unmarshal(resp.Body, data)
		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, err
	}

	return data, nil
}

type AuthResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	Scope            string `json:"scope"`
	SessionKey       string `json:"session_key"`
	SessionSecret    string `json:"session_secret"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ExpiresAt        int64  `json:"ExpiresAt"` // 自定义字段
}

func (this *Client) GetAccessToken() (string, error) {
	authresponse := new(AuthResponse)
	if this.accessToken == "" {
		accessToken, err := fileutil.ReadFile(tokenFile)
		if err != nil {
			logrus.WithError(err).Warn("access_file not found")
		}
		if len(accessToken) > 0 {
			err = json.Unmarshal([]byte(accessToken), authresponse)
			if err == nil {
				this.setAccessToken(authresponse)
			}
		}
	}
	if this.accessToken != "" && time.Now().Before(this.expireAt) {
		return this.accessToken, nil
	}

	resp, err := curl.Post(tokenUtl, url.Values{
		"grant_type":    []string{"client_credentials"},
		"client_id":     []string{this.ak},
		"client_secret": []string{this.sk},
	}, nil)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(resp.Body, authresponse)
	if err != nil {
		return "", err
	}
	if authresponse.Error != "" || authresponse.AccessToken == "" {
		return "", errors.New("授权失败:" + authresponse.ErrorDescription)
	}

	this.setAccessToken(authresponse)

	return this.accessToken, err
}

func (this *Client) setAccessToken(authresponse *AuthResponse) {
	this.accessToken = authresponse.AccessToken
	if authresponse.ExpiresAt > 0 {
		this.expireAt = time.Unix(authresponse.ExpiresAt, 0)
	} else {
		this.expireAt = time.Now().Add(time.Second * time.Duration(authresponse.ExpiresIn-60))
	}
	authresponse.ExpiresAt = this.expireAt.Unix()
	accessText, _ := json.Marshal(authresponse)
	fileutil.CreateFile(tokenFile, string(accessText), true)
}
