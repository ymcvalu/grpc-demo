package pool

import (
	"google.golang.org/grpc"
	"sync"
	"time"
)

type Options struct {
	Dial        Dialer
	MaxConn     int
	MaxIdle     int
	WaitTimeout time.Duration
}

type Dialer func() (*grpc.ClientConn, error)

type Pool struct {
	dial    Dialer
	maxConn int
	maxIdle int

	waitTimeout time.Duration
	connCh      chan *grpc.ClientConn

	curConnNum int
	freeConn   []*grpc.ClientConn
	sync.Mutex
}

func NewPool(opts Options) *Pool {
	if opts.MaxConn <= 0 {
		opts.MaxConn = 10
	}
	if opts.MaxIdle <= 0 {
		opts.MaxIdle = 5
	}
	if opts.MaxIdle > opts.MaxConn {
		opts.MaxIdle = opts.MaxIdle
	}

	return &Pool{
		dial:        opts.Dial,
		maxConn:     opts.MaxConn,
		maxIdle:     opts.MaxIdle,
		waitTimeout: opts.WaitTimeout,
		connCh:      make(chan *grpc.ClientConn),
		freeConn:    make([]*grpc.ClientConn, 0, opts.MaxIdle),
	}

}

func (p *Pool) Get() (conn *grpc.ClientConn) {
	p.Lock()
	// 已经到达最大连接数
	if p.curConnNum >= p.maxConn {
		if p.waitTimeout == 0 {
			p.Unlock()
			return
		}

		var tm <-chan time.Time
		if p.waitTimeout > 0 {
			tm = time.After(p.waitTimeout)
		}
		p.Unlock()
		select {
		case <-tm:
		case conn = <-p.connCh:
		}
		return
	}

	if ln := len(p.freeConn); ln > 0 {
		conn = p.freeConn[0]
		p.freeConn[0] = p.freeConn[ln-1]
		p.freeConn = p.freeConn[:ln-1]
	} else {
		c, err := p.dial()
		if err != nil {
			conn = nil
		} else {
			p.curConnNum++
			conn = c
		}
	}
	p.Unlock()
	return
}

func (p *Pool) Put(conn *grpc.ClientConn) error {
	if conn == nil {
		return nil
	}
	select {
	case p.connCh <- conn:
		return nil
	default:
	}
	p.Lock()
	defer p.Unlock()
	if len(p.freeConn) < p.maxIdle {
		p.freeConn = append(p.freeConn, conn)
		return nil
	}
	select {
	case p.connCh <- conn:
		return nil
	default:
		p.curConnNum--
		return conn.Close()
	}
}

func (p *Pool) Stat() PoolStat {
	p.Lock()
	p.Unlock()
	return PoolStat{
		ConnNum:     p.curConnNum,
		IdleConnNum: len(p.freeConn),
	}
}

type PoolStat struct {
	ConnNum     int
	IdleConnNum int
}
