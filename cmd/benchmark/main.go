package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type BenchmarkConfig struct {
	Host        string
	Port        int
	Clients     int
	Requests    int
	Tests       []string
	ValueSize   int
	OutputMode  string
}

type TestResult struct {
	Name        string        `json:"test_name"`
	TotalReqs   int           `json:"total_requests"`
	Concurrency int           `json:"concurrency"`
	Duration    time.Duration `json:"duration"`
	RPS         float64       `json:"requests_per_sec"`
	MinLatency  time.Duration `json:"min_latency"`
	MeanLatency time.Duration `json:"mean_latency"`
	P50Latency  time.Duration `json:"p50_latency"`
	P90Latency  time.Duration `json:"p90_latency"`
	P95Latency  time.Duration `json:"p95_latency"`
	P99Latency  time.Duration `json:"p99_latency"`
	MaxLatency  time.Duration `json:"max_latency"`
	Errors      int           `json:"error_count"`
}

func readRESPResponse(r *bufio.Reader) error {
	line, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	if len(line) < 3 {
		return fmt.Errorf("malformed response: %q", line)
	}
	prefix := line[0]
	switch prefix {
	case '+', '-', ':':
		return nil
	case '$':
		// Bulk string
		lenStr := strings.TrimSpace(line[1:])
		length, err := strconv.Atoi(lenStr)
		if err != nil {
			return err
		}
		if length == -1 {
			return nil
		}
		buf := make([]byte, length+2) // data + \r\n
		_, err = io.ReadFull(r, buf)
		return err
	case '*':
		// Array
		countStr := strings.TrimSpace(line[1:])
		count, err := strconv.Atoi(countStr)
		if err != nil {
			return err
		}
		if count == -1 {
			return nil
		}
		for i := 0; i < count; i++ {
			if err := readRESPResponse(r); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown RESP prefix: %c in %q", prefix, line)
	}
}

func formatRESPCommand(args ...string) []byte {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*%d\r\n", len(args)))
	for _, arg := range args {
		sb.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg))
	}
	return []byte(sb.String())
}

func runBenchmarkTest(cfg BenchmarkConfig, testName string) TestResult {
	reqPerClient := cfg.Requests / cfg.Clients
	totalRequests := reqPerClient * cfg.Clients

	latencies := make([]time.Duration, totalRequests)
	var errCount int64

	valPayload := strings.Repeat("x", cfg.ValueSize)

	var wg sync.WaitGroup
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	startTotal := time.Now()

	for c := 0; c < cfg.Clients; c++ {
		wg.Add(1)
		clientIdx := c
		go func(workerID int) {
			defer wg.Done()

			conn, err := net.Dial("tcp", addr)
			if err != nil {
				for i := 0; i < reqPerClient; i++ {
					idx := workerID*reqPerClient + i
					latencies[idx] = 0
				}
				syncAtomicAdd(&errCount, int64(reqPerClient))
				return
			}
			defer conn.Close()

			reader := bufio.NewReader(conn)
			offset := workerID * reqPerClient

			for i := 0; i < reqPerClient; i++ {
				var cmdBytes []byte
				key := fmt.Sprintf("bench:%d:%d", workerID, i%1000)

				switch testName {
				case "PING":
					cmdBytes = formatRESPCommand("PING")
				case "SET":
					cmdBytes = formatRESPCommand("SET", key, valPayload)
				case "GET":
					cmdBytes = formatRESPCommand("GET", key)
				case "INCR":
					cmdBytes = formatRESPCommand("INCR", fmt.Sprintf("bench:counter:%d", workerID))
				case "RPUSH":
					cmdBytes = formatRESPCommand("RPUSH", fmt.Sprintf("bench:list:%d", workerID), valPayload)
				case "LRANGE":
					cmdBytes = formatRESPCommand("LRANGE", fmt.Sprintf("bench:list:%d", workerID), "0", "10")
				case "MIXED":
					if i%2 == 0 {
						cmdBytes = formatRESPCommand("GET", key)
					} else {
						cmdBytes = formatRESPCommand("SET", key, valPayload)
					}
				case "TX":
					t0 := time.Now()
					// Pipeline MULTI, SET, INCR, EXEC
					if _, err := conn.Write(formatRESPCommand("MULTI")); err != nil {
						syncAtomicAdd(&errCount, 1)
						continue
					}
					if err := readRESPResponse(reader); err != nil {
						syncAtomicAdd(&errCount, 1)
						continue
					}
					if _, err := conn.Write(formatRESPCommand("SET", key, valPayload)); err != nil {
						syncAtomicAdd(&errCount, 1)
						continue
					}
					if err := readRESPResponse(reader); err != nil {
						syncAtomicAdd(&errCount, 1)
						continue
					}
					if _, err := conn.Write(formatRESPCommand("INCR", fmt.Sprintf("bench:tx:cnt:%d", workerID))); err != nil {
						syncAtomicAdd(&errCount, 1)
						continue
					}
					if err := readRESPResponse(reader); err != nil {
						syncAtomicAdd(&errCount, 1)
						continue
					}
					if _, err := conn.Write(formatRESPCommand("EXEC")); err != nil {
						syncAtomicAdd(&errCount, 1)
						continue
					}
					if err := readRESPResponse(reader); err != nil {
						syncAtomicAdd(&errCount, 1)
						continue
					}
					lat := time.Since(t0)
					latencies[offset+i] = lat
					continue
				default:
					cmdBytes = formatRESPCommand("PING")
				}

				t0 := time.Now()
				if _, err := conn.Write(cmdBytes); err != nil {
					syncAtomicAdd(&errCount, 1)
					continue
				}
				if err := readRESPResponse(reader); err != nil {
					syncAtomicAdd(&errCount, 1)
					continue
				}
				lat := time.Since(t0)
				latencies[offset+i] = lat
			}
		}(clientIdx)
	}

	wg.Wait()
	duration := time.Since(startTotal)

	// Filter valid latencies
	validLats := make([]time.Duration, 0, totalRequests)
	var sum time.Duration
	for _, l := range latencies {
		if l > 0 {
			validLats = append(validLats, l)
			sum += l
		}
	}

	res := TestResult{
		Name:        testName,
		TotalReqs:   totalRequests,
		Concurrency: cfg.Clients,
		Duration:    duration,
		Errors:      int(errCount),
	}

	if len(validLats) > 0 {
		sort.Slice(validLats, func(i, j int) bool { return validLats[i] < validLats[j] })
		res.RPS = float64(len(validLats)) / duration.Seconds()
		res.MinLatency = validLats[0]
		res.MaxLatency = validLats[len(validLats)-1]
		res.MeanLatency = time.Duration(int64(sum) / int64(len(validLats)))
		res.P50Latency = percentile(validLats, 50)
		res.P90Latency = percentile(validLats, 90)
		res.P95Latency = percentile(validLats, 95)
		res.P99Latency = percentile(validLats, 99)
	}

	return res
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(float64(len(sorted))*(p/100.0))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

var atomicMu sync.Mutex

func syncAtomicAdd(val *int64, delta int64) {
	atomicMu.Lock()
	*val += delta
	atomicMu.Unlock()
}

func warmUp(cfg BenchmarkConfig) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	// Seed some keys for GET / LRANGE
	for i := 0; i < 500; i++ {
		k := fmt.Sprintf("bench:warmup:%d", i)
		conn.Write(formatRESPCommand("SET", k, "warmup_value"))
		_ = readRESPResponse(reader)
		conn.Write(formatRESPCommand("RPUSH", fmt.Sprintf("bench:list:warmup"), "elem"))
		_ = readRESPResponse(reader)
	}
}

func main() {
	host := flag.String("host", "127.0.0.1", "Server host")
	port := flag.Int("port", 6379, "Server port")
	clients := flag.Int("c", 50, "Number of concurrent connections")
	requests := flag.Int("n", 50000, "Total number of requests")
	testsStr := flag.String("t", "PING,SET,GET,INCR,RPUSH,LRANGE,MIXED,TX", "Comma-separated list of tests")
	valSize := flag.Int("d", 32, "Data size in bytes for SET/RPUSH values")
	outputMode := flag.String("format", "text", "Output format: text, markdown")
	flag.Parse()

	var tests []string
	for _, t := range strings.Split(*testsStr, ",") {
		tt := strings.TrimSpace(t)
		if tt != "" {
			tests = append(tests, strings.ToUpper(tt))
		}
	}

	cfg := BenchmarkConfig{
		Host:       *host,
		Port:       *port,
		Clients:    *clients,
		Requests:   *requests,
		Tests:      tests,
		ValueSize:  *valSize,
		OutputMode: *outputMode,
	}

	fmt.Printf("=== Redish Benchmark Suite ===\n")
	fmt.Printf("Target:       %s:%d\n", cfg.Host, cfg.Port)
	fmt.Printf("Concurrency:  %d parallel clients\n", cfg.Clients)
	fmt.Printf("Requests:     %d per benchmark\n", cfg.Requests)
	fmt.Printf("Payload size: %d bytes\n\n", cfg.ValueSize)

	warmUp(cfg)

	var results []TestResult
	for _, t := range cfg.Tests {
		fmt.Printf("Running benchmark: %-8s ... ", t)
		res := runBenchmarkTest(cfg, t)
		results = append(results, res)
		fmt.Printf("DONE (%.2f req/s, p50: %v, p99: %v)\n", res.RPS, res.P50Latency, res.P99Latency)
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("BENCHMARK RESULTS")
	fmt.Println(strings.Repeat("=", 80))

	if cfg.OutputMode == "markdown" {
		printMarkdownTable(results)
	} else {
		printTextTable(results)
	}
}

func printTextTable(results []TestResult) {
	fmt.Printf("%-10s | %-12s | %-10s | %-10s | %-10s | %-10s | %-8s\n",
		"Test", "Throughput", "Avg Lat", "p50 Lat", "p95 Lat", "p99 Lat", "Errors")
	fmt.Println(strings.Repeat("-", 80))
	for _, r := range results {
		fmt.Printf("%-10s | %10.1f/s | %10v | %10v | %10v | %10v | %8d\n",
			r.Name, r.RPS, r.MeanLatency.Round(time.Microsecond),
			r.P50Latency.Round(time.Microsecond),
			r.P95Latency.Round(time.Microsecond),
			r.P99Latency.Round(time.Microsecond),
			r.Errors)
	}
}

func printMarkdownTable(results []TestResult) {
	fmt.Println("| Command | Concurrency | Throughput (Req/sec) | Avg Latency | p50 (Median) | p95 Latency | p99 Latency | Error Rate |")
	fmt.Println("| :--- | :---: | :---: | :---: | :---: | :---: | :---: | :---: |")
	for _, r := range results {
		errRate := (float64(r.Errors) / float64(r.TotalReqs)) * 100.0
		fmt.Printf("| `%s` | %d | **%.1f** | %v | %v | %v | %v | %.2f%% |\n",
			r.Name, r.Concurrency, r.RPS,
			r.MeanLatency.Round(time.Microsecond),
			r.P50Latency.Round(time.Microsecond),
			r.P95Latency.Round(time.Microsecond),
			r.P99Latency.Round(time.Microsecond),
			errRate)
	}
}
