package pim_server

import "time"

func (s *server) StartAutoClearFileTaskService() {

	s.logger.Info("Start StartAutoClearFileTaskService")
	ticker := time.NewTicker(time.Minute * 5)

	for true {
		select {
		case <-s.closeServer:
			s.logger.Debug("Close StartAutoClearFileTaskService")
			//timeNow := time.Now()
			//doAutoClearFileTaskService(s, timeNow)

			return
		case timeNow := <-ticker.C:
			//doAutoClearFileTaskService(s, timeNow)
			s.dao.DoClearKey(timeNow)
		}
	}

}
