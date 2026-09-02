package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	dryRunFlag    bool
	valuesDirFlag string
	namespaceFlag string
)

var services = []string{
	"identity-service",
	"transaction-manager",
	"financing-engine",
	"tokenization-engine",
	"ledger-service",
	"property-registry-service",
	"feedback-service",
}

var offlineBaselines = map[string]struct {
	cpu string
	mem string
}{
	"identity-service":          {cpu: "45m", mem: "64Mi"},
	"transaction-manager":       {cpu: "60m", mem: "80Mi"},
	"financing-engine":          {cpu: "40m", mem: "64Mi"},
	"tokenization-engine":       {cpu: "45m", mem: "64Mi"},
	"ledger-service":            {cpu: "50m", mem: "72Mi"},
	"property-registry-service": {cpu: "35m", mem: "48Mi"},
	"feedback-service":          {cpu: "25m", mem: "32Mi"},
}

const (
	bufferPercentage      = 1.20
	minCPUMillis          = 25
	minMemoryMi           = 32
	limitMemoryMultiplier = 1.5
)

type VPARecommendation struct {
	Status struct {
		Recommendation struct {
			ContainerRecommendations []struct {
				Target struct {
					CPU    string `json:"cpu"`
					Memory string `json:"memory"`
				} `json:"target"`
			} `json:"containerRecommendations"`
		} `json:"recommendation"`
	} `json:"status"`
}

type RightSizeResult struct {
	Service  string
	OldCPU   string
	NewCPU   string
	OldMem   string
	NewMem   string
	LimitMem string
}

var finopsCmd = &cobra.Command{
	Use:   "finops",
	Short: "Kubernetes FinOps and Workload Right-Sizing Utilities",
	Long:  `Analyze, monitor, and right-size microservice resource requests and limits using VPA and OpenCost metrics.`,
}

var rightsizeCmd = &cobra.Command{
	Use:   "rightsize",
	Short: "Right-size microservice CPU and memory requests from VPA recommendations",
	Long:  `Fetch historical p95/p99 consumption metrics from VPA, apply +20% safety headroom, and update Helm values files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRightSize(valuesDirFlag, namespaceFlag, dryRunFlag)
	},
}

func init() {
	rightsizeCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Preview recommendations without writing to disk")
	rightsizeCmd.Flags().StringVar(&valuesDirFlag, "values-dir", "infra/kind/values", "Path to microservice Helm values directory")
	rightsizeCmd.Flags().StringVar(&namespaceFlag, "namespace", "realestate-trust", "Kubernetes namespace to query VPA from")

	finopsCmd.AddCommand(rightsizeCmd)
	rootCmd.AddCommand(finopsCmd)
}

func parseCPUToMillis(cpuStr string) int {
	cpuStr = strings.TrimSpace(cpuStr)
	if cpuStr == "" {
		return minCPUMillis
	}
	if strings.HasSuffix(cpuStr, "m") {
		val, err := strconv.Atoi(strings.TrimSuffix(cpuStr, "m"))
		if err == nil {
			return val
		}
	}
	val, err := strconv.ParseFloat(cpuStr, 64)
	if err == nil {
		return int(val * 1000)
	}
	return minCPUMillis
}

func parseMemoryToMi(memStr string) int {
	memStr = strings.TrimSpace(memStr)
	if memStr == "" {
		return minMemoryMi
	}
	if strings.HasSuffix(memStr, "Gi") {
		val, err := strconv.ParseFloat(strings.TrimSuffix(memStr, "Gi"), 64)
		if err == nil {
			return int(val * 1024)
		}
	}
	if strings.HasSuffix(memStr, "Mi") {
		val, err := strconv.Atoi(strings.TrimSuffix(memStr, "Mi"))
		if err == nil {
			return val
		}
	}
	if strings.HasSuffix(memStr, "Ki") || strings.HasSuffix(memStr, "k") {
		val, err := strconv.Atoi(strings.TrimSuffix(strings.TrimSuffix(memStr, "Ki"), "k"))
		if err == nil {
			res := val / 1024
			if res < 1 {
				return 1
			}
			return res
		}
	}
	bytesVal, err := strconv.ParseInt(memStr, 10, 64)
	if err == nil {
		return int(bytesVal / (1024 * 1024))
	}
	return minMemoryMi
}

func getVPARecommendation(service, namespace string) (string, string) {
	cmd := exec.Command("kubectl", "get", "vpa", fmt.Sprintf("%s-vpa", service), "-n", namespace, "-o", "json")
	out, err := cmd.Output()
	if err == nil {
		var vpa VPARecommendation
		if jsonErr := json.Unmarshal(out, &vpa); jsonErr == nil {
			if len(vpa.Status.Recommendation.ContainerRecommendations) > 0 {
				rec := vpa.Status.Recommendation.ContainerRecommendations[0].Target
				if rec.CPU != "" && rec.Memory != "" {
					return rec.CPU, rec.Memory
				}
			}
		}
	}

	fallback, exists := offlineBaselines[service]
	if exists {
		return fallback.cpu, fallback.mem
	}
	return "50m", "64Mi"
}

func runRightSize(valuesDir, namespace string, dryRun bool) error {
	var results []RightSizeResult

	for _, svc := range services {
		targetFile := filepath.Join(valuesDir, fmt.Sprintf("%s.yaml", svc))
		content, err := os.ReadFile(targetFile)
		if err != nil {
			continue
		}

		var node yaml.Node
		if err := yaml.Unmarshal(content, &node); err != nil {
			continue
		}

		recCPU, recMem := getVPARecommendation(svc, namespace)
		rawCPUMillis := parseCPUToMillis(recCPU)
		rawMemMi := parseMemoryToMi(recMem)

		targetCPUMillis := int(float64(rawCPUMillis) * bufferPercentage)
		if targetCPUMillis < minCPUMillis {
			targetCPUMillis = minCPUMillis
		}

		targetMemMi := int(float64(rawMemMi) * bufferPercentage)
		if targetMemMi < minMemoryMi {
			targetMemMi = minMemoryMi
		}

		targetLimitMemMi := int(float64(targetMemMi) * limitMemoryMultiplier)

		// Parse existing values to read previous settings
		var docMap map[string]interface{}
		_ = yaml.Unmarshal(content, &docMap)

		oldCPU := "default"
		oldMem := "default"
		if res, ok := docMap["resources"].(map[string]interface{}); ok {
			if reqs, ok := res["requests"].(map[string]interface{}); ok {
				if c, ok := reqs["cpu"].(string); ok && c != "" {
					oldCPU = c
				}
				if m, ok := reqs["memory"].(string); ok && m != "" {
					oldMem = m
				}
			}
		} else {
			docMap["resources"] = make(map[string]interface{})
		}

		// Update docMap
		docMap["resources"] = map[string]interface{}{
			"requests": map[string]interface{}{
				"cpu":    fmt.Sprintf("%dm", targetCPUMillis),
				"memory": fmt.Sprintf("%dMi", targetMemMi),
			},
			"limits": map[string]interface{}{
				"memory": fmt.Sprintf("%dMi", targetLimitMemMi),
			},
		}

		if !dryRun {
			updatedYAML, err := yaml.Marshal(docMap)
			if err == nil {
				_ = os.WriteFile(targetFile, updatedYAML, 0644)
			}
		}

		results = append(results, RightSizeResult{
			Service:  svc,
			OldCPU:   oldCPU,
			NewCPU:   fmt.Sprintf("%dm", targetCPUMillis),
			OldMem:   oldMem,
			NewMem:   fmt.Sprintf("%dMi", targetMemMi),
			LimitMem: fmt.Sprintf("%dMi", targetLimitMemMi),
		})
	}

	mode := "APPLIED"
	if dryRun {
		mode = "DRY RUN"
	}

	fmt.Println()
	fmt.Println("========================================================")
	fmt.Printf("💰 FinOps Right-Sizing Analysis (%s)\n", mode)
	fmt.Println("========================================================")
	fmt.Printf("%-28s | %-18s | %-18s | %-12s\n", "SERVICE", "CPU REQUEST", "MEMORY REQUEST", "MEMORY LIMIT")
	fmt.Println(strings.Repeat("-", 84))
	for _, r := range results {
		fmt.Printf("%-28s | %s -> %-7s | %s -> %-7s | %s\n", r.Service, r.OldCPU, r.NewCPU, r.OldMem, r.NewMem, r.LimitMem)
	}
	fmt.Println("========================================================")
	fmt.Println()

	return nil
}
