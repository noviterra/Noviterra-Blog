package snowflake

import (
	"errors"
	"sync"
)

const (
	// 每个ID的位数
	workerIDBits   = 7  // 机器ID
	sequenceIDBits = 3  // 时钟序列
	sequenceBits   = 12 // 序列号

	// 每个ID的最大值
	maxWorkID     = -1 ^ (-1 << workerIDBits) // 128
	maxSequenceID = -1 ^ (-1 << sequenceBits) //
	maxSequence   = -1 ^ (-1 << sequenceBits) // 4095 同一毫秒内最多自增4096个ID

	// 每个ID的左偏移量 用户雪花数字组成
	workerIDShift   = sequenceBits + sequenceIDBits                // 12+3=15
	sequenceIDShift = sequenceBits                                 // 12
	timestampShift  = workerIDBits + sequenceIDBits + sequenceBits // 7+3+12=22

	// 起始时间戳 (2025-01-01 00:00:00的毫秒级)，一旦固定不能修改
	epoch = int64(1735689600000)
)

type SnowFlake struct {
	workerID      int64
	sequence      int64
	sequenceID    int64
	lastTimeStamp int64
	mu            sync.Mutex // 互斥锁
}

func NewSnowFlake(workerID int64) (*SnowFlake, error) {
	if workerID < 0 || workerID > maxWorkID {
		return nil, errors.New("机器码超出限制范围！")
	}
	return &SnowFlake{
		workerID:      workerID,
		sequence:      0,
		sequenceID:    0,
		lastTimeStamp: -1,
		mu:            sync.Mutex{},
	}, nil
}
