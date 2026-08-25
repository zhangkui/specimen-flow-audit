# 地震波形质量分析工具

`specimen-flow-audit` 是一个 Go 编写的地震波形质量分析服务。项目名沿用既有仓库名称，业务功能已调整为波形质量分析。

服务接受 JSON 波形片段，校验台站、起始时间、采样率和样本值，计算片段时长、零值缺失段、RMS 噪声和完整率，并保存每个台站最新一次分析的数据。

## 运行

```powershell
go run .
```

服务默认监听 `:8080`。可通过 `SPECIMEN_FLOW_AUDIT_ADDR` 指定地址。

## 分析请求

```powershell
Invoke-RestMethod -Method Post http://localhost:8080/traces/analyze -ContentType application/json -Body '{"station":"HN01","start":"2026-08-24T00:00:00Z","sample_rate":4,"samples":[1,2,0,0,3,4]}'
```

返回的质量报告包含 `duration_seconds`、`gap_count`、`noise_rms` 与 `completeness`。

## 验证

```powershell
go test ./...
go build ./...
```
