package mid

import (
	g "wp_template_display/internal/global"
	m "wp_template_display/internal/models"

	"encoding/json"
	"fmt"
	"time"
	"wp_template_display/pkg"

	"github.com/gofiber/fiber/v2"
)

func getLocalsString(c *fiber.Ctx, key string) string {
	if c.Locals(key) == nil {
		return ""
	}
	return c.Locals(key).(string)
}

func RequestLogMiddleware(c *fiber.Ctx) error {
	if c.Locals("Action") == nil {
		return c.Next()
	}

	// 获取action
	action := c.Locals("Action").(string)
	username := getLocalsString(c, "Username")

	// 获取请求参数并转为JSON字符串
	params := make(map[string]interface{})
	params["body"] = string(c.Body())

	// ContentBody
	contentBody := c.Locals("ContentBody")
	if contentBody != nil {
		// contentBody 为 map[string]interface{}
		// 转为JSON 字符串 替换 params["body"]
		contentBodyString, _ := json.Marshal(contentBody)
		params["body"] = string(contentBodyString)
	}
	params["params"] = c.AllParams()
	paramsString, _ := json.Marshal(params)
	paramsJson := string(paramsString)

	// 获取头部并转为JSON字符串
	headerBytes, _ := json.Marshal(c.GetReqHeaders())
	headerString := string(headerBytes)

	resStatus := getLocalsString(c, "ResStatus")
	resultString := string(c.Response().Body())
	if len(resultString) > 1000 {
		resultString = fmt.Sprintf(`{"code":%s,"data":null,"msg":"success"}`, resStatus)
	}

	// 请求时间
	triggerTime := c.Locals("TriggerTime").(time.Time)
	completTime := time.Now()
	duration := completTime.Sub(triggerTime).Milliseconds()

	log := m.SysLog{
		IP:          c.IP(),
		Location:    pkg.IplibGetInfo(c.IP()),
		Username:    username,
		Content:     getLocalsString(c, "Content"),
		Action:      action,
		Header:      headerString,
		Params:      paramsJson,
		Result:      resultString,
		Status:      resStatus,
		Method:      c.Method(),
		TriggerTime: triggerTime,
		CompletTime: completTime,
		Duration:    uint32(duration),
	}

	// 写入数据库
	g.DB.Create(&log)
	return c.Next()
}
