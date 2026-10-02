package main

import (
	"os"
	"testing"
)

// 验证 ZHIZHANG_ 前缀环境变量能覆盖嵌套配置（需要 SetEnvKeyReplacer）
func TestLoadConfigFromEnv(t *testing.T) {
	// config.yaml 在仓库根目录 configs/ 下，loadConfig 的相对路径基于仓库根
	if err := os.Chdir("../.."); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ZHIZHANG_WECOM_CORPID", "ww_from_env")
	t.Setenv("ZHIZHANG_WECOM_AGENTID", "1000002")
	t.Setenv("ZHIZHANG_WECOM_SECRET", "secret_from_env")
	t.Setenv("ZHIZHANG_WECOM_INVITE_QR_URL", "https://example.com/qr.png")
	t.Setenv("ZHIZHANG_WECHAT_MP_APPID", "wx_from_env")

	cfg, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Wecom.CorpID != "ww_from_env" {
		t.Fatalf("env override failed, got corpid=%q", cfg.Wecom.CorpID)
	}
	if cfg.Wecom.AgentID != 1000002 {
		t.Fatalf("env override failed, got agentid=%d", cfg.Wecom.AgentID)
	}
	if cfg.Wecom.InviteQRURL != "https://example.com/qr.png" {
		t.Fatalf("invite qr url mismatch: %q", cfg.Wecom.InviteQRURL)
	}
	if cfg.WechatMP.AppID != "wx_from_env" {
		t.Fatalf("wxmp appid mismatch: %q", cfg.WechatMP.AppID)
	}
}
