package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Gray   = "\033[38;5;245m"
	Dim    = "\033[38;5;237m"
	White  = "\033[97m"
	Blue   = "\033[38;5;75m"
	Green  = "\033[38;5;78m"
	GreenB = "\033[38;5;29m"
	Orange = "\033[38;5;214m"
	Red    = "\033[38;5;203m"
)

type Config struct {
	URL           string
	TotalRequests int
	Concurrency   int
	Timeout       time.Duration
}

type Result struct {
	StatusCode int
	Duration   time.Duration
	Error      error
}

func sendRequest(client *http.Client, url string) Result {
	start := time.Now()
	resp, err := client.Get(url)
	duration := time.Since(start)
	if err != nil {
		return Result{StatusCode: 0, Duration: duration, Error: err}
	}
	resp.Body.Close()
	return Result{StatusCode: resp.StatusCode, Duration: duration, Error: nil}
}

func worker(client *http.Client, url string, jobs <-chan int, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for range jobs {
		results <- sendRequest(client, url)
	}
}

func makeBar(count int, total int, color string) string {
	const width = 50
	filled := (count * width) / total
	if count > 0 && filled == 0 {
		filled = 1
	}
	empty := width - filled
	return color + strings.Repeat("█", filled) + Dim + strings.Repeat("█", empty) + Reset
}
func printProgress(current, total int) {
	const width = 40
	filled := (current * width) / total
	empty := width - filled
	pct := (float64(current) / float64(total)) * 100

	bar := Green + strings.Repeat("█", filled) + Dim + strings.Repeat("█", empty) + Reset
	fmt.Printf("\r %s⚡ Exécution en cours...%s %s %s%6.1f%%%s %s(%d/%d)%s",
		Bold+Blue, Reset, bar, Bold+White, pct, Reset, Gray, current, total, Reset)
}

func printDashboard(cfg Config, totalDur time.Duration, rps float64, minLat, avgLat, maxLat time.Duration, statusCounts map[int]int, errCount int) {
	
	fmt.Printf("\r%-100s\r\n", "")
	
	fmt.Printf(" %s┌────────────────────────────────────────────┐%s\n", Blue, Reset)
	fmt.Printf(" %s│%s %s%-42s%s %s│%s\n", Blue, Reset, Bold+White, "GOSURGE v1.0 - Testeur de Charge HTTP en Go", Reset, Blue, Reset)
	fmt.Printf(" %s└────────────────────────────────────────────┘%s\n\n", Blue, Reset)

	goroutinesStr := fmt.Sprintf("%d goroutines", cfg.Concurrency)
	timeoutStr := cfg.Timeout.String()

	fmt.Printf(" %s╭──────────────────────────────────────────────────────────────────────────────────────╮%s\n", Gray, Reset)
	fmt.Printf(" %s│%s  %s%-28s %-18s %-20s %-14s%s %s│%s\n",
		Gray, Reset, Gray, "CIBLE URL", "REQUÊTES (-N)", "WORKERS (-C)", "TIMEOUT (-T)", Reset, Gray, Reset)
	fmt.Printf(" %s│%s  %s%-28s%s %s%-18d%s %s%-20s%s %s%-14s%s %s│%s\n",
		Gray, Reset,
		Bold+White, cfg.URL, Reset,
		Bold+Blue, cfg.TotalRequests, Reset,
		Bold+Blue, goroutinesStr, Reset,
		Bold+Blue, timeoutStr, Reset,
		Gray, Reset)
	fmt.Printf(" %s╰──────────────────────────────────────────────────────────────────────────────────────╯%s\n\n", Gray, Reset)

	var durStr string
	if totalDur < time.Second {
		durStr = fmt.Sprintf("%.1f ms", float64(totalDur.Microseconds())/1000.0)
	} else {
		durStr = fmt.Sprintf("%.2f s", totalDur.Seconds())
	}

	rpsStr := fmt.Sprintf("%.2f", rps)
	avgStr := fmt.Sprintf("%.1f ms", float64(avgLat.Microseconds())/1000.0)
	minMaxStr := fmt.Sprintf("Min %.1fms | Max %.1fms",
		float64(minLat.Microseconds())/1000.0,
		float64(maxLat.Microseconds())/1000.0)

	fmt.Printf(" %s╭──────────────────────────╮%s  %s╭──────────────────────────╮%s  %s╭──────────────────────────╮%s\n",
		GreenB, Reset, GreenB, Reset, Gray, Reset)

	fmt.Printf(" %s│%s %s%-24s%s %s│%s  %s│%s %s%-24s%s %s│%s  %s│%s %s%-24s%s %s│%s\n",
		GreenB, Reset, Gray, "Temps Total", Reset, GreenB, Reset,
		GreenB, Reset, Gray, "Débit RPS", Reset, GreenB, Reset,
		Gray, Reset, Gray, "Latence Moyenne", Reset, Gray, Reset)

	fmt.Printf(" %s│%s %s%-24s%s %s│%s  %s│%s %s%-24s%s %s│%s  %s│%s %s%-24s%s %s│%s\n",
		GreenB, Reset, Bold+Green, durStr, Reset, GreenB, Reset,
		GreenB, Reset, Bold+Blue, rpsStr, Reset, GreenB, Reset,
		Gray, Reset, Bold+Orange, avgStr, Reset, Gray, Reset)

	fmt.Printf(" %s│%s %s%-24s%s %s│%s  %s│%s %s%-24s%s %s│%s  %s│%s %s%-24s%s %s│%s\n",
		GreenB, Reset, Gray, "Durée d'exécution", Reset, GreenB, Reset,
		GreenB, Reset, Gray, "req / seconde", Reset, GreenB, Reset,
		Gray, Reset, Gray, minMaxStr, Reset, Gray, Reset)

	fmt.Printf(" %s╰──────────────────────────╯%s  %s╰──────────────────────────╯%s  %s╰──────────────────────────╯%s\n\n",
		GreenB, Reset, GreenB, Reset, Gray, Reset)


	fmt.Printf(" %sRÉPARTITION DES RÉPONSES HTTP%s\n\n", Gray, Reset)

	c200 := statusCounts[200]
	pct200 := (float64(c200) / float64(cfg.TotalRequests)) * 100
	fmt.Printf(" %s%-11s%s %s  %s%5d (%5.1f%%)%s\n\n",
		Green, "[200 OK]", Reset, makeBar(c200, cfg.TotalRequests, Green), Bold+White, c200, pct200, Reset)

	c500 := statusCounts[500]
	pct500 := (float64(c500) / float64(cfg.TotalRequests)) * 100
	fmt.Printf(" %s%-11s%s %s  %s%5d (%5.1f%%)%s\n\n",
		Orange, "[500 Err]", Reset, makeBar(c500, cfg.TotalRequests, Orange), Bold+White, c500, pct500, Reset)

	pctErr := (float64(errCount) / float64(cfg.TotalRequests)) * 100
	fmt.Printf(" %s%-11s%s %s  %s%5d (%5.1f%%)%s\n\n",
		Red, "[Timeout]", Reset, makeBar(errCount, cfg.TotalRequests, Red), Bold+White, errCount, pctErr, Reset)
}
func main() {

	var cfg Config
	var wg sync.WaitGroup

	rootCmd := &cobra.Command{
		Use:   "gosurge",
		Short: "Tool for load testing HTTP/HTTPS requests",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if cfg.TotalRequests <= 0 {
				return fmt.Errorf("total requests must be greater than zero")
			}
			if cfg.Concurrency <= 0 || cfg.Concurrency > cfg.TotalRequests {
				return fmt.Errorf("concurrency must be greater than zero or must not be greater than total requests")
			}
			if cfg.Timeout < 50*time.Millisecond {
				return fmt.Errorf("timeout must be at least 50ms")
			}
			if !(strings.HasPrefix(cfg.URL, "https://") || strings.HasPrefix(cfg.URL, "http://")) {
				return fmt.Errorf("URL must start with 'https://' or 'http://'")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			client := &http.Client{Timeout: cfg.Timeout}
			jobs := make(chan int, cfg.TotalRequests)
			results := make(chan Result, cfg.TotalRequests)

			startTest := time.Now()
			for w := 1; w <= cfg.Concurrency; w++ {
				wg.Add(1)
				go worker(client, cfg.URL, jobs, results, &wg)
			}

			for i := 1; i <= cfg.TotalRequests; i++ {
				jobs <- i
			}

			close(jobs)
			
			go func() {
				wg.Wait()
				close(results)
			}()

			statusCounts := make(map[int]int)

			var errorCount int

			var minLatency time.Duration
			var maxLatency time.Duration
			var sumLatency time.Duration
			
			completed := 0
			for res := range results {
				completed++
				printProgress(completed, cfg.TotalRequests)
				sumLatency += res.Duration
				if res.Duration > maxLatency {
					maxLatency = res.Duration
				}
				if minLatency == 0 || res.Duration < minLatency {
					minLatency = res.Duration
				}
				if res.Error != nil {
					errorCount++
				}
				statusCounts[res.StatusCode]++
			}
			
			totalTestDuration := time.Since(startTest)
			rps := float64(cfg.TotalRequests) / totalTestDuration.Seconds()
			avgLatency := sumLatency / time.Duration(cfg.TotalRequests)

			printDashboard(cfg, totalTestDuration, rps, minLatency, avgLatency, maxLatency, statusCounts, errorCount)
		},
	}

	rootCmd.Flags().StringVarP(&cfg.URL, "url", "u", "", "Target URL")
	rootCmd.Flags().IntVarP(&cfg.TotalRequests, "requests", "n", 100, "Number of requests")
	rootCmd.Flags().IntVarP(&cfg.Concurrency, "concurrency", "c", 10, "Number of concurrent requests")
	rootCmd.Flags().DurationVarP(&cfg.Timeout, "timeout", "t", 5*time.Second, "Timeout")
	rootCmd.MarkFlagRequired("url")

	rootCmd.Execute()
}
