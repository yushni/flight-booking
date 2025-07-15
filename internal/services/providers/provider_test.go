package providers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"flight-booking/internal/config"
	"flight-booking/internal/models"
	"flight-booking/internal/services/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockCache struct {
	mock.Mock
}

func (m *MockCache) GetOrLoad(key string, ttl time.Duration, loader func() (any, error)) (any, error) {
	args := m.Called(key, ttl, loader)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0), args.Error(1)
}

func createTestConfig(provider1URL, provider2URL string) config.Config {
	return config.Config{
		Providers: config.ProvidersConfig{
			Provider1BaseURL:  provider1URL,
			Provider1Timeout:  30 * time.Second,
			Provider1CacheTTL: 60 * time.Second,
			Provider2BaseURL:  provider2URL,
			Provider2Timeout:  30 * time.Second,
			Provider2CacheTTL: 60 * time.Second,
		},
	}
}

func createMockRoutes(provider string) []models.Route {
	return []models.Route{
		{
			Airline:            "AA",
			SourceAirport:      "JFK",
			DestinationAirport: "LAX",
			CodeShare:          "Y",
			Stops:              0,
			Provider:           provider,
		},
		{
			Airline:            "UA",
			SourceAirport:      "JFK",
			DestinationAirport: "SFO",
			CodeShare:          "N",
			Stops:              1,
			Provider:           provider,
		},
	}
}

func TestProvider_GetRoutes_Success(t *testing.T) {
	t.Parallel()

	provider1Routes := createMockRoutes("provider1")
	provider2Routes := createMockRoutes("provider2")

	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{
				"airline": "AA",
				"sourceAirport": "JFK",
				"destinationAirport": "LAX",
				"codeShare": "Y",
				"stops": 0,
				"provider": "provider1"
			},
			{
				"airline": "UA",
				"sourceAirport": "JFK",
				"destinationAirport": "SFO",
				"codeShare": "N",
				"stops": 1,
				"provider": "provider1"
			}
		]`))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{
				"airline": "AA",
				"sourceAirport": "JFK",
				"destinationAirport": "LAX",
				"codeShare": "Y",
				"stops": 0,
				"provider": "provider2"
			},
			{
				"airline": "UA",
				"sourceAirport": "JFK",
				"destinationAirport": "SFO",
				"codeShare": "N",
				"stops": 1,
				"provider": "provider2"
			}
		]`))
	}))
	defer server2.Close()

	mockCache := &MockCache{}
	mockCache.On("GetOrLoad", "provider1_routes", mock.AnythingOfType("time.Duration"), mock.AnythingOfType("func() (interface {}, error)")).Return(provider1Routes, nil)
	mockCache.On("GetOrLoad", "provider2_routes", mock.AnythingOfType("time.Duration"), mock.AnythingOfType("func() (interface {}, error)")).Return(provider2Routes, nil)

	cfg := createTestConfig(server1.URL, server2.URL)
	provider := New(cfg, mockCache)

	ctx := t.Context()
	filters := models.RouteFilters{}
	routes, err := provider.GetRoutes(ctx, filters)

	require.NoError(t, err)
	assert.Len(t, routes, 4)
	mockCache.AssertExpectations(t)
}

func TestProvider_GetRoutes_Provider2Fails(t *testing.T) {
	t.Parallel()

	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"airline": "AA", "sourceAirport": "JFK", "destinationAirport": "LAX", "codeShare": "Y", "stops": 0, "provider": "provider1"}]`))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server2.Close()

	cfg := createTestConfig(server1.URL, server2.URL)
	provider := New(cfg, cache.New())

	ctx := t.Context()
	filters := models.RouteFilters{}
	routes, err := provider.GetRoutes(ctx, filters)

	require.NoError(t, err)
	assert.Len(t, routes, 1)
}

func TestProvider_GetRoutes_BothProvidersFail(t *testing.T) {
	t.Parallel()

	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server2.Close()

	cfg := createTestConfig(server1.URL, server2.URL)
	provider := New(cfg, cache.New())

	ctx := t.Context()
	filters := models.RouteFilters{}
	routes, err := provider.GetRoutes(ctx, filters)

	require.NoError(t, err)
	assert.Empty(t, routes)
}

func TestProvider_CircuitBreakerFunctionality(t *testing.T) {
	t.Parallel()

	callCount := 0

	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server2.Close()

	cfg := createTestConfig(server1.URL, server2.URL)
	provider := New(cfg, cache.New())

	ctx := t.Context()
	filters := models.RouteFilters{}

	routes, err := provider.GetRoutes(ctx, filters)

	require.NoError(t, err)
	assert.Empty(t, routes)

	for range 5 {
		routes, err = provider.GetRoutes(ctx, filters)
		require.NoError(t, err)
		assert.Empty(t, routes)
	}

	assert.Equal(t, 3, callCount, "Server should not be called after circuit breaker opens")
}

func TestProvider_RetryFunctionality(t *testing.T) {
	t.Parallel()

	callCountServer1 := 0

	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCountServer1++
		if callCountServer1 <= 1 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"airline": "AA", "sourceAirport": "JFK", "destinationAirport": "LAX", "codeShare": "Y", "stops": 0, "provider": "provider1"}]`))
		}
	}))
	defer server1.Close()

	callCountServer2 := 0
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCountServer2++
		if callCountServer2 <= 1 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"airline": "UA", "sourceAirport": "JFK", "destinationAirport": "SFO", "codeShare": "N", "stops": 1, "provider": "provider2"}]`))
		}
	}))

	cfg := createTestConfig(server1.URL, server2.URL)
	provider := New(cfg, cache.New())

	ctx := t.Context()
	filters := models.RouteFilters{}

	routes, err := provider.GetRoutes(ctx, filters)
	require.NoError(t, err)

	assert.Len(t, routes, 2)
	assert.Equal(t, 2, callCountServer1)
	assert.Equal(t, 2, callCountServer2)
}

func TestProvider_CacheHit(t *testing.T) {
	t.Parallel()

	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Server should not be called on cache hit")
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Server should not be called on cache hit")
	}))
	defer server2.Close()

	provider1Routes := createMockRoutes("provider1")
	provider2Routes := createMockRoutes("provider2")

	mockCache := &MockCache{}
	mockCache.On("GetOrLoad", "provider1_routes", mock.AnythingOfType("time.Duration"), mock.AnythingOfType("func() (interface {}, error)")).Return(provider1Routes, nil)
	mockCache.On("GetOrLoad", "provider2_routes", mock.AnythingOfType("time.Duration"), mock.AnythingOfType("func() (interface {}, error)")).Return(provider2Routes, nil)

	cfg := createTestConfig(server1.URL, server2.URL)
	provider := New(cfg, mockCache)

	ctx := t.Context()
	filters := models.RouteFilters{}
	routes, err := provider.GetRoutes(ctx, filters)

	require.NoError(t, err)
	assert.Len(t, routes, 4)
}

func TestProvider_ApplyFilters(t *testing.T) {
	t.Parallel()

	cfg := createTestConfig("http://test1.com", "http://test2.com")
	provider := New(cfg, cache.New()).(provider)

	routes := []models.Route{
		{Airline: "AA", SourceAirport: "JFK", DestinationAirport: "LAX", Stops: 0, Provider: "provider1"},
		{Airline: "UA", SourceAirport: "JFK", DestinationAirport: "SFO", Stops: 1, Provider: "provider1"},
		{Airline: "AA", SourceAirport: "LAX", DestinationAirport: "JFK", Stops: 0, Provider: "provider2"},
		{Airline: "DL", SourceAirport: "JFK", DestinationAirport: "LAX", Stops: 2, Provider: "provider2"},
	}

	tests := []struct {
		name     string
		filters  models.RouteFilters
		expected int
	}{
		{
			name:     "No filters",
			filters:  models.RouteFilters{},
			expected: 4,
		},
		{
			name:     "Filter by airline",
			filters:  models.RouteFilters{Airline: "AA"},
			expected: 2,
		},
		{
			name:     "Filter by source airport",
			filters:  models.RouteFilters{SourceAirport: "JFK"},
			expected: 3,
		},
		{
			name:     "Filter by destination airport",
			filters:  models.RouteFilters{DestinationAirport: "LAX"},
			expected: 2,
		},
		{
			name: "Filter by max stops",
			filters: models.RouteFilters{MaxStops: func() *int {
				i := 1

				return &i
			}()},
			expected: 3,
		},
		{
			name:     "Filter with limit",
			filters:  models.RouteFilters{Limit: 2},
			expected: 2,
		},
		{
			name:     "Filter with offset",
			filters:  models.RouteFilters{Offset: 1, Limit: 2},
			expected: 2,
		},
		{
			name:     "Combined filters",
			filters:  models.RouteFilters{Airline: "AA", SourceAirport: "JFK", Limit: 1},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := provider.ApplyFilters(tt.filters, routes)
			assert.Len(t, result, tt.expected)
		})
	}
}
