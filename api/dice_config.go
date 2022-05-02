package api

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"sealdice-core/dice"
	"strconv"
	"strings"
)

type DiceConfigInfo struct {
	// 注：form其实不需要
	CommandPrefix           []string `json:"commandPrefix" form:"commandPrefix"`                     // 指令前缀
	DiceMasters             []string `json:"diceMasters" form:"diceMasters"`                         // 骰主设置，需要格式: 平台:帐号
	NoticeIds               []string `json:"noticeIds"`                                              // 通知设置，需要格式: 平台:帐号
	OnlyLogCommandInGroup   bool     `json:"onlyLogCommandInGroup" form:"onlyLogCommandInGroup"`     // 日志中仅记录命令
	OnlyLogCommandInPrivate bool     `json:"onlyLogCommandInPrivate" form:"onlyLogCommandInPrivate"` // 日志中仅记录命令
	WorkInQQChannel         bool     `json:"workInQQChannel"`                                        // 在QQ频道中开启
	MessageDelayRangeStart  float64  `json:"messageDelayRangeStart" form:"messageDelayRangeStart"`   // 指令延迟区间
	MessageDelayRangeEnd    float64  `json:"messageDelayRangeEnd" form:"messageDelayRangeEnd"`
	UIPassword              string   `json:"uiPassword" form:"uiPassword"`
	HelpDocEngineType       int      `json:"helpDocEngineType"`
	MasterUnlockCode        string   `json:"masterUnlockCode" form:"masterUnlockCode"`
	ServeAddress            string   `json:"serveAddress" form:"serveAddress"`
	MasterUnlockCodeTime    int64    `json:"masterUnlockCodeTime"`
	LogPageItemLimit        int64    `json:"logPageItemLimit"`
	FriendAddComment        string   `json:"friendAddComment"`
	AutoReloginEnable       bool     `json:"autoReloginEnable"`
	QQChannelAutoOn         bool     `json:"QQChannelAutoOn"`
	QQChannelLogMessage     bool     `json:"QQChannelLogMessage"`

	HelpMasterInfo     string                        `json:"helpMasterInfo"`                               // help中骰主信息
	HelpMasterLicense  string                        `json:"helpMasterLicense"`                            // help中使用协议
	ExtDefaultSettings []*dice.ExtDefaultSettingItem `yaml:"extDefaultSettings" json:"extDefaultSettings"` // 新群扩展按此顺序加载
}

func DiceConfig(c echo.Context) error {
	if !doAuth(c) {
		return c.JSON(http.StatusForbidden, nil)
	}

	password := ""
	if myDice.Parent.UIPasswordHash != "" {
		password = "------"
	}

	limit := myDice.UILogLimit
	if limit == 0 {
		limit = 100
	}
	myDice.UnlockCodeUpdate(false)

	info := DiceConfigInfo{
		CommandPrefix:           myDice.CommandPrefix,
		DiceMasters:             myDice.DiceMasters,
		NoticeIds:               myDice.NoticeIds,
		OnlyLogCommandInPrivate: myDice.OnlyLogCommandInPrivate,
		OnlyLogCommandInGroup:   myDice.OnlyLogCommandInGroup,
		MessageDelayRangeStart:  myDice.MessageDelayRangeStart,
		MessageDelayRangeEnd:    myDice.MessageDelayRangeEnd,
		UIPassword:              password,
		MasterUnlockCode:        myDice.MasterUnlockCode,
		MasterUnlockCodeTime:    myDice.MasterUnlockCodeTime,
		WorkInQQChannel:         myDice.WorkInQQChannel,
		LogPageItemLimit:        limit,
		FriendAddComment:        myDice.FriendAddComment,
		AutoReloginEnable:       myDice.AutoReloginEnable,
		QQChannelAutoOn:         myDice.QQChannelAutoOn,
		QQChannelLogMessage:     myDice.QQChannelLogMessage,

		HelpMasterInfo:     myDice.HelpMasterInfo,
		HelpMasterLicense:  myDice.HelpMasterLicense,
		ExtDefaultSettings: myDice.ExtDefaultSettings,

		ServeAddress:      myDice.Parent.ServeAddress,
		HelpDocEngineType: myDice.Parent.HelpDocEngineType,
	}
	return c.JSON(http.StatusOK, info)
}

func DiceConfigSet(c echo.Context) error {
	if !doAuth(c) {
		return c.JSON(http.StatusForbidden, nil)
	}

	jsonMap := make(map[string]interface{})
	err := json.NewDecoder(c.Request().Body).Decode(&jsonMap)

	stringConvert := func(val interface{}) []string {
		var lst []string
		for _, i := range val.([]interface{}) {
			t := i.(string)
			if t != "" {
				lst = append(lst, t)
			}
		}
		return lst
	}

	if err == nil {
		if val, ok := jsonMap["commandPrefix"]; ok {
			myDice.CommandPrefix = stringConvert(val)
		}

		if val, ok := jsonMap["diceMasters"]; ok {
			myDice.DiceMasters = stringConvert(val)
		}

		if val, ok := jsonMap["noticeIds"]; ok {
			myDice.NoticeIds = stringConvert(val)
		}

		if val, ok := jsonMap["onlyLogCommandInGroup"]; ok {
			myDice.OnlyLogCommandInGroup = val.(bool)
		}

		if val, ok := jsonMap["onlyLogCommandInPrivate"]; ok {
			myDice.OnlyLogCommandInPrivate = val.(bool)
		}

		if val, ok := jsonMap["autoReloginEnable"]; ok {
			myDice.AutoReloginEnable = val.(bool)
		}

		if val, ok := jsonMap["workInQQChannel"]; ok {
			myDice.WorkInQQChannel = val.(bool)
		}

		if val, ok := jsonMap["QQChannelLogMessage"]; ok {
			myDice.QQChannelLogMessage = val.(bool)
		}

		if val, ok := jsonMap["QQChannelAutoOn"]; ok {
			myDice.QQChannelAutoOn = val.(bool)
		}

		if val, ok := jsonMap["UILogLimit"]; ok {
			val, err := strconv.ParseInt(val.(string), 10, 64)
			if err == nil {
				if val >= 0 {
					myDice.LogWriter.LogLimit = val
				}
			}
		}

		if val, ok := jsonMap["messageDelayRangeStart"]; ok {
			f, err := strconv.ParseFloat(val.(string), 64)
			if err == nil {
				if f < 0 {
					f = 0
				}
				if myDice.MessageDelayRangeEnd < f {
					myDice.MessageDelayRangeEnd = f
				}
				myDice.MessageDelayRangeStart = f
			}
		}

		if val, ok := jsonMap["messageDelayRangeEnd"]; ok {
			f, err := strconv.ParseFloat(val.(string), 64)
			if err == nil {
				if f < 0 {
					f = 0
				}

				if f >= myDice.MessageDelayRangeStart {
					myDice.MessageDelayRangeEnd = f
				}
			}
		}

		if val, ok := jsonMap["friendAddComment"]; ok {
			myDice.FriendAddComment = strings.TrimSpace(val.(string))
		}

		if val, ok := jsonMap["uiPassword"]; ok {
			if !dm.JustForTest {
				myDice.Parent.UIPasswordHash = val.(string)
			}
		}

		if val, ok := jsonMap["helpMasterInfo"]; ok {
			myDice.HelpMasterInfo = strings.TrimSpace(val.(string))
		}

		if val, ok := jsonMap["helpMasterLicense"]; ok {
			myDice.HelpMasterLicense = strings.TrimSpace(val.(string))
		}

		if val, ok := jsonMap["extDefaultSettings"]; ok {
			data, err := json.Marshal(val)
			if err == nil {
				items := []*dice.ExtDefaultSettingItem{}
				err := json.Unmarshal(data, &items)
				if err == nil {
					myDice.ExtDefaultSettings = items
					myDice.ApplyExtDefaultSettings()
				}
			}
		}

		if val, ok := jsonMap["serveAddress"]; ok {
			if !dm.JustForTest {
				myDice.Parent.ServeAddress = val.(string)
			}
		}

		myDice.Parent.Save()
	} else {
		fmt.Println(err)
	}
	return c.JSON(http.StatusOK, nil)
}