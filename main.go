package raft

import "sync"

// 1.实现3节点选举
// 2.改造代码分布式选举代码

// 定义三节点敞亮
const raftCount = 3

// leader
type Leader struct {
	// 任期
	Term int
	// LeaderId 编号
	LeaderId int
}

type Raft struct {
	mu              sync.Mutex // 锁
	me              int        // 节点编号
	currentTerm     int        // 当前任期
	voteFor         int        // 为谁投票
	state           int        // 状态 0:follower 1:candidate 2:leader
	lastMessageTime int64      // 发送最后一条数据的时间
	currentLeader   int        // 当前节点领导
	message         chan bool  // 节点间发信息的通道
	electCh         chan bool  // 选举通道
	heartBeat       chan bool  // 心跳信号的通道
	heartbeatTr     chan bool  // 返回心跳信号的头痛到
	timeout         int        // 超时时间
}
