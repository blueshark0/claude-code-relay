package common

import (
	"strings"
)

// TestRequestBodyTemplate 测试用的标准请求体模板
const TestRequestBodyTemplate = `{
	"model": "claude-3-5-haiku-20241022",
	"max_tokens": 512,
    "messages": [
        {
            "role": "user",
            "content": "Please write a 5-10 word title for the following conversation:\n\nUser: hello\n\nRespond with the title for the conversation and nothing else."
        }
    ],
    "system": [
        {
            "type": "text",
            "text": "Summarize this coding conversation in under 50 characters.\nCapture the main task, key files, problems addressed, and current status.",
            "cache_control": {
                "type": "ephemeral"
            }
        }
    ],
	"metadata": {
		"user_id": "user_cc036db0165a4c6a12bfb24190e994e65bc28f0625db9ba581806b633f14d672_account__session_598af368-4d3c-4ebe-9149-925d142c4119"
	},
	"stream": true
}`

// GetTestRequestBody 获取带指定max_tokens的测试请求体
func GetTestRequestBody(maxTokens int) string {
	//return fmt.Sprintf(TestRequestBodyTemplate, maxTokens)
	return TestRequestBodyTemplate
}

// TestRequestBody 默认测试请求体（64000 tokens）
var TestRequestBody = GetTestRequestBody(64000)

// getGlobalClaudeCodeHeaders 获取全局Claude Code请求头
func getGlobalClaudeCodeHeaders() map[string]string {
	return map[string]string{
		"anthropic-version":                         "2023-06-01",
		"X-Stainless-Retry-Count":                   "0",
		"X-Stainless-Timeout":                       "600",
		"X-Stainless-Lang":                          "js",
		"X-Stainless-Package-Version":               "0.55.1",
		"X-Stainless-OS":                            "MacOS",
		"X-Stainless-Arch":                          "arm64",
		"X-Stainless-Runtime":                       "node",
		"x-stainless-helper-method":                 "stream",
		"x-app":                                     "cli",
		"User-Agent":                                "claude-cli/1.0.120 (external, cli)",
		"anthropic-beta":                            "claude-code-20250219,oauth-2025-04-20,interleaved-thinking-2025-05-14,fine-grained-tool-streaming-2025-05-14",
		"X-Stainless-Runtime-Version":               "v20.18.1",
		"anthropic-dangerous-direct-browser-access": "true",
	}
}

// MergeHeaders 合并全局Claude Code请求头和用户提供的请求头
// 用户提供的头部优先级更高，可以覆盖全局头部
func MergeHeaders(headers map[string]string, anthropicBeta string) map[string]string {
	globalHeaders := getGlobalClaudeCodeHeaders()

	result := make(map[string]string, len(globalHeaders)+len(headers))

	for k, v := range globalHeaders {
		result[k] = v
	}

	// 用户提供的头部优先级更高，可以覆盖全局头部
	for k, v := range headers {
		result[k] = v
	}

	// 使用原始的 anthropic-beta 请求头（如果存在）
	if anthropicBeta != "" {
		if strings.Contains(anthropicBeta, "oauth-") {
			result["anthropic-beta"] = anthropicBeta
		} else {
			result["anthropic-beta"] = "oauth-2025-04-20," + anthropicBeta
		}
	}

	return result
}
