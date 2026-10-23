package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type quotaWindow struct {
	UsedPercent        float64 `json:"usedPercent"`
	WindowDurationMins *int    `json:"windowDurationMins"`
	ResetsAt           *int64  `json:"resetsAt"`
}

type quotaBucket struct {
	PlanType  string       `json:"planType"`
	Primary   *quotaWindow `json:"primary"`
	Secondary *quotaWindow `json:"secondary"`
}

type resetCredits struct {
	AvailableCount int `json:"availableCount"`
}

type quotaResult struct {
	RateLimits            quotaBucket            `json:"rateLimits"`
	RateLimitsByLimitID   map[string]quotaBucket `json:"rateLimitsByLimitId"`
	RateLimitResetCredits *resetCredits          `json:"rateLimitResetCredits"`
}

type rpcError struct {
	Message string `json:"message"`
}

type rpcResponse struct {
	ID     *int            `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *rpcError       `json:"error"`
}

type Usage struct {
	Plan              string
	FiveHourRemaining *float64
	WeeklyRemaining   *float64
	ResetCredits      *int
}

func ReadUsage(name string) (Usage, error) {
	var usage Usage

	profileDir, err := ProfileDir(name)
	if err != nil {
		return usage, err
	}
	result, err := readUsage(profileDir)
	if err != nil {
		return usage, err
	}

	bucket := result.RateLimits
	if codex, ok := result.RateLimitsByLimitID["codex"]; ok {
		bucket = codex
	}
	usage.Plan = bucket.PlanType

	for _, window := range []*quotaWindow{bucket.Primary, bucket.Secondary} {
		if window == nil || window.WindowDurationMins == nil {
			continue
		}
		remaining := 100 - window.UsedPercent
		switch *window.WindowDurationMins {
		case 300:
			usage.FiveHourRemaining = &remaining
		case 10080:
			usage.WeeklyRemaining = &remaining
		}
	}
	if result.RateLimitResetCredits != nil {
		usage.ResetCredits = &result.RateLimitResetCredits.AvailableCount
	}
	return usage, nil
}

func readUsage(profileDir string) (quotaResult, error) {
	var result quotaResult

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "codex", "app-server")
	cmd.Env = append(os.Environ(), "CODEX_HOME="+profileDir)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return result, err
	}
	defer stdin.Close()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return result, err
	}

	if err := cmd.Start(); err != nil {
		return result, err
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	encoder := json.NewEncoder(stdin)
	decoder := json.NewDecoder(stdout)

	call := func(id int, method string, params any) (json.RawMessage, error) {
		if err := encoder.Encode(map[string]any{
			"id":     id,
			"method": method,
			"params": params,
		}); err != nil {
			return nil, err
		}

		for {
			var response rpcResponse
			if err := decoder.Decode(&response); err != nil {
				if ctx.Err() != nil {
					return nil, ctx.Err()
				}
				return nil, err
			}
			if response.ID == nil || *response.ID != id {
				continue
			}
			if response.Error != nil {
				return nil, fmt.Errorf("%s: %s", method, response.Error.Message)
			}
			return response.Result, nil
		}
	}

	if _, err := call(1, "initialize", map[string]any{
		"clientInfo": map[string]string{
			"name":    "codex-profiles",
			"version": "0.1.0",
		},
	}); err != nil {
		return result, err
	}

	if err := encoder.Encode(map[string]any{
		"method": "initialized",
		"params": map[string]any{},
	}); err != nil {
		return result, err
	}

	data, err := call(2, "account/rateLimits/read", nil)
	if err != nil {
		return result, err
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return result, err
	}

	return result, nil
}
