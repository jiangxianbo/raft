package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/rpc"
	"sync"
	"time"
)

// 1.实现3节点选举
// 2.改造代码分布式选举代码

// 定义三节点敞亮
const raftCount = 3

// Leader leader对象
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
	heartbeatRe     chan bool  // 返回心跳信号的通道
	timeout         int        // 超时时间
}

// 0 还没上任 -1 没有编号
var leader = Leader{0, -1}

func main() {
	// 过程，有三个节点，最初都是follower
	// 若有candidate, 进行投票和拉票
	// 产生leader

	// 创建三个节点
	for i := 0; i < raftCount; i++ {
		// 创建3个raft节点
		Make(i)
	}

	// 加入rpc服务端 监听
	rpc.Register(new(Raft))
	rpc.HandleHTTP()
	// 监听
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}

	for {

	}
}

func Make(me int) *Raft {
	rf := &Raft{}
	rf.me = me
	rf.voteFor = -1 // -1代表谁都不投，此时节点刚创建
	rf.state = 0
	rf.timeout = 0
	rf.currentLeader = -1
	// 节点任期
	rf.setTerm(0)
	// 初始化通道
	rf.message = make(chan bool)
	rf.electCh = make(chan bool)
	rf.heartBeat = make(chan bool)
	rf.heartbeatRe = make(chan bool)
	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	// 选举协程
	go rf.election()

	// 心跳检查协程
	go rf.sendLeaderHeartBeat()

	return rf
}

func (rf *Raft) setTerm(term int) {
	rf.currentTerm = term
}

// election 选举
func (rf *Raft) election() {
	// 设置标记，判断是否选出了leader
	var result bool
	for {
		// 设置超时
		timeout := randRange(159, 300)
		rf.lastMessageTime = millisecond()
		select {
		// 延迟等待1毫秒
		case <-time.After(time.Duration(timeout) * time.Millisecond):
			fmt.Println("当前节点状态为：", rf.state)
		}
		result = false
		for !result {
			// 选主逻辑
			result = rf.electionOneRound(&leader)
		}
	}
}

// 随机值
func randRange(min, max int64) int64 {
	return rand.Int63n(max-min) + min
}

func millisecond() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// 实现选主的逻辑
func (rf *Raft) electionOneRound(leader *Leader) bool {
	// 定义超时
	var timeout int64
	timeout = 100
	// 投票数量
	var vote int
	// 定义是否开始心跳信号的检查
	var triggerHeartbeat bool
	// 时间
	last := millisecond()
	// 用于返回值
	success := false

	// 给当前节点变成candidate
	rf.mu.Lock()
	// 修改状态
	rf.becomeCandidate()
	rf.mu.Unlock()
	fmt.Println("start electing leader")
	for {
		// 遍历素有节点拉选票
		for i := 0; i < raftCount; i++ {
			if i != rf.me {
				// 拉选票
				go func() {
					if leader.LeaderId < 0 {
						// 设置投票
						rf.electCh <- true
					}
				}()
			}
		}
		// 设置投票数量
		vote = 1
		// 遍历节点
		for i := 0; i < raftCount; i++ {
			// 计算投票统计
			select {
			case ok := <-rf.electCh:
				if ok {
					// 投票数量+1
					vote++
					// 选票格式，大于节点个数 / 2，则成功
					success = vote > raftCount/2
					if success && !triggerHeartbeat {
						// 变成主节点，选主成功
						// 开始触发心跳检测
						triggerHeartbeat = true
						// 变主
						rf.mu.Lock()
						rf.becomeLeader()
						rf.mu.Unlock()
						// 由leader向其他节点发送心跳信号
						rf.heartBeat <- true
						fmt.Println(rf.me, "号节点成为了leader")
						fmt.Println("leader开始发送心跳信号了")
					}
				}
			}
		}
		// 做最后校验工作
		// 若不超时，且票数大于一般，则选举成功，break
		if timeout+last < millisecond() || vote > raftCount/2 || rf.currentLeader > -1 {
			break
		} else {
			// 等待操作
			select {
			case <-time.After(time.Duration(10) * time.Millisecond):
			}
		}
	}
	return success
}

// 修改状态candidate
func (rf *Raft) becomeCandidate() {
	rf.state = 1
	rf.setTerm(rf.currentTerm + 1)
	rf.voteFor = rf.me
	rf.currentLeader = -1
}

// 成为
func (rf *Raft) becomeLeader() {
	rf.state = 2
	rf.currentLeader = rf.me
}

// leader节点发送心跳信号
// 顺便完成数据同步 -- 这里先不实现
// 产看小弟挂没挂
func (rf *Raft) sendLeaderHeartBeat() {
	for {
		select {
		case <-rf.heartBeat:
			rf.sendAppendEntriesTmpl()
		}
	}
}

// 用于返回给leader的确认信号
func (rf *Raft) sendAppendEntriesTmpl() {
	// 此时leader不用给自己返回信号
	if rf.currentLeader == rf.me {
		// 确定信号的节点个数
		var success_count = 0
		// 设置确认信号
		for i := 0; i < raftCount; i++ {
			if i != rf.me {
				go func() {
					//rf.heartbeatRe <- true

					rp, err := rpc.DialHTTP("tcp", "127.0.0.1:8080") // rpc方式：此处相当于客户端
					if err != nil {
						log.Fatal(err)
					}
					// 接受服务器返回信息
					// 接受服务端返回信息变量
					var ok = false
					err = rp.Call("Raft.Communication", Param{Msg: "hello"}, &ok)
					if err != nil {
						log.Fatal(err)
					}
					if ok {
						rf.heartbeatRe <- true
					}
				}()
			}
		}
		// 计算返回确认信号个数
		for i := 0; i < raftCount; i++ {
			select {
			case ok := <-rf.heartbeatRe:
				if ok {
					success_count++
					if success_count > raftCount/2 {
						fmt.Println("投票选举成功，心跳信号OK")
						log.Fatal("程序结束")
					}
				}
			}
		}
	}
}

// 首字母大写，RPC规范
// 分布式同心
type Param struct {
	Msg string
}

func (rf Raft) Communication(p Param, a *bool) error {
	fmt.Println(p.Msg)
	*a = true
	return nil
}
