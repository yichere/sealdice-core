package dice

import (
	"encoding/json"
	"github.com/sacOO7/gowebsocket"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type oneBotCommand struct {
	Action string      `json:"action"`
	Params interface{} `json:"params"`
	Echo   int64       `json:"echo"`
}

func (pa *PlatformAdapterQQOnebot) mustExtractId(id string) int64 {
	if strings.HasPrefix(id, "QQ:") {
		num, _ := strconv.ParseInt(id[len("QQ:"):], 10, 64)
		return num
	}
	if strings.HasPrefix(id, "QQ-Group:") {
		num, _ := strconv.ParseInt(id[len("QQ-Group:"):], 10, 64)
		return num
	}
	if strings.HasPrefix(id, "PG-QQ:") {
		num, _ := strconv.ParseInt(id[len("PG-QQ:"):], 10, 64)
		return num
	}
	return 0
}

func (pa *PlatformAdapterQQOnebot) mustExtractChannelId(id string) string {
	if strings.HasPrefix(id, "QQ-Channel:") {
		return id[len("QQ-Channel:"):]
	}
	if strings.HasPrefix(id, "QQ-Channel-Group:") {
		return id[len("QQ-Channel-Group:"):]
	}
	return ""
}

// GetGroupInfoAsync 异步获取群聊信息
func (pa *PlatformAdapterQQOnebot) GetGroupInfoAsync(groupId string) {
	type GroupMessageParams struct {
		GroupId int64 `json:"group_id"`
	}
	realGroupId := pa.mustExtractId(groupId)

	a, _ := json.Marshal(oneBotCommand{
		"get_group_info",
		GroupMessageParams{
			realGroupId,
		},
		-2,
	})

	socketSendText(pa.Socket, string(a))
}

func socketSendText(socket *gowebsocket.Socket, s string) {
	defer func() {
		if r := recover(); r != nil {
			//core.GetLogger().Error(r)
		}
	}()

	if socket != nil {
		socket.SendText(s)
	}
}

func (pa *PlatformAdapterQQOnebot) ReplyToSenderRaw(ctx *MsgContext, msg *Message, text string, flag string) {
	inGroup := msg.MessageType == "group"
	if inGroup {
		pa.ReplyGroupWithFlag(ctx, msg, text, flag)
	} else {
		pa.ReplyPersonWithFlag(ctx, msg, text, flag)
	}
}

func (pa *PlatformAdapterQQOnebot) ReplyToSender(ctx *MsgContext, msg *Message, text string) {
	pa.ReplyToSenderRaw(ctx, msg, text, "")
}

// ReplyPerson 私聊回复
func (pa *PlatformAdapterQQOnebot) ReplyPerson(ctx *MsgContext, msg *Message, text string) {
	pa.ReplyPersonWithFlag(ctx, msg, text, "")
}

func (pa *PlatformAdapterQQOnebot) ReplyPersonWithFlag(ctx *MsgContext, msg *Message, text string, flag string) {
	pa.replyPersonRaw(ctx, msg.Sender.UserId, text, flag)
}

func (pa *PlatformAdapterQQOnebot) replyPersonRaw(ctx *MsgContext, userId string, text string, flag string) {
	rawId := pa.mustExtractId(userId)
	for _, i := range ctx.Dice.ExtList {
		if i.OnMessageSend != nil {
			i.OnMessageSend(ctx, "private", userId, text, flag)
		}
	}

	if ctx.Dice != nil {
		ctx.Dice.Logger.Infof("发给(帐号%s): %s", userId, text)
	}
	time.Sleep(time.Duration((0.4 + rand.Float64()/2) * float64(time.Second)))

	type GroupMessageParams struct {
		MessageType string `json:"message_type"`
		UserId      int64  `json:"user_id"`
		Message     string `json:"message"`
	}

	a, _ := json.Marshal(oneBotCommand{
		Action: "send_msg",
		Params: GroupMessageParams{
			MessageType: "private",
			UserId:      rawId,
			Message:     text,
		},
	})

	socketSendText(pa.Socket, string(a))
}

func (pa *PlatformAdapterQQOnebot) SendTo(ctx *MsgContext, uid string, text string) {
	if strings.HasPrefix(uid, "QQ:") {
		pa.replyPersonRaw(ctx, uid, text, "")
	}
	if strings.HasPrefix(uid, "QQ-Group:") {
		pa.replyGroupRaw(ctx, uid, text, "")
	}
}

func (pa *PlatformAdapterQQOnebot) ReplyGroup(ctx *MsgContext, msg *Message, text string) {
	pa.ReplyGroupWithFlag(ctx, msg, text, "")
}

func (pa *PlatformAdapterQQOnebot) ReplyGroupWithFlag(ctx *MsgContext, msg *Message, text string, flag string) {
	pa.replyGroupRaw(ctx, msg.GroupId, text, flag)
}

func (pa *PlatformAdapterQQOnebot) replyGroupRaw(ctx *MsgContext, groupId string, text string, flag string) {
	rawId := pa.mustExtractId(groupId)

	if ctx.Session.ServiceAtNew[groupId] != nil {
		for _, i := range ctx.Session.ServiceAtNew[groupId].ActivatedExtList {
			if i.OnMessageSend != nil {
				i.OnMessageSend(ctx, "group", groupId, text, flag)
			}
		}
	}

	if ctx.Dice != nil {
		ctx.Dice.Logger.Infof("发给(群%s): %s", groupId, text)
	}
	time.Sleep(time.Duration((0.4 + rand.Float64()/2) * float64(time.Second)))

	type GroupMessageParams struct {
		GroupId int64  `json:"group_id"`
		Message string `json:"message"`
	}

	a, _ := json.Marshal(oneBotCommand{
		Action: "send_group_msg",
		Params: GroupMessageParams{
			rawId,
			text, // "golang client test",
		},
	})

	socketSendText(pa.Socket, string(a))
	//ctx.Session.Socket.SendText(string(a))
}

// SetGroupAddRequest 同意加群
func (pa *PlatformAdapterQQOnebot) SetGroupAddRequest(flag string, subType string, approve bool, reason string) {
	type DetailParams struct {
		Flag    string `json:"flag"`
		SubType string `json:"sub_type"`
		Approve bool   `json:"approve"`
		Reason  string `json:"reason"`
	}

	a, _ := json.Marshal(oneBotCommand{
		Action: "set_group_add_request",
		Params: DetailParams{
			Flag:    flag,
			SubType: subType,
			Approve: approve,
			Reason:  reason,
		},
	})

	socketSendText(pa.Socket, string(a))
}

// GetGroupMemberInfo 获取群成员信息
func (pa *PlatformAdapterQQOnebot) GetGroupMemberInfo(GroupId int64, UserId int64) {
	type DetailParams struct {
		GroupId int64 `json:"group_id"`
		UserId  int64 `json:"user_id"`
		NoCache bool  `json:"no_cache"`
	}

	a, _ := json.Marshal(oneBotCommand{
		"get_group_member_info",
		DetailParams{
			GroupId: GroupId,
			UserId:  UserId,
		},
		-3,
	})

	socketSendText(pa.Socket, string(a))
}

// SetGroupCard 设置群名片
func (pa *PlatformAdapterQQOnebot) SetGroupCard(GroupId int64, UserId int64, Card string) {
	type DetailParams struct {
		GroupId int64  `json:"group_id"`
		UserId  int64  `json:"user_id"`
		Card    string `json:"card"`
	}

	a, _ := json.Marshal(struct {
		Action string       `json:"action"`
		Params DetailParams `json:"params"`
	}{
		"set_group_card",
		DetailParams{
			GroupId: GroupId,
			UserId:  UserId,
			Card:    Card,
		},
	})

	socketSendText(pa.Socket, string(a))
}

func (pa *PlatformAdapterQQOnebot) SetFriendAddRequest(flag string, approve bool, remark string) {
	type DetailParams struct {
		Flag    string `json:"flag"`
		Remark  string `json:"remark"` // 备注名
		Approve bool   `json:"approve"`
		//Reason  string `json:"reason"`
	}

	a, _ := json.Marshal(struct {
		Action string       `json:"action"`
		Params DetailParams `json:"params"`
	}{
		"set_friend_add_request",
		DetailParams{
			Flag:    flag,
			Approve: approve,
			Remark:  remark,
		},
	})

	socketSendText(pa.Socket, string(a))
}

func (pa *PlatformAdapterQQOnebot) QuitGroup(ctx *MsgContext, id string) {
	groupId := pa.mustExtractId(id)
	type GroupMessageParams struct {
		GroupId int64 `json:"group_id"`
	}

	a, _ := json.Marshal(oneBotCommand{
		Action: "set_group_leave",
		Params: GroupMessageParams{
			groupId,
		},
	})

	socketSendText(pa.Socket, string(a))
}

func (pa *PlatformAdapterQQOnebot) GetLoginInfo() {
	a, _ := json.Marshal(struct {
		Action string `json:"action"`
		Echo   int64  `json:"echo"`
	}{
		Action: "get_login_info",
		Echo:   -1,
	})

	//if s.Socket != nil {
	socketSendText(pa.Socket, string(a))
	//s.Socket.SendText(string(a))
	//}
}