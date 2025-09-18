# 安全修复建议

## 紧急修复项

### 1. 修复TLS证书验证

**问题：** 所有HTTP客户端都禁用了TLS证书验证

**修复方案：**

```go
// 将所有的 InsecureSkipVerify: true 改为 false
// 或者添加可配置的选项

// 在 .env.example 中添加
SKIP_TLS_VERIFY=false

// 在代码中使用环境变量
skipTLSVerify, _ := strconv.ParseBool(os.Getenv("SKIP_TLS_VERIFY"))
transport := &http.Transport{
    TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLSVerify},
}
```

### 2. 清理敏感日志

**问题：** OAuth响应和授权码在日志中暴露

**修复方案：**

```go
// common/oauth.go 中移除或脱敏日志
// 删除这行：
// SysLog(fmt.Sprintf("OAuth token exchange response: %s", string(body)))

// 改为：
SysLog("OAuth token exchange completed successfully")

// 对于授权码日志，只记录长度：
SysLog(fmt.Sprintf("Attempting OAuth token exchange - Code length: %d", len(cleanedCode)))
```

### 3. 添加数据传输告知

**建议在前端添加明确的数据使用声明：**

```typescript
// web/src/components/DataPrivacyNotice.vue
<template>
  <div class="privacy-notice">
    <h3>数据传输声明</h3>
    <p>使用本服务时，您的对话内容将被传输至以下第三方服务商：</p>
    <ul>
      <li>Anthropic (Claude API) - api.anthropic.com</li>
      <li>用于处理AI对话请求</li>
    </ul>
    <p>请确保您已了解并同意此数据传输行为。</p>
  </div>
</template>
```

## 环境变量配置改进

### 新增环境变量

```bash
# 在 .env.example 中添加
# TLS安全配置
SKIP_TLS_VERIFY=false

# 外部服务地址(允许自定义)
CLAUDE_API_URL=https://api.anthropic.com/v1/messages
CLAUDE_OAUTH_URL=https://console.anthropic.com/v1/oauth/token
CLAUDE_CLIENT_ID=9d1c250a-e61b-44d9-88ed-5944d1962f5e

# 日志级别控制
LOG_SENSITIVE_DATA=false
```

## 代理安全改进

```go
// 添加代理白名单验证
var allowedProxyDomains = []string{
    "proxy.example.com",
    "corporate-proxy.company.com",
}

func isProxyAllowed(proxyURI string) bool {
    if proxyURI == "" {
        return true // 无代理是允许的
    }
    
    proxyURL, err := url.Parse(proxyURI)
    if err != nil {
        return false
    }
    
    for _, domain := range allowedProxyDomains {
        if strings.Contains(proxyURL.Host, domain) {
            return true
        }
    }
    return false
}
```

## 用户身份匿名化

```go
// common/utils.go 中添加
func GenerateAnonymousUserID(userID uint) string {
    // 使用用户ID和系统密钥生成匿名ID
    salt := os.Getenv("ANONYMOUS_ID_SALT")
    if salt == "" {
        salt = "claude-relay-anonymous"
    }
    
    data := fmt.Sprintf("%d-%s", userID, salt)
    hash := sha256.Sum256([]byte(data))
    return fmt.Sprintf("anon_%x", hash[:8]) // 使用前8字节
}

// 在 relay/claude.go 中使用
anonymousID := common.GenerateAnonymousUserID(uint(groupID.(int)))
body, _ = sjson.SetBytes(body, "metadata.user_id", anonymousID)
```

## 配置文件安全检查

```go
// 添加配置安全检查函数
func validateSecurityConfig() []string {
    var warnings []string
    
    if os.Getenv("SKIP_TLS_VERIFY") == "true" {
        warnings = append(warnings, "TLS证书验证已禁用，存在安全风险")
    }
    
    if os.Getenv("LOG_SENSITIVE_DATA") == "true" {
        warnings = append(warnings, "敏感数据日志记录已启用，请谨慎使用")
    }
    
    return warnings
}
```

## 实施优先级

1. **立即修复 (24小时内)**
   - 启用TLS证书验证
   - 清理敏感日志

2. **短期修复 (1周内)**
   - 添加数据传输告知
   - 环境变量配置化
   - 用户身份匿名化

3. **长期改进 (1月内)**
   - 代理白名单机制
   - 完整的安全配置检查
   - 审计日志系统

这些修复将显著提高系统的安全性，同时保持功能完整性。