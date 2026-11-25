package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// 角色常量，与 OpenAI 协议一致
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// ErrEmbeddingDisabled 未配置向量模型时的哨兵错误，调用方据此降级为词面检索
var ErrEmbeddingDisabled = errors.New("未配置向量模型")

// Message 一条对话消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatOptions 生成参数
type ChatOptions struct {
	Temperature float64
	MaxTokens   int
	TopP        float64
}

// ClientConfig 客户端配置
type ClientConfig struct {
	BaseURL        string
	APIKey         string
	Model          string
	EmbeddingModel string
	Timeout        time.Duration
}

// Client 一个 OpenAI 兼容协议的模型客户端
// OpenAI / DeepSeek / Moonshot / SiliconFlow / Ollama(/v1) 都能直接复用
type Client struct {
	baseURL        string
	apiKey         string
	model          string
	embeddingModel string
	chatHTTP       *http.Client // 非流式：整体超时
	streamHTTP     *http.Client // 流式：只限制首包时间，整体由 ctx 控制
}

// sharedTransport 进程级连接池，避免每次请求都新建 Transport 导致连接无法复用
var sharedTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	ForceAttemptHTTP2:     true,
	MaxIdleConns:          100,
	MaxIdleConnsPerHost:   16,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	ResponseHeaderTimeout: 60 * time.Second,
}

func NewClient(cfg ClientConfig) *Client {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 120 * time.Second
	}

	return &Client{
		baseURL:        normalizeBaseURL(cfg.BaseURL),
		apiKey:         strings.TrimSpace(cfg.APIKey),
		model:          strings.TrimSpace(cfg.Model),
		embeddingModel: strings.TrimSpace(cfg.EmbeddingModel),
		chatHTTP: &http.Client{
			Transport: sharedTransport,
			Timeout:   timeout,
		},
		streamHTTP: &http.Client{
			Transport: sharedTransport,
		},
	}
}

// Model 返回对话模型名
func (c *Client) Model() string {
	return c.model
}

// EmbeddingModel 返回向量模型名
func (c *Client) EmbeddingModel() string {
	return c.embeddingModel
}

// Probe 用一次极小的对话请求验证 BaseURL / APIKey / Model 是否可用
func (c *Client) Probe(ctx context.Context) error {
	_, err := c.Chat(ctx, []Message{{Role: RoleUser, Content: "ping"}}, ChatOptions{MaxTokens: 4})
	return err
}

// Chat 非流式对话
func (c *Client) Chat(ctx context.Context, messages []Message, opts ChatOptions) (string, error) {
	if c.model == "" {
		return "", errors.New("对话模型未配置")
	}

	body, err := json.Marshal(chatCompletionRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: opts.Temperature,
		MaxTokens:   opts.MaxTokens,
		TopP:        opts.TopP,
	})
	if err != nil {
		return "", fmt.Errorf("构造对话请求失败：%w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.endpoint("chat/completions"),
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("构造对话请求失败：%w", err)
	}
	c.setHeaders(req)
	req.Header.Set("Accept", "application/json")

	resp, err := c.chatHTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求模型服务失败：%w", err)
	}
	defer closeBody(resp.Body)

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", c.statusError(resp)
	}

	var out chatCompletionResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&out); err != nil {
		return "", fmt.Errorf("解析模型响应失败：%w", err)
	}
	if out.Error != nil && out.Error.Message != "" {
		return "", fmt.Errorf("模型服务错误：%s", out.Error.Message)
	}
	if len(out.Choices) == 0 {
		return "", errors.New("模型未返回任何内容")
	}
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}

// StreamChat 流式对话
// onDelta 每收到一个增量片段回调一次；返回 error 立即中断（用于客户端断开）
// 返回值是拼接后的完整文本（可能因中断而只包含一部分）
func (c *Client) StreamChat(ctx context.Context, messages []Message, opts ChatOptions,
	onDelta func(delta string) error) (string, error) {
	if c.model == "" {
		return "", errors.New("对话模型未配置")
	}

	body, err := json.Marshal(chatCompletionRequest{
		Model:       c.model,
		Messages:    messages,
		Stream:      true,
		Temperature: opts.Temperature,
		MaxTokens:   opts.MaxTokens,
		TopP:        opts.TopP,
	})
	if err != nil {
		return "", fmt.Errorf("构造对话请求失败：%w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint("chat/completions"), bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("构造对话请求失败：%w", err)
	}
	c.setHeaders(req)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := c.streamHTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求模型服务失败：%w", err)
	}
	defer closeBody(resp.Body)

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", c.statusError(resp)
	}

	var (
		builder strings.Builder
		scanner = bufio.NewScanner(resp.Body)
	)
	// 单行可能很长（含 base64 图片等），放宽到 1MB
	scanner.Buffer(make([]byte, 0, 64<<10), 1<<20)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 空行是事件分隔符，以 ":" 开头的是注释/心跳
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		// 只关心 data 行，event:/id:/retry: 一律忽略
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" {
			continue
		}
		if data == "[DONE]" {
			break
		}

		var chunk chatCompletionChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			// 个别服务会插入非标准心跳，跳过即可
			continue
		}
		if chunk.Error != nil && chunk.Error.Message != "" {
			return builder.String(), fmt.Errorf("模型服务错误：%s", chunk.Error.Message)
		}
		for _, choice := range chunk.Choices {
			delta := choice.Delta.Content
			if delta == "" {
				continue
			}
			builder.WriteString(delta)
			if onDelta != nil {
				if err := onDelta(delta); err != nil {
					return builder.String(), err
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		if ctx.Err() != nil {
			return builder.String(), ctx.Err()
		}
		return builder.String(), fmt.Errorf("读取模型流失败：%w", err)
	}
	return builder.String(), nil
}

// Embed 批量向量化，返回顺序与入参一致
func (c *Client) Embed(ctx context.Context, inputs []string) ([][]float32, error) {
	if c.embeddingModel == "" {
		return nil, ErrEmbeddingDisabled
	}
	if len(inputs) == 0 {
		return nil, nil
	}

	body, err := json.Marshal(embeddingRequest{
		Model:          c.embeddingModel,
		Input:          inputs,
		EncodingFormat: "float",
	})
	if err != nil {
		return nil, fmt.Errorf("构造向量请求失败：%w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint("embeddings"), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("构造向量请求失败：%w", err)
	}
	c.setHeaders(req)
	req.Header.Set("Accept", "application/json")

	resp, err := c.chatHTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求向量服务失败：%w", err)
	}
	defer closeBody(resp.Body)

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, c.statusError(resp)
	}

	var out embeddingResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64<<20)).Decode(&out); err != nil {
		return nil, fmt.Errorf("解析向量响应失败：%w", err)
	}
	if out.Error != nil && out.Error.Message != "" {
		return nil, fmt.Errorf("向量服务错误：%s", out.Error.Message)
	}
	if len(out.Data) != len(inputs) {
		return nil, fmt.Errorf("向量数量不匹配：期望 %d，实际 %d", len(inputs), len(out.Data))
	}

	// 服务返回的 index 不一定有序，按 index 回填
	vectors := make([][]float32, len(inputs))
	for _, item := range out.Data {
		if item.Index < 0 || item.Index >= len(vectors) {
			return nil, fmt.Errorf("向量下标越界：%d", item.Index)
		}
		vectors[item.Index] = item.Embedding
	}
	for i, v := range vectors {
		if len(v) == 0 {
			return nil, fmt.Errorf("第 %d 条文本没有返回向量", i)
		}
	}
	return vectors, nil
}

// ==================================================内部实现===========================================================

func (c *Client) endpoint(path string) string {
	return c.baseURL + "/" + strings.TrimLeft(path, "/")
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
}

// statusError 把非 2xx 响应转换为带响应体的错误，便于后台看到具体原因
func (c *Client) statusError(resp *http.Response) error {
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	msg := strings.TrimSpace(string(raw))
	if msg == "" {
		return fmt.Errorf("模型服务返回状态码 %d", resp.StatusCode)
	}
	return fmt.Errorf("模型服务返回状态码 %d：%s", resp.StatusCode, msg)
}

func closeBody(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, io.LimitReader(body, 1<<20))
	_ = body.Close()
}

func normalizeBaseURL(raw string) string {
	base := strings.TrimSpace(raw)
	base = strings.TrimRight(base, "/")
	for _, suffix := range []string{"/chat/completions", "/embeddings", "/models"} {
		if strings.HasSuffix(base, suffix) {
			base = strings.TrimSuffix(base, suffix)
		}
	}
	return strings.TrimRight(base, "/")
}

// ==================================================协议结构体=========================================================

type chatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *apiError `json:"error"`
}

type chatCompletionChunk struct {
	Choices []struct {
		Delta struct {
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"` // DeepSeek 推理模型，正文里不混入思考过程
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *apiError `json:"error"`
}

type apiError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

type embeddingRequest struct {
	Model          string   `json:"model"`
	Input          []string `json:"input"`
	EncodingFormat string   `json:"encoding_format,omitempty"`
}

type embeddingResponse struct {
	Data []struct {
		Index     int       `json:"index"`
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Error *apiError `json:"error"`
}
