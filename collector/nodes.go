package collector

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	defaultNodeLabels       = []string{"cluster", "host", "name"}
	defaultThreadPoolLabels = append(defaultNodeLabels, "type")
	defaultBreakerLabels    = append(defaultNodeLabels, "breaker")
	defaultFilesystemLabels = append(defaultNodeLabels, "mount", "path")

	defaultNodeLabelValues = func(cluster string, node NodeStatsNodeResponse) []string {
		return []string{cluster, node.Host, node.Name}
	}
	defaultThreadPoolLabelValues = func(cluster string, node NodeStatsNodeResponse, pool string) []string {
		return append(defaultNodeLabelValues(cluster, node), pool)
	}
	defaultFilesystemLabelValues = func(cluster string, node NodeStatsNodeResponse, mount string, path string) []string {
		return append(defaultNodeLabelValues(cluster, node), mount, path)
	}
)

type nodeMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(node NodeStatsNodeResponse) float64
	Labels func(cluster string, node NodeStatsNodeResponse) []string
}

type gcCollectionMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(gcStats NodeStatsJVMGCCollectorResponse) float64
	Labels func(cluster string, node NodeStatsNodeResponse, collector string) []string
}

type breakerMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(breakerStats NodeStatsBreakersResponse) float64
	Labels func(cluster string, node NodeStatsNodeResponse, breaker string) []string
}

type threadPoolMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(threadPoolStats NodeStatsThreadPoolPoolResponse) float64
	Labels func(cluster string, node NodeStatsNodeResponse, breaker string) []string
}

type filesystemMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(fsStats NodeStatsFSDataResponse) float64
	Labels func(cluster string, node NodeStatsNodeResponse, mount string, path string) []string
}

type Nodes struct {
	logger log.Logger
	client *http.Client
	url    url.URL
	all    bool

	nodeMetrics         []*nodeMetric
	gcCollectionMetrics []*gcCollectionMetric
	breakerMetrics      []*breakerMetric
	threadPoolMetrics   []*threadPoolMetric
	filesystemMetrics   []*filesystemMetric
}

func NewNodes(logger log.Logger, client *http.Client, url url.URL, all bool) *Nodes {
	return &Nodes{
		logger: logger,
		client: client,
		url:    url,
		all:    all,

		nodeMetrics: []*nodeMetric{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "fielddata_memory_size_bytes"),
					"Field data cache memory usage in bytes",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.FieldData.MemorySize)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "filter_cache_memory_size_bytes"),
					"Filter cache memory usage in bytes",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.FilterCache.MemorySize)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "query_cache_memory_size_bytes"),
					"Query cache memory usage in bytes",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.QueryCache.MemorySize)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "request_cache_memory_size_bytes"),
					"Request cache memory usage in bytes",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.RequestCache.MemorySize)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "docs"),
					"Count of documents on this node",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Docs.Count)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "docs_deleted"),
					"Count of deleted documents on this node",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Docs.Deleted)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "store_size_bytes"),
					"Current size of stored index data in bytes",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return 0
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "segments_memory_bytes"),
					"Current memory size of segments in bytes",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Segments.Memory)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "indices", "segments_count"),
					"Count of index segments on this node",
					defaultNodeLabels, nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.Indices.Segments.Count)
				},
				Labels: defaultNodeLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "jvm_memory", "used_bytes"),
					"JVM memory currently used by area",
					append(defaultNodeLabels, "area"), nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.JVM.Mem.HeapUsed)
				},
				Labels: func(cluster string, node NodeStatsNodeResponse) []string {
					return append(defaultNodeLabelValues(cluster, node), "heap")
				},
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "jvm_memory", "used_bytes"),
					"JVM memory currently used by area",
					append(defaultNodeLabels, "area"), nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.JVM.Mem.NonHeapUsed)
				},
				Labels: func(cluster string, node NodeStatsNodeResponse) []string {
					return append(defaultNodeLabelValues(cluster, node), "non-heap")
				},
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "jvm_memory", "max_bytes"),
					"JVM memory max",
					append(defaultNodeLabels, "area"), nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.JVM.Mem.HeapMax)
				},
				Labels: func(cluster string, node NodeStatsNodeResponse) []string {
					return append(defaultNodeLabelValues(cluster, node), "heap")
				},
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "jvm_memory", "committed_bytes"),
					"JVM memory currently committed by area",
					append(defaultNodeLabels, "area"), nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.JVM.Mem.HeapCommitted)
				},
				Labels: func(cluster string, node NodeStatsNodeResponse) []string {
					return append(defaultNodeLabelValues(cluster, node), "heap")
				},
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "jvm_memory", "committed_bytes"),
					"JVM memory currently committed by area",
					append(defaultNodeLabels, "area"), nil,
				),
				Value: func(node NodeStatsNodeResponse) float64 {
					return float64(node.JVM.Mem.NonHeapCommitted)
				},
				Labels: func(cluster string, node NodeStatsNodeResponse) []string {
					return append(defaultNodeLabelValues(cluster, node), "non-heap")
				},
			},
		},
		gcCollectionMetrics: []*gcCollectionMetric{
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "jvm_gc", "collection_seconds_count"),
					"Count of JVM GC runs",
					append(defaultNodeLabels, "gc"), nil,
				),
				Value: func(gcStats NodeStatsJVMGCCollectorResponse) float64 {
					return float64(gcStats.CollectionCount)
				},
				Labels: func(cluster string, node NodeStatsNodeResponse, collector string) []string {
					return append(defaultNodeLabelValues(cluster, node), collector)
				},
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "jvm_gc", "collection_seconds_sum"),
					"GC run time in seconds",
					append(defaultNodeLabels, "gc"), nil,
				),
				Value: func(gcStats NodeStatsJVMGCCollectorResponse) float64 {
					return float64(gcStats.CollectionTime / 1000)
				},
				Labels: func(cluster string, node NodeStatsNodeResponse, collector string) []string {
					return append(defaultNodeLabelValues(cluster, node), collector)
				},
			},
		},
		breakerMetrics: []*breakerMetric{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "breakers", "estimated_size_bytes"),
					"Estimated size in bytes of breaker",
					defaultBreakerLabels, nil,
				),
				Value: func(breakerStats NodeStatsBreakersResponse) float64 {
					return float64(breakerStats.EstimatedSize)
				},
				Labels: func(cluster string, node NodeStatsNodeResponse, breaker string) []string {
					return append(defaultNodeLabelValues(cluster, node), breaker)
				},
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "breakers", "limit_size_bytes"),
					"Limit size in bytes for breaker",
					defaultBreakerLabels, nil,
				),
				Value: func(breakerStats NodeStatsBreakersResponse) float64 {
					return float64(breakerStats.LimitSize)
				},
				Labels: func(cluster string, node NodeStatsNodeResponse, breaker string) []string {
					return append(defaultNodeLabelValues(cluster, node), breaker)
				},
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "breakers", "tripped"),
					"tripped for breaker",
					defaultBreakerLabels, nil,
				),
				Value: func(breakerStats NodeStatsBreakersResponse) float64 {
					return float64(breakerStats.Tripped)
				},
				Labels: func(cluster string, node NodeStatsNodeResponse, breaker string) []string {
					return append(defaultNodeLabelValues(cluster, node), breaker)
				},
			},
		},
		threadPoolMetrics: []*threadPoolMetric{
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "thread_pool", "completed_count"),
					"Thread Pool operations completed",
					defaultThreadPoolLabels, nil,
				),
				Value: func(threadPoolStats NodeStatsThreadPoolPoolResponse) float64 {
					return float64(threadPoolStats.Completed)
				},
				Labels: defaultThreadPoolLabelValues,
			},
			{
				Type: prometheus.CounterValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "thread_pool", "rejected_count"),
					"Thread Pool operations rejected",
					defaultThreadPoolLabels, nil,
				),
				Value: func(threadPoolStats NodeStatsThreadPoolPoolResponse) float64 {
					return float64(threadPoolStats.Rejected)
				},
				Labels: defaultThreadPoolLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "thread_pool", "active_count"),
					"Thread Pool threads active",
					defaultThreadPoolLabels, nil,
				),
				Value: func(threadPoolStats NodeStatsThreadPoolPoolResponse) float64 {
					return float64(threadPoolStats.Active)
				},
				Labels: defaultThreadPoolLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "thread_pool", "largest_count"),
					"Thread Pool largest threads count",
					defaultThreadPoolLabels, nil,
				),
				Value: func(threadPoolStats NodeStatsThreadPoolPoolResponse) float64 {
					return float64(threadPoolStats.Largest)
				},
				Labels: defaultThreadPoolLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "thread_pool", "queue_count"),
					"Thread Pool operations queued",
					defaultThreadPoolLabels, nil,
				),
				Value: func(threadPoolStats NodeStatsThreadPoolPoolResponse) float64 {
					return float64(threadPoolStats.Queue)
				},
				Labels: defaultThreadPoolLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "thread_pool", "threads_count"),
					"Thread Pool current threads count",
					defaultThreadPoolLabels, nil,
				),
				Value: func(threadPoolStats NodeStatsThreadPoolPoolResponse) float64 {
					return float64(threadPoolStats.Threads)
				},
				Labels: defaultThreadPoolLabelValues,
			},
		},
		filesystemMetrics: []*filesystemMetric{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "filesystem_data", "available_bytes"),
					"Available space on block device in bytes",
					defaultFilesystemLabels, nil,
				),
				Value: func(fsStats NodeStatsFSDataResponse) float64 {
					return float64(fsStats.Available)
				},
				Labels: defaultFilesystemLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "filesystem_data", "free_bytes"),
					"Free space on block device in bytes",
					defaultFilesystemLabels, nil,
				),
				Value: func(fsStats NodeStatsFSDataResponse) float64 {
					return float64(fsStats.Free)
				},
				Labels: defaultFilesystemLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "filesystem_data", "size_bytes"),
					"Size of block device in bytes",
					defaultFilesystemLabels, nil,
				),
				Value: func(fsStats NodeStatsFSDataResponse) float64 {
					return float64(fsStats.Total)
				},
				Labels: defaultFilesystemLabelValues,
			},
		},
	}
}

func (c *Nodes) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.nodeMetrics {
		ch <- metric.Desc
	}
	for _, metric := range c.gcCollectionMetrics {
		ch <- metric.Desc
	}
	for _, metric := range c.threadPoolMetrics {
		ch <- metric.Desc
	}
	for _, metric := range c.filesystemMetrics {
		ch <- metric.Desc
	}
}

func (c *Nodes) Collect(ch chan<- prometheus.Metric) {
	path := "/_nodes/_local/stats"
	if c.all {
		path = "/_nodes/stats"
	}
	c.url.Path = path

	res, err := c.client.Get(c.url.String())
	if err != nil {
		level.Warn(c.logger).Log(
			"msg", "failed to get nodes",
			"url", c.url.String(),
			"err", err,
		)
		return
	}
	defer res.Body.Close()

	dec := json.NewDecoder(res.Body)

	var nodeStatsResponse NodeStatsResponse
	if err := dec.Decode(&nodeStatsResponse); err != nil {
		level.Warn(c.logger).Log(
			"msg", "failed to decode nodes",
			"err", err,
		)
		return
	}

	for _, node := range nodeStatsResponse.Nodes {
		for _, metric := range c.nodeMetrics {
			ch <- prometheus.MustNewConstMetric(
				metric.Desc,
				metric.Type,
				metric.Value(node),
				metric.Labels(nodeStatsResponse.ClusterName, node)...,
			)
		}

		// GC Stats
		for collector, gcStats := range node.JVM.GC.Collectors {
			for _, metric := range c.gcCollectionMetrics {
				ch <- prometheus.MustNewConstMetric(
					metric.Desc,
					metric.Type,
					metric.Value(gcStats),
					metric.Labels(nodeStatsResponse.ClusterName, node, collector)...,
				)
			}
		}

		// Breaker stats
		for breaker, bstats := range node.Breakers {
			for _, metric := range c.breakerMetrics {
				ch <- prometheus.MustNewConstMetric(
					metric.Desc,
					metric.Type,
					metric.Value(bstats),
					metric.Labels(nodeStatsResponse.ClusterName, node, breaker)...,
				)
			}
		}

		// Thread Pool stats
		for pool, pstats := range node.ThreadPool {
			for _, metric := range c.threadPoolMetrics {
				ch <- prometheus.MustNewConstMetric(
					metric.Desc,
					metric.Type,
					metric.Value(pstats),
					metric.Labels(nodeStatsResponse.ClusterName, node, pool)...,
				)
			}
		}

		// File System Stats
		for _, fsStats := range node.FS.Data {
			for _, metric := range c.filesystemMetrics {
				ch <- prometheus.MustNewConstMetric(
					metric.Desc,
					metric.Type,
					metric.Value(fsStats),
					metric.Labels(nodeStatsResponse.ClusterName, node, fsStats.Mount, fsStats.Path)...,
				)
			}
		}
	}
}
