package common

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

// CreateSecureHTTPClient 创建安全的HTTP客户端
// 统一管理TLS配置和代理设置，提高安全性
func CreateSecureHTTPClient(timeout time.Duration, proxyURI string) *http.Client {
	// 默认启用TLS证书验证，除非明确配置为跳过
	skipTLSVerify := os.Getenv("SKIP_TLS_VERIFY") == "true"
	
	// 如果跳过TLS验证，记录警告
	if skipTLSVerify {
		SysLog("警告: TLS证书验证已禁用，存在安全风险")
	}
	
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLSVerify},
	}

	// 配置代理（如果提供）
	if proxyURI != "" {
		proxyURL, err := url.Parse(proxyURI)
		if err != nil {
			SysLog("代理URL解析失败: " + err.Error())
		} else {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

// ValidateProxyURI 验证代理URI的安全性
func ValidateProxyURI(proxyURI string) error {
	if proxyURI == "" {
		return nil
	}
	
	proxyURL, err := url.Parse(proxyURI)
	if err != nil {
		return err
	}
	
	// 检查是否使用安全协议
	if proxyURL.Scheme != "http" && proxyURL.Scheme != "https" && proxyURL.Scheme != "socks5" {
		SysLog("代理协议可能不安全: " + proxyURL.Scheme)
	}
	
	return nil
}

// GetSecurityConfig 获取安全配置并返回警告
func GetSecurityConfig() []string {
	var warnings []string
	
	if os.Getenv("SKIP_TLS_VERIFY") == "true" {
		warnings = append(warnings, "TLS证书验证已禁用")
	}
	
	if os.Getenv("LOG_SENSITIVE_DATA") == "true" {
		warnings = append(warnings, "敏感数据日志记录已启用")
	}
	
	return warnings
}

// LogSecurityWarnings 在启动时记录安全警告
func LogSecurityWarnings() {
	warnings := GetSecurityConfig()
	if len(warnings) > 0 {
		SysLog("=== 安全警告 ===")
		for _, warning := range warnings {
			SysLog("⚠️  " + warning)
		}
		SysLog("===============")
	}
}