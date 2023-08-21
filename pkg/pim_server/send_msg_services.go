package pim_server

import (
	"google.golang.org/protobuf/types/known/anypb"
	"pim/api"
	"pim/pkg/tools"
)

// StartSenderMessageEventService 监听发送消息服务
func (s *server) StartSenderMessageEventService() {

	// 处理发送消息的数据

	procSenderMsg := func(msg *api.Message) {
		defer tools.HandlePanic(s.logger, "StartSenderMessageEventServiceProcMessage")

		// 转发到指定的用户
		// 新消息事件

		if msg.ChatID < 0 {
			// 这是群

			// TODO 群处理
			targetChatType := msg.ChatID & 0xff0000000000 // 群标志
			if targetChatType == 0x010000000000 {
				// 普通群 -- 200 人以下规模 私有

				// 获取群ID
				// 获取
				// 修改信息状态
				// 生成Message的数据报（Any）
				// 耗时操作（向群成员分发）
				// 获取所有群成员
				// go for分发

			} else if targetChatType == 0x100000000000 {
				// 超级群 -- 100000 人规模 公开

			}

		} else {
			// 这是个人
			// 查找 个人id
			// 查找用户id
			localUserID := msg.ChatID & 0xffff

			// 搜索网络的用户

			// 修改 消息状态
			msg.Status = api.MessageStatusEnum_MessageStatusSuccess

			// 先推送给我对方

			// 拷贝一份数据

			// 这里的时候已经序列化了
			targetUserMessageBody, _ := anypb.New(msg)
			pushTargetEventData := &api.UpdateEventDataType{
				Type: api.UpdateEventDataType_NewMessage,
				Body: targetUserMessageBody,
			}

			s.pim.UserStreamClientMap.PushUserEvent(localUserID, pushTargetEventData)

			// 推给我自己

			if msg.ChatID == localUserID {
				msg.ChatID = msg.Sender
			}
			myUserMessageBody, _ := anypb.New(msg)
			pushMySelfEventData := &api.UpdateEventDataType{
				Type: api.UpdateEventDataType_NewMessage,
				Body: myUserMessageBody,
			}

			s.pim.UserStreamClientMap.PushUserEvent(msg.Sender, pushMySelfEventData)

			//recvUserConn := s.
			//myUserConn := s.wsUserToConn[msg.Sender]

			// 接收的用户id
			// 将我的消息

			//msg.MsgStatus = models.SenderMsgStateSuccess
			//
			//mySenderBuffer, _ := json.Marshal(msg)
			//
			//eventPackage := tools.ProtocolPackage{
			//	Body:    mySenderBuffer,
			//	Version: 1,
			//	ModType: codes.EventMod,
			//	SubType: codes.EventModNewMessage,
			//}
			// 新消息事件
			//senderData := eventPackage.Bytes()
			//
			//for streamID, conn := range myUserConn {
			//	s.logger.Info("推送消息事件给多端", zap.Int64("streamID", streamID), zap.Int64("chat_id", msg.ChatID), zap.Int64("Sender", msg.Sender), zap.Int64("msg_id", msg.MsgID))
			//	conn.WriteMessage(websocket.BinaryMessage, senderData)
			//}
			//
			//// TODO 离线消息处理
			//// 接收方需要改写 chat id
			//if msg.ChatID == localUserID {
			//	msg.ChatID = msg.Sender
			//}
			//
			//otherSenderBuffer, _ := json.Marshal(msg)
			//
			//otherEventPackage := tools.ProtocolPackage{
			//	Body:    otherSenderBuffer,
			//	Version: 1,
			//	ModType: codes.EventMod,
			//	SubType: codes.EventModNewMessage,
			//}
			//// 新消息事件
			//otherSenderData := otherEventPackage.Bytes()
			//
			//for streamID, conn := range recvUserConn {
			//	s.logger.Info("推送消息事件给用户", zap.Int64("streamID", streamID), zap.Int64("chat_id", msg.ChatID), zap.Int64("Sender", msg.Sender), zap.Int64("msg_id", msg.MsgID))
			//	// 判断接收的用户有没有缓存这个发送者
			//	canSender := conn.SenderMsgCheck(msg.Sender, msg.Sender, false)
			//	// 判断是否可以发送 , 不可以发送则跳过
			//	if !canSender {
			//		continue
			//	}
			//	conn.WriteMessage(websocket.BinaryMessage, otherSenderData)
			//}

		}

	}

	for {

		select {
		case item := <-s.sendMessageChan:

			// 处理发送的消息
			procSenderMsg(item)
		case <-s.closeServer:
			s.logger.Warn("StartSenderMessageEventService stop")
			return
		}
	}

}
