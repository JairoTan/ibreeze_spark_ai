package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type UsageDetail struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
}

type WenXinResp struct {
	Id               string      `json:"id"`
	Object           string      `json:"object"`
	Created          int64       `json:"created"`
	Result           string      `json:"result"`
	IsTruncated      bool        `json:"is_truncated"`
	NeedClearHistory bool        `json:"need_clear_history"`
	Usage            UsageDetail `json:"usage"`
}

const API_KEY = "PRiRux1KiXtvIvDeqLLqov8G"
const SECRET_KEY = "trdRGgcN31pyVhDBeGyzllzenDem8Wo0"

func WenXinAnswer(question string) string {

	url := "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions_pro?access_token=" + GetAccessToken()

	// 构建请求参数
	reqData := map[string]interface{}{
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": question,
			},
		},
		"temperature": 0.1,
		//说明：
		//（1）较高的数值会使输出更加随机，而较低的数值会使其更加集中和确定
		//（2）默认0.95，范围 (0, 1.0]，不能为0
		//（3）建议该参数和top_p只设置1个
		//（4）建议top_p和temperature不要同时更改
		"system": "我是铂睿思AI客服，有什么可以帮您",
	}
	reqBody, _ := json.Marshal(reqData)

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))

	if err != nil {
		fmt.Println(err)
		return ""
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	fmt.Println("文心一样回调:", string(body))

	var resp WenXinResp
	err = json.Unmarshal(body, &resp)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return resp.Result
}

/**
 * 使用 AK，SK 生成鉴权签名（Access Token）
 * @return string 鉴权签名信息（Access Token）
 */
func GetAccessToken() string {
	url := "https://aip.baidubce.com/oauth/2.0/token"
	postData := fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s", API_KEY, SECRET_KEY)
	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(postData))
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	accessTokenObj := map[string]string{}
	json.Unmarshal([]byte(body), &accessTokenObj)
	return accessTokenObj["access_token"]
}
