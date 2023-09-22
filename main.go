package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"ibreeze_spark_ai/util"

	"github.com/gin-gonic/gin"
)

// 在主函数外部定义一个通道，用于接收异步处理中的答案
var answerChannel = make(chan string)

func main() {
	r := gin.Default()

	// 客服消息
	r.POST("/customer/message", func(c *gin.Context) {
		// 获取请求参数
		var reqMsg struct {
			ToUserName   string `json:"ToUserName"`   // 小程序/公众号的原始ID，资源复用配置多个时可以区别消息是给谁的
			FromUserName string `json:"FromUserName"` // 该小程序/公众号的用户身份openid
			MsgType      string `json:"MsgType"`
			Content      string `json:"Content"`
			CreateTime   int64  `json:"CreateTime"`
		}

		if err := c.BindJSON(&reqMsg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "JSON数据包解析失败"})
			return
		}

		fmt.Printf("公众号 %s 接收用户openid为 %s 的 %s 消息：%s\n", reqMsg.ToUserName, reqMsg.FromUserName, reqMsg.MsgType, reqMsg.Content)
		fmt.Println("ToUserName:", reqMsg.ToUserName)
		fmt.Println("FromUserName:", reqMsg.FromUserName)

		// 异步处理获取星火AI回答
		go func() {
			var answer string
			if reqMsg.MsgType == "text" {
				answer = util.SparkAnswer(reqMsg.Content)
				fmt.Println("星火AI回答：", answer)

				// 将答案发送到通道中
				answerChannel <- answer
			} else {
				answer = "暂不支持文本以外的消息回复"
				fmt.Println("不是文本消息,拒绝回答!")
			}

			// 构建请求参数
			msgBody := struct {
				ToUser  string `json:"touser"`
				MsgType string `json:"msgtype"`
				Text    struct {
					Content string `json:"content"`
				} `json:"text"`
			}{
				ToUser:  reqMsg.FromUserName,
				MsgType: "text",
			}
			msgBody.Text.Content = answer

			// 将消息体转换为 JSON
			requestBody, err := json.Marshal(msgBody)
			if err != nil {
				fmt.Println("JSON编码失败:", err)
				return
			}

			// 发送 POST 请求到微信接口
			url := "http://api.weixin.qq.com/cgi-bin/message/custom/send"
			resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
			fmt.Println("微信接口url：", url)
			if err != nil {
				fmt.Println("请求微信接口报错：", err)
				return
			}

			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("读取响应体失败:", err)
				return
			}
			responseStr := string(bodyBytes)
			fmt.Println("微信接口返回内容：", responseStr)

			defer resp.Body.Close()

			// 处理响应
			// 这里你可以解析 resp.Body 中的响应数据或者处理其他逻辑
			// 例如，检查响应状态码和返回的 JSON 数据等

			if resp.StatusCode != http.StatusOK {
				fmt.Println("请求失败，状态码:", resp.StatusCode)
				return
			}

			fmt.Println("异步处理完成")
		}()

		// 在主 goroutine 中等待答案并将其返回到客户端
		answerFromAsync := <-answerChannel

		c.JSON(http.StatusOK, gin.H{
			"ToUserName":   reqMsg.FromUserName,
			"FromUserName": reqMsg.ToUserName,
			"CreateTime":   reqMsg.CreateTime,
			"MsgType":      reqMsg.MsgType,
			"Content":      answerFromAsync, // 使用从通道接收到的答案
		})
	})

	////客服消息
	//r.POST("/customer/message", func(c *gin.Context) {
	//	// 获取请求参数
	//	var reqMsg struct {
	//		ToUserName   string `json:"ToUserName"`   // 小程序/公众号的原始ID，资源复用配置多个时可以区别消息是给谁的
	//		FromUserName string `json:"FromUserName"` // 该小程序/公众号的用户身份openid
	//		MsgType      string `json:"MsgType"`
	//		Content      string `json:"Content"`
	//		CreateTime   int64  `json:"CreateTime"`
	//	}
	//
	//	if err := c.BindJSON(&reqMsg); err != nil {
	//		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON数据包解析失败"})
	//		return
	//	}
	//
	//	fmt.Printf("公众号 %s 接收用户openid为 %s 的 %s 消息：%s", reqMsg.ToUserName, reqMsg.FromUserName, reqMsg.MsgType, reqMsg.Content)
	//	fmt.Println("ToUserName:", reqMsg.ToUserName)
	//	fmt.Println("FromUserName:", reqMsg.FromUserName)
	//	//异步获取星火AI答案
	//	var answer string
	//	if reqMsg.MsgType == "text" {
	//		answer = util.SparkAnswer(reqMsg.Content)
	//		fmt.Println("星火AI回答：", answer)
	//	} else {
	//		answer = "暂不支持文本以外的消息回复"
	//		fmt.Println("不是文本消息,拒绝回答!")
	//	}
	//
	//	// 构建请求参数
	//	msgBody := struct {
	//		ToUser  string `json:"touser"`
	//		MsgType string `json:"msgtype"`
	//		Text    struct {
	//			Content string `json:"content"`
	//		} `json:"text"`
	//	}{
	//		ToUser:  reqMsg.FromUserName,
	//		MsgType: "text",
	//	}
	//	msgBody.Text.Content = answer
	//
	//	// 将消息体转换为 JSON
	//	requestBody, err := json.Marshal(msgBody)
	//	if err != nil {
	//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	//		return
	//	}
	//
	//	// 发送 POST 请求到微信接口
	//	url := "http://api.weixin.qq.com/cgi-bin/message/custom/send"
	//	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	//	fmt.Println("微信接口url：", url)
	//	fmt.Println("微信接口返回内容：", resp)
	//	if err != nil {
	//		fmt.Println("请求微信接口报错：", err)
	//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	//		return
	//	}
	//	defer resp.Body.Close()
	//
	//	// 处理响应
	//	// 这里你可以解析 resp.Body 中的响应数据或者处理其他逻辑
	//	// 例如，检查响应状态码和返回的 JSON 数据等
	//
	//	if resp.StatusCode != http.StatusOK {
	//		c.JSON(http.StatusInternalServerError, gin.H{"error": "请求失败"})
	//		return
	//	}
	//
	//	c.JSON(http.StatusOK, gin.H{
	//		"ToUserName":   reqMsg.FromUserName,
	//		"FromUserName": reqMsg.ToUserName,
	//		"CreateTime":   reqMsg.CreateTime,
	//		"MsgType":      reqMsg.MsgType,
	//		"Content":      answer,
	//	})
	//})

	//被动消息回复
	r.POST("/spark/answer", func(c *gin.Context) {
		var reqMsg struct {
			ToUserName   string `json:"ToUserName"`
			FromUserName string `json:"FromUserName"`
			MsgType      string `json:"MsgType"`
			Content      string `json:"Content"`
			CreateTime   int64  `json:"CreateTime"`
		}

		if err := c.BindJSON(&reqMsg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "JSON数据包解析失败"})
			return
		}

		if reqMsg.MsgType == "text" {
			//获取星火AI
			answer := util.SparkAnswer(reqMsg.Content)
			c.JSON(http.StatusOK, gin.H{
				"ToUserName":   reqMsg.FromUserName,
				"FromUserName": reqMsg.ToUserName,
				"CreateTime":   reqMsg.CreateTime,
				"MsgType":      "text",
				"Content":      answer,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"ToUserName":   reqMsg.FromUserName,
				"FromUserName": reqMsg.ToUserName,
				"CreateTime":   reqMsg.CreateTime,
				"MsgType":      "text",
				"Content":      "暂时不支持文本之外的回复",
			})
		}
	})

	r.Run(":80")
}
