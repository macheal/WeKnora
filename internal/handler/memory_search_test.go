package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Tencent/WeKnora/internal/application/service/memory"
	"github.com/Tencent/WeKnora/internal/middleware"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// memoryReadStub serves the read/write surface the new endpoints use. It embeds
// the interface so unimplemented methods fail loudly instead of returning zero
// values.
type memoryReadStub struct {
	interfaces.MemoryService

	available bool
	items     []*types.MemoryItem
	prompt    string
	remember  func(item types.MemoryItem) (*types.MemoryItem, error)
}

func (s memoryReadStub) MemoryAvailable(context.Context) bool { return s.available }

func (s memoryReadStub) SearchMemory(_ context.Context, query string, limit int) interfaces.MemorySearchResult {
	if !s.available {
		return interfaces.MemorySearchResult{}
	}
	out := s.items
	if limit > 0 && limit < len(out) {
		out = out[:limit]
	}
	return interfaces.MemorySearchResult{Available: true, Items: out}
}

func (s memoryReadStub) Recall(context.Context, string) interfaces.MemoryRecall {
	if !s.available {
		return interfaces.MemoryRecall{}
	}
	return interfaces.MemoryRecall{Prompt: s.prompt, Items: s.items}
}

func (s memoryReadStub) Remember(_ context.Context, item types.MemoryItem) (*types.MemoryItem, error) {
	if s.remember != nil {
		return s.remember(item)
	}
	return &types.MemoryItem{ID: "id", Kind: item.Kind, Content: item.Content}, nil
}

func newMemoryRouter(svc interfaces.MemoryService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.ErrorHandler())
	handler := NewMemoryHandler(svc)
	router.POST("/memory/search", handler.SearchMemory)
	router.GET("/memory/search", handler.SearchMemoryQuery)
	router.POST("/memory/recall", handler.Recall)
	router.POST("/memory/items/batch", handler.CreateItems)
	return router
}

func TestMemorySearchReturnsMatches(t *testing.T) {
	router := newMemoryRouter(memoryReadStub{
		available: true,
		items: []*types.MemoryItem{
			{Kind: types.MemoryKindFact, Content: "生产库是 PostgreSQL"},
		},
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/memory/search",
		strings.NewReader(`{"query":"数据库","limit":5}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Available bool               `json:"available"`
			Count     int                `json:"count"`
			Memories  []types.MemoryItem `json:"memories"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.True(t, body.Success)
	require.True(t, body.Data.Available)
	require.Equal(t, 1, body.Data.Count)
	require.Equal(t, "生产库是 PostgreSQL", body.Data.Memories[0].Content)
}

// A disabled scope must not read as an empty store: the caller has to be able
// to tell "nothing matched" from "memory is off".
func TestMemorySearchReportsDisabled(t *testing.T) {
	router := newMemoryRouter(memoryReadStub{available: false})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/memory/search",
		strings.NewReader(`{"query":"数据库"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	require.Contains(t, recorder.Body.String(), `"available":false`)
}

func TestMemorySearchQueryRejectsBadLimit(t *testing.T) {
	router := newMemoryRouter(memoryReadStub{available: true})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/memory/search?query=x&limit=abc", nil)
	router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusBadRequest, recorder.Code, recorder.Body.String())
}

func TestMemoryRecallRequiresQuery(t *testing.T) {
	router := newMemoryRouter(memoryReadStub{available: true})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/memory/recall",
		strings.NewReader(`{"query":"  "}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusBadRequest, recorder.Code, recorder.Body.String())
}

func TestMemoryRecallReturnsPrompt(t *testing.T) {
	router := newMemoryRouter(memoryReadStub{available: true, prompt: "<user_memory>…</user_memory>"})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/memory/recall",
		strings.NewReader(`{"query":"偏好"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	require.Contains(t, recorder.Body.String(), "user_memory")
}

func TestMemoryBatchWritesAndSkips(t *testing.T) {
	router := newMemoryRouter(memoryReadStub{
		available: true,
		remember: func(item types.MemoryItem) (*types.MemoryItem, error) {
			if item.Content == "bad" {
				return nil, memory.ErrSensitiveContent
			}
			return &types.MemoryItem{ID: "id", Kind: item.Kind, Content: item.Content}, nil
		},
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/memory/items/batch",
		strings.NewReader(`{"items":[{"kind":"fact","content":"ok"},{"kind":"fact","content":"bad"}]}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	var body struct {
		Data struct {
			Created int `json:"created"`
			Skipped int `json:"skipped"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, 1, body.Data.Created)
	require.Equal(t, 1, body.Data.Skipped)
}

func TestMemoryBatchRejectsEmpty(t *testing.T) {
	router := newMemoryRouter(memoryReadStub{available: true})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/memory/items/batch",
		strings.NewReader(`{"items":[]}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusBadRequest, recorder.Code, recorder.Body.String())
}

func TestMemoryBatchRejectsWhenDisabled(t *testing.T) {
	router := newMemoryRouter(memoryReadStub{available: false})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/memory/items/batch",
		strings.NewReader(`{"items":[{"kind":"fact","content":"x"}]}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusBadRequest, recorder.Code, recorder.Body.String())
	require.Contains(t, recorder.Body.String(), "memory is disabled")
}

func TestMemoryBatchRejectsOversize(t *testing.T) {
	router := newMemoryRouter(memoryReadStub{available: true})
	items := make([]string, 0, memoryBatchMaxItems+1)
	for i := 0; i < memoryBatchMaxItems+1; i++ {
		items = append(items, `{"kind":"fact","content":"x"}`)
	}
	payload := `{"items":[` + strings.Join(items, ",") + `]}`
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/memory/items/batch", strings.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusBadRequest, recorder.Code, recorder.Body.String())
}
