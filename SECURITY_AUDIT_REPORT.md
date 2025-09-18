# Claude Code Relay 安全审查报告

## 概述

本报告对 `claude-code-relay` 项目进行全面的安全审查，重点关注数据传输、固定地址通信、敏感信息处理等安全风险。

## 执行摘要

**严重安全风险发现：**
- ✅ **确认存在向固定地址发送数据**
- ⚠️ **TLS证书验证被禁用**
- ⚠️ **敏感信息可能通过日志泄露**
- ⚠️ **硬编码的客户端ID和固定URL**
- ✅ **代理配置存在安全风险**

## 详细安全发现

### 1. 固定地址数据传输 (严重风险) ✅

**发现：** 项目确实向多个固定的外部地址发送用户数据

**固定地址列表：**
```go
// relay/claude.go
ClaudeAPIURL         = "https://api.anthropic.com/v1/messages?beta=true"
ClaudeCountTokensURL = "https://api.anthropic.com/v1/messages/count_tokens?beta=true"
ClaudeOAuthTokenURL  = "https://console.anthropic.com/v1/oauth/token"

// common/oauth.go
AuthorizeURL: "https://claude.ai/oauth/authorize"
TokenURL:     "https://console.anthropic.com/v1/oauth/token"
RedirectURI:  "https://console.anthropic.com/oauth/code/callback"
```

**数据传输内容：**
- 用户的完整聊天消息和系统提示
- API密钥和访问令牌
- 用户身份标识符
- 代理配置信息

**风险等级：** 🔴 严重
**影响：** 所有用户数据都会被转发到Anthropic的服务器

### 2. TLS证书验证禁用 (高风险) ⚠️

**发现：** 多个HTTP客户端配置中禁用了TLS证书验证

**受影响的文件：**
```go
// relay/claude.go:574, 448, 157
TLSClientConfig: &tls.Config{InsecureSkipVerify: true}

// relay/openai.go:244, 1110
TLSClientConfig: &tls.Config{InsecureSkipVerify: true}

// relay/claude_console.go:134
TLSClientConfig: &tls.Config{InsecureSkipVerify: true}
```

**风险等级：** 🟡 高
**影响：** 易受中间人攻击，数据传输安全性降低

### 3. 敏感信息日志泄露 ⚠️

**发现：** 日志中可能包含敏感信息

**问题代码：**
```go
// common/oauth.go:332
SysLog(fmt.Sprintf("OAuth token exchange response: %s", string(body)))

// common/oauth.go:289
SysLog(fmt.Sprintf("Attempting OAuth token exchange - URL: %s, Code length: %d, Code prefix: %s...", 
    o.config.TokenURL, len(cleanedCode), cleanedCode[:min(10, len(cleanedCode))]))
```

**风险等级：** 🟡 中等
**影响：** 访问令牌和授权码可能通过日志泄露

### 4. 硬编码凭据 ⚠️

**发现：** 项目中存在硬编码的客户端ID

**问题代码：**
```go
// relay/claude.go:33
ClaudeOAuthClientID = "9d1c250a-e61b-44d9-88ed-5944d1962f5e"

// common/oauth.go:31
ClientID: "9d1c250a-e61b-44d9-88ed-5944d1962f5e"
```

**风险等级：** 🟡 中等
**影响：** 客户端ID泄露，无法轮换凭据

### 5. 代理配置安全风险 ⚠️

**发现：** 代理配置缺乏验证和限制

**问题代码：**
```go
// 各个relay文件中
if account.EnableProxy && account.ProxyURI != "" {
    proxyURL, err := url.Parse(account.ProxyURI)
    if err == nil {
        transport.Proxy = http.ProxyURL(proxyURL)
    }
}
```

**风险等级：** 🟡 中等
**影响：** 恶意代理可能截获所有通信

### 6. 身份标识符传输

**发现：** 向外部服务发送用户身份标识符

**问题代码：**
```go
// relay/claude.go
body, _ = sjson.SetBytes(body, "metadata.user_id", model.GetInstanceID(uint(groupID.(int))))
body, _ = sjson.SetBytes(body, "metadata.user_id", common.GetInstanceID())
```

**风险等级：** 🟡 中等
**影响：** 用户可被外部服务跟踪和识别

## 正面安全措施

### 1. 身份验证机制 ✅
- JWT token验证
- API密钥认证
- 用户状态检查

### 2. 输入验证 ✅
- Gin框架的数据绑定验证
- 邮箱格式验证
- 密码长度要求

### 3. 数据库安全 ✅
- 使用GORM ORM防止SQL注入
- 软删除机制
- 密码哈希存储

### 4. 限流机制 ✅
- API请求限流
- 邮件发送频率限制

## 安全建议

### 优先级 1 (立即修复)

1. **禁用TLS证书跳过**
   ```go
   // 移除所有 InsecureSkipVerify: true
   TLSClientConfig: &tls.Config{InsecureSkipVerify: false}
   ```

2. **清理敏感日志**
   ```go
   // 移除或脱敏敏感信息日志
   // 不要记录token响应内容
   ```

3. **添加数据传输警告**
   - 在用户界面添加明确的数据传输声明
   - 说明用户数据将发送至Anthropic服务器

### 优先级 2 (短期内修复)

1. **代理白名单机制**
   ```go
   // 添加代理URL白名单验证
   allowedProxies := []string{"proxy1.com", "proxy2.com"}
   ```

2. **环境变量化配置**
   ```go
   // 将硬编码的URL和客户端ID移至环境变量
   ClientID: os.Getenv("CLAUDE_CLIENT_ID")
   ```

3. **用户身份匿名化**
   ```go
   // 使用哈希或随机ID代替真实用户ID
   anonymousID := generateAnonymousID(userID)
   ```

### 优先级 3 (长期改进)

1. **实现端到端加密**
2. **添加数据加密存储**
3. **实现审计日志**
4. **添加安全头**

## 合规建议

1. **数据隐私声明**
   - 明确告知用户数据传输到第三方
   - 说明数据用途和存储期限

2. **用户控制选项**
   - 提供数据删除功能
   - 允许用户选择是否发送到外部服务

3. **安全通信协议**
   - 制定事件响应计划
   - 建立安全漏洞报告机制

## 结论

该项目确实向固定的外部地址（主要是Anthropic的服务器）发送用户数据。虽然这是项目的核心功能，但存在多个安全风险需要立即处理，特别是TLS证书验证禁用和敏感信息日志记录问题。

建议优先修复高风险问题，并在用户界面明确告知数据传输情况，确保用户知情同意。

---
**审查日期：** 2024年12月18日  
**审查范围：** 完整代码库  
**审查工具：** 手动代码审查 + 自动化搜索  