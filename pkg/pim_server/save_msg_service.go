package pim_server

import (
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"pim/pkg/models"
	"time"
)

func (s *server) StartSaveMessageEventService() {
	ticker := time.NewTicker(time.Second * 10)

	var saveMessageItems []*models.SingleMessage
	var saveStart = atomic.NewBool(false)

	doSaveMessage := func(timeNow time.Time) {
		if len(saveMessageItems) == 0 {
			return
		}

		if saveStart.Load() {
			// 有任务再保存
			return
		}

		saveStart.Store(true)
		defer saveStart.Store(false)

		// 保存数据

		err := s.db.Create(saveMessageItems).Error
		if err != nil {
			s.logger.Debug("Save info ", zap.Int("size", len(saveMessageItems)))
		}

		saveMessageItems = saveMessageItems[0:0]

	}

	for true {
		select {
		case timeNow := <-ticker.C:
			doSaveMessage(timeNow)

		case item := <-s.saveMessageChan:
			saveMessageItems = append(saveMessageItems, item)

		case <-s.closeServer:
			s.logger.Warn("StartSaveMessageEventService stop")
			doSaveMessage(time.Now())
			return

		}
	}

}
