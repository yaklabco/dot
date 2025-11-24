package dot_test

import (
	"context"
	"io/fs"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yaklabco/dot/pkg/dot"
)

// MockFS is a mock implementation of the FS interface for testing.
type MockFS struct {
	mock.Mock
}

func (m *MockFS) Stat(ctx context.Context, name string) (dot.FileInfo, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(dot.FileInfo), args.Error(1)
}

func (m *MockFS) ReadDir(ctx context.Context, name string) ([]dot.DirEntry, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dot.DirEntry), args.Error(1)
}

func (m *MockFS) ReadLink(ctx context.Context, name string) (string, error) {
	args := m.Called(ctx, name)
	return args.String(0), args.Error(1)
}

func (m *MockFS) ReadFile(ctx context.Context, name string) ([]byte, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockFS) WriteFile(ctx context.Context, name string, data []byte, perm fs.FileMode) error {
	args := m.Called(ctx, name, data, perm)
	return args.Error(0)
}

func (m *MockFS) Mkdir(ctx context.Context, name string, perm fs.FileMode) error {
	args := m.Called(ctx, name, perm)
	return args.Error(0)
}

func (m *MockFS) MkdirAll(ctx context.Context, name string, perm fs.FileMode) error {
	args := m.Called(ctx, name, perm)
	return args.Error(0)
}

func (m *MockFS) Remove(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockFS) RemoveAll(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockFS) Symlink(ctx context.Context, oldname, newname string) error {
	args := m.Called(ctx, oldname, newname)
	return args.Error(0)
}

func (m *MockFS) Rename(ctx context.Context, oldname, newname string) error {
	args := m.Called(ctx, oldname, newname)
	return args.Error(0)
}

func (m *MockFS) Exists(ctx context.Context, name string) bool {
	args := m.Called(ctx, name)
	return args.Bool(0)
}

func (m *MockFS) IsDir(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

func (m *MockFS) IsSymlink(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

func TestMockFS(t *testing.T) {
	ctx := context.Background()
	mockFS := new(MockFS)

	// Test Exists
	mockFS.On("Exists", ctx, "/test/file").Return(true)
	exists := mockFS.Exists(ctx, "/test/file")
	assert.True(t, exists)
	mockFS.AssertExpectations(t)
}

// MockLogger is a mock implementation of the Logger interface.
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(ctx context.Context, msg string, args ...any) {
	m.Called(ctx, msg, args)
}

func (m *MockLogger) Info(ctx context.Context, msg string, args ...any) {
	m.Called(ctx, msg, args)
}

func (m *MockLogger) Warn(ctx context.Context, msg string, args ...any) {
	m.Called(ctx, msg, args)
}

func (m *MockLogger) Error(ctx context.Context, msg string, args ...any) {
	m.Called(ctx, msg, args)
}

func (m *MockLogger) With(args ...any) dot.Logger {
	callArgs := m.Called(args)
	return callArgs.Get(0).(dot.Logger)
}

func TestMockLogger(t *testing.T) {
	ctx := context.Background()
	mockLogger := new(MockLogger)

	mockLogger.On("Info", ctx, "test message", mock.Anything).Return()
	mockLogger.Info(ctx, "test message", "key", "value")
	mockLogger.AssertExpectations(t)
}

// MockTracer is a mock implementation of the Tracer interface.
type MockTracer struct {
	mock.Mock
}

func (m *MockTracer) Start(ctx context.Context, name string, opts ...dot.SpanOption) (context.Context, dot.Span) {
	args := m.Called(ctx, name, opts)
	return args.Get(0).(context.Context), args.Get(1).(dot.Span)
}

// MockSpan is a mock implementation of the Span interface.
type MockSpan struct {
	mock.Mock
}

func (m *MockSpan) End() {
	m.Called()
}

func (m *MockSpan) RecordError(err error) {
	m.Called(err)
}

func (m *MockSpan) SetAttributes(attrs ...dot.Attribute) {
	m.Called(attrs)
}

func TestMockTracer(t *testing.T) {
	ctx := context.Background()
	mockTracer := new(MockTracer)
	mockSpan := new(MockSpan)

	mockTracer.On("Start", ctx, "test.operation", mock.Anything).Return(ctx, mockSpan)
	mockSpan.On("End").Return()

	newCtx, span := mockTracer.Start(ctx, "test.operation")
	assert.NotNil(t, newCtx)
	assert.NotNil(t, span)
	span.End()

	mockTracer.AssertExpectations(t)
	mockSpan.AssertExpectations(t)
}

// MockMetrics is a mock implementation of the Metrics interface.
type MockMetrics struct {
	mock.Mock
}

func (m *MockMetrics) Counter(name string, labels ...string) dot.Counter {
	args := m.Called(name, labels)
	return args.Get(0).(dot.Counter)
}

func (m *MockMetrics) Histogram(name string, labels ...string) dot.Histogram {
	args := m.Called(name, labels)
	return args.Get(0).(dot.Histogram)
}

func (m *MockMetrics) Gauge(name string, labels ...string) dot.Gauge {
	args := m.Called(name, labels)
	return args.Get(0).(dot.Gauge)
}

// MockCounter is a mock implementation of the Counter interface.
type MockCounter struct {
	mock.Mock
}

func (m *MockCounter) Inc(labels ...string) {
	m.Called(labels)
}

func (m *MockCounter) Add(delta float64, labels ...string) {
	m.Called(delta, labels)
}

// MockHistogram is a mock implementation of the Histogram interface.
type MockHistogram struct {
	mock.Mock
}

func (m *MockHistogram) Observe(value float64, labels ...string) {
	m.Called(value, labels)
}

// MockGauge is a mock implementation of the Gauge interface.
type MockGauge struct {
	mock.Mock
}

func (m *MockGauge) Set(value float64, labels ...string) {
	m.Called(value, labels)
}

func (m *MockGauge) Inc(labels ...string) {
	m.Called(labels)
}

func (m *MockGauge) Dec(labels ...string) {
	m.Called(labels)
}

func TestMockMetrics(t *testing.T) {
	mockMetrics := new(MockMetrics)
	mockCounter := new(MockCounter)

	mockMetrics.On("Counter", "operations", mock.Anything).Return(mockCounter)
	mockCounter.On("Inc", mock.Anything).Return()

	counter := mockMetrics.Counter("operations")
	counter.Inc()

	mockMetrics.AssertExpectations(t)
	mockCounter.AssertExpectations(t)
}

// MockFileInfo is a mock implementation of FileInfo.
type MockFileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
}

func (m MockFileInfo) Name() string       { return m.name }
func (m MockFileInfo) Size() int64        { return m.size }
func (m MockFileInfo) Mode() fs.FileMode  { return m.mode }
func (m MockFileInfo) ModTime() time.Time { return m.modTime }
func (m MockFileInfo) IsDir() bool        { return m.isDir }
func (m MockFileInfo) Sys() any           { return nil }

func TestMockFileInfo(t *testing.T) {
	info := MockFileInfo{
		name:    "test.txt",
		size:    100,
		mode:    0644,
		modTime: time.Now(),
		isDir:   false,
	}

	assert.Equal(t, "test.txt", info.Name())
	assert.Equal(t, int64(100), info.Size())
	assert.Equal(t, fs.FileMode(0644), info.Mode())
	assert.False(t, info.IsDir())
}
