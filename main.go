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

func printDashboard(cfg Config, totalDur time.Duration, rps float64, minLat, avgLat, maxLat time.Duration, statusCounts map[int]int, errCount int) {
	fmt.Println()

	// 1. BANNIÈRE TITRE
	fmt.Printf(" %s┌──────────────────────────────────────────┐%s\n", Blue, Reset)
	fmt.Printf(" %s│%s %sGOSURGE v1.0 – Testeur de Charge HTTP en Go%s %s│%s\n", Blue, Reset, Bold+White, Reset, Blue, Reset)
	fmt.Printf(" %s└──────────────────────────────────────────┘%s\n\n", Blue, Reset)

	// 2. CADRE DE CONFIGURATION (4 COLONNES)
	goroutinesStr := fmt.Sprintf("%d goroutines", cfg.Concurrency)
	fmt.Printf(" %s╭──────────────────────────────────────────────────────────────────────────────────────╮%s\n", Gray, Reset)
	fmt.Printf(" %s│%s  %-24s %-18s %-20s %-18s %s│%s\n",
		Gray, Reset, Gray+"CIBLE URL"+Reset, Gray+"REQUÊTES (-N)"+Reset, Gray+"WORKERS (-C)"+Reset, Gray+"TIMEOUT (-T)"+Reset, Gray, Reset)
	fmt.Printf(" %s│%s  %s%-24s%s %s%-18d%s %s%-20s%s %s%-18v%s %s│%s\n",
		Gray, Reset,
		Bold+White, cfg.URL, Reset,
		Bold+Blue, cfg.TotalRequests, Reset,
		Bold+Blue, goroutinesStr, Reset,
		Bold+Blue, cfg.Timeout, Reset,
		Gray, Reset)
	fmt.Printf(" %s╰──────────────────────────────────────────────────────────────────────────────────────╯%s\n\n", Gray, Reset)

	// Formatage des valeurs pour les 3 boîtes côte à côte
	durStr := fmt.Sprintf("%.2f s", totalDur.Seconds())
	rpsStr := fmt.Sprintf("%.2f", rps)
	avgStr := fmt.Sprintf("%.1f ms", float64(avgLat.Microseconds())/1000.0)
	minMaxStr := fmt.Sprintf("Min %.1fms │ Max %.1fms",
		float64(minLat.Microseconds())/1000.0,
		float64(maxLat.Microseconds())/1000.0)

	// 3. LES 3 CARTES CÔTÉ À CÔTE (Temps Total | Débit RPS | Latence Moyenne)
	// Ligne du haut des 3 boîtes
	fmt.Printf(" %s╭──────────────────────────╮%s  %s╭──────────────────────────╮%s  %s╭──────────────────────────╮%s\n",
		GreenB, Reset, GreenB, Reset, Gray, Reset)

	// Ligne 1 : Titres
	fmt.Printf(" %s│%s %s%-24s%s %s│%s  %s│%s %s%-24s%s %s│%s  %s│%s %s%-24s%s %s│%s\n",
		GreenB, Reset, Gray, "Temps Total", Reset, GreenB, Reset,
		GreenB, Reset, Gray, "Débit RPS", Reset, GreenB, Reset,
		Gray, Reset, Gray, "Latence Moyenne", Reset, Gray, Reset)

	// Ligne 2 : Valeurs géantes en couleur
	fmt.Printf(" %s│%s %s%-24s%s %s│%s  %s│%s %s%-24s%s %s│%s  %s│%s %s%-24s%s %s│%s\n",
		GreenB, Reset, Bold+Green, durStr, Reset, GreenB, Reset,
		GreenB, Reset, Bold+Blue, rpsStr, Reset, GreenB, Reset,
		Gray, Reset, Bold+Orange, avgStr, Reset, Gray, Reset)

	// Ligne 3 : Sous-titres
	fmt.Printf(" %s│%s %s%-24s%s %s│%s  %s│%s %s%-24s%s %s│%s  %s│%s %s%-24s%s %s│%s\n",
		GreenB, Reset, Gray, "Durée d'exécution", Reset, GreenB, Reset,
		GreenB, Reset, Gray, "req / seconde", Reset, GreenB, Reset,
		Gray, Reset, Gray, minMaxStr, Reset, Gray, Reset)

	// Ligne du bas des 3 boîtes
	fmt.Printf(" %s╰──────────────────────────╯%s  %s╰──────────────────────────╯%s  %s╰──────────────────────────╯%s\n\n",
		GreenB, Reset, GreenB, Reset, Gray, Reset)

	// 4. RÉPARTITION DES RÉPONSES HTTP
	fmt.Printf(" %sRÉPARTITION DES RÉPONSES HTTP%s\n\n", Gray, Reset)

	// Affichage 200 OK
	c200 := statusCounts[200]
	pct200 := (float64(c200) / float64(cfg.TotalRequests)) * 100
	fmt.Printf(" %s%-11s%s %s  %s%5d (%.1f%%)%s\n\n",
		Green, "[200 OK]", Reset, makeBar(c200, cfg.TotalRequests, Green), Bold+White, c200, pct200, Reset)

	// Affichage des autres codes (ex: 500, 404...)
	c500 := statusCounts[500]
	pct500 := (float64(c500) / float64(cfg.TotalRequests)) * 100
	fmt.Printf(" %s%-11s%s %s  %s%5d (%.1f%%)%s\n\n",
		Orange, "[500 Err]", Reset, makeBar(c500, cfg.TotalRequests, Orange), Bold+White, c500, pct500, Reset)

	// Affichage des Timeouts / Erreurs réseau
	pctErr := (float64(errCount) / float64(cfg.TotalRequests)) * 100
	fmt.Printf(" %s%-11s%s %s  %s%5d (%.1f%%)%s\n\n",
		Red, "[Timeout]", Reset, makeBar(errCount, cfg.TotalRequests, Red), Bold+White, errCount, pctErr, Reset)
}

func main() {

	var cfg Config
	var wg sync.WaitGroup

	rootCmd := &cobra.Command{
		Use:   "gosurge",
		Short: "Tool for testing HTTP requests",
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
			wg.Wait()
			close(results)

			totalTestDuration := time.Since(startTest)

			rps := float64(cfg.TotalRequests) / totalTestDuration.Seconds()

			statusCounts := make(map[int]int)

			var errorCount int

			var minLatency time.Duration
			var maxLatency time.Duration
			var sumLatency time.Duration

			for res := range results {
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
