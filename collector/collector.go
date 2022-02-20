package collector

import (
	"context"
	"github.com/adshao/go-binance/v2"
	ftxapi "github.com/aibotsoft/ftx-api"
	"github.com/aibotsoft/latency-metric/config"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Collector struct {
	cfg       *config.Config
	log       *zap.Logger
	ctx       context.Context
	errCh     chan error
	ftxClient *ftxapi.Client
	binClient *binance.Client
}

func NewCollector(cfg *config.Config, log *zap.Logger, ctx context.Context) *Collector {
	ftxConfig := ftxapi.Config{
		ApiKey:    cfg.Ftx.Key,
		ApiSecret: cfg.Ftx.Secret,
		Logger:    log.WithOptions(zap.IncreaseLevel(zap.InfoLevel)).Sugar(),
	}
	ftxClient := ftxapi.NewClient(ftxConfig)
	binClient := binance.NewClient(cfg.Binance.Key, cfg.Binance.Secret)
	return &Collector{
		cfg:       cfg,
		log:       log,
		ctx:       ctx,
		errCh:     make(chan error),
		ftxClient: ftxClient,
		binClient: binClient,
		//binanceClient:         binance_ws.NewWsClient(cfg, log, ctx),
		//binanceOtherSymbolMap: make(map[string]string),
	}
}
func (c *Collector) BinMetric(wg *sync.WaitGroup) {
	defer wg.Done()
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	resp, err := c.binClient.NewGetAccountService().Do(ctx)
	cancel()
	if err != nil {
		c.log.Error("warmup_request_error", zap.Error(err), zap.Any("resp", resp))
	}
	var latencyList []time.Duration
	ts := time.Now()
	for i := 0; i < c.cfg.Count; i++ {
		start := time.Now()
		ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
		_, err := c.binClient.NewGetAccountService().Do(ctx)
		cancel()
		if err != nil {
			c.log.Error("get_account_error", zap.Error(err))
			continue
		}
		latencyList = append(latencyList, time.Since(start))
	}
	c.log.Info("bin",
		zap.Int("config_count", c.cfg.Count),
		zap.Int("count", len(latencyList)),
		zap.Duration("avg_latency", avgLatency(latencyList)),
		zap.Duration("min_latency", minLatency(latencyList)),
		zap.Duration("max_latency", maxLatency(latencyList)),
		zap.Duration("total_time", time.Since(ts)),
	)
	c.log.Debug("latency_list", zap.Any("list", latencyList))
}
func (c *Collector) FtxMetric(wg *sync.WaitGroup) {
	defer wg.Done()
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	resp, err := c.ftxClient.NewGetAccountService().Do(ctx)
	cancel()
	if err != nil {
		c.log.Error("warmup_request_error", zap.Error(err), zap.Any("resp", resp))
	}
	var latencyList []time.Duration
	ts := time.Now()
	for i := 0; i < c.cfg.Count; i++ {
		start := time.Now()
		ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
		_, err := c.ftxClient.NewGetAccountService().Do(ctx)
		cancel()
		if err != nil {
			c.log.Error("get_account_error", zap.Error(err))
			continue
		}
		latencyList = append(latencyList, time.Since(start))
	}
	c.log.Info("ftx",
		zap.Int("config_count", c.cfg.Count),
		zap.Int("count", len(latencyList)),
		zap.Duration("avg_latency", avgLatency(latencyList)),
		zap.Duration("min_latency", minLatency(latencyList)),
		zap.Duration("max_latency", maxLatency(latencyList)),
		zap.Duration("total_time", time.Since(ts)),
	)
	c.log.Debug("latency_list", zap.Any("list", latencyList))
}
func (c *Collector) Run() error {
	//statTick := time.Tick(time.Minute * 1)
	var wg sync.WaitGroup
	wg.Add(2)
	go c.FtxMetric(&wg)
	go c.BinMetric(&wg)
	wg.Wait()
	return nil
	//c.BinMetric()
	//c.BinMetric()
	//c.BinMetric()
	//go func() {
	//	c.errCh <- c.binanceClient.Serve()
	//}()
	//go func() {
	//	c.errCh <- c.ftxClient.Serve()
	//}()
	//var binanceCount, ftxCount int
	//for {
	//	select {
	//	//case err := <-c.binanceClient.ReadErr:
	//	//	c.log.Warn("binance_read_error", zap.Error(err))
	//	//case event := <-c.binanceClient.EventCh:
	//	//	binanceCount++
	//	//	c.lastBinTime = event.ReceiveTime
	//	//	c.binanceHandler(&event)
	//	//case event := <-c.ftxClient.EventCh:
	//	//	ftxCount++
	//	//	c.fxtHandler(&event)
	//	case <-c.ctx.Done():
	//		return c.ctx.Err()
	//	case err := <-c.errCh:
	//		return err
	//	case <-statTick:
	//		c.log.Info("stat",
	//			zap.Int("binance", binanceCount),
	//			zap.Int("ftx", ftxCount),
	//		)
	//		binanceCount = 0
	//		ftxCount = 0
	//	}
	//}
}
func avgLatency(list []time.Duration) time.Duration {
	var total time.Duration
	for _, d := range list {
		total = total + d
	}
	return total / time.Duration(len(list))
}
func minLatency(list []time.Duration) time.Duration {
	var min = time.Hour
	for _, d := range list {
		if d < min {
			min = d
		}
	}
	return min
}
func maxLatency(list []time.Duration) time.Duration {
	var max time.Duration
	for _, d := range list {
		if d > max {
			max = d
		}
	}
	return max
}
