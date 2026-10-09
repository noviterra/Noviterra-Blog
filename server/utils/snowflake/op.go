package snowflake

import "time"

// getCurrentTime 获取当前的时间戳
func (s *SnowFlake) getCurrentTime() int64 {
	return time.Now().UnixNano()/int64(time.Millisecond) - epoch // 毫秒级
}

// waitForNextMillisecond sleep进入下一毫秒
func (s *SnowFlake) waitForNextMillisecond(lastTimeStamp int64) int64 {
	timeStamp := s.getCurrentTime()
	for timeStamp <= lastTimeStamp {
		timeStamp = s.getCurrentTime()
	}
	return timeStamp
}

// GenerateSnowFlakeID 生成雪花ID 时间戳+机器ID+时钟序列+序列号
func (s *SnowFlake) GenerateSnowFlakeID() int64 {
	// 互斥锁
	s.mu.Lock()
	defer s.mu.Unlock()

	timeStamp := s.getCurrentTime()

	// 如果当前时间戳与上次时间戳不同 重置时钟序列
	if timeStamp != s.lastTimeStamp {
		s.sequenceID = 0
		s.sequence = 0
		s.workerID = 0
		s.lastTimeStamp = timeStamp
	} else if s.sequence < maxSequence { // 先增加序列号
		s.sequence++
	} else if s.sequenceID < maxSequenceID { // 序列号满后 进位时钟序列+1 序列号归0
		s.sequenceID++
		s.sequence = 0
	} else { // 序列号和时钟序列都满 进入下一毫秒
		s.lastTimeStamp = s.waitForNextMillisecond(s.lastTimeStamp)
		s.sequence = 0
		s.sequenceID = 0
	}

	// 拼接得到唯一ID
	id := (timeStamp << timestampShift) |
		(s.workerID << workerIDShift) |
		(s.sequenceID << sequenceIDShift) |
		s.sequence

	return id
}
