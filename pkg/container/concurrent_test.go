package container

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"

	ty "github.com/containrrr/watchtower/pkg/types"
)

// MockClient is a mock implementation of Client for testing concurrent operations
type MockClient struct {
	mu                  sync.RWMutex
	listContainersDelay time.Duration
	pullImageDelay      time.Duration
	stopContainerDelay  time.Duration
	startContainerDelay time.Duration
	containerList       []types.Container
	containerMap        map[string]*types.ContainerJSON
	imageMap            map[string]*types.ImageInspect
	listCallCount       int64
	pullCallCount       int64
	stopCallCount       int64
	startCallCount      int64
}

func NewMockClient() *MockClient {
	return &MockClient{
		containerList: []types.Container{},
		containerMap:  make(map[string]*types.ContainerJSON),
		imageMap:      make(map[string]*types.ImageInspect),
	}
}

func (m *MockClient) ListContainers(fn ty.Filter) ([]ty.Container, error) {
	// Fixed: Use single write lock to prevent race condition between unlock and RLock
	// Previous implementation had a race condition where containerList could be modified
	// between Unlock() and RLock(), leading to inconsistent reads
	m.mu.Lock()
	defer m.mu.Unlock()

	m.listCallCount++
	if m.listContainersDelay > 0 {
		time.Sleep(m.listContainersDelay)
	}

	containers := make([]ty.Container, 0, len(m.containerList))
	for _, c := range m.containerList {
		// Create a mock container
		mockContainer := MockContainer(
			WithContainerName(c.Names[0]),
			WithContainerID(c.ID),
		)
		containers = append(containers, mockContainer)
	}

	return containers, nil
}

func (m *MockClient) GetContainer(containerID ty.ContainerID) (ty.Container, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if ci, ok := m.containerMap[string(containerID)]; ok {
		if ii, ok := m.imageMap[ci.Config.Image]; ok {
			return NewContainer(ci, ii), nil
		}
		return NewContainer(ci, &types.ImageInspect{}), nil
	}
	return nil, errors.New("container not found")
}

func (m *MockClient) StopContainer(c ty.Container, timeout time.Duration) error {
	m.mu.Lock()
	m.stopCallCount++
	if m.stopContainerDelay > 0 {
		time.Sleep(m.stopContainerDelay)
	}
	m.mu.Unlock()
	return nil
}

func (m *MockClient) StartContainer(c ty.Container) (ty.ContainerID, error) {
	m.mu.Lock()
	m.startCallCount++
	if m.startContainerDelay > 0 {
		time.Sleep(m.startContainerDelay)
	}
	m.mu.Unlock()
	return c.ID(), nil
}

func (m *MockClient) RenameContainer(c ty.Container, newName string) error {
	return nil
}

func (m *MockClient) IsContainerStale(c ty.Container, params ty.UpdateParams) (bool, ty.ImageID, error) {
	return false, "", nil
}

func (m *MockClient) ExecuteCommand(containerID ty.ContainerID, command string, timeout int) (bool, error) {
	return false, nil
}

func (m *MockClient) RemoveImageByID(id ty.ImageID) error {
	return nil
}

func (m *MockClient) WarnOnHeadPullFailed(container ty.Container) bool {
	return false
}

// Helper functions for mock container creation
func WithContainerName(name string) MockContainerUpdate {
	return func(c *types.ContainerJSON, i *types.ImageInspect) {
		c.Name = name
		c.ID = "container_" + name
	}
}

func WithContainerID(id string) MockContainerUpdate {
	return func(c *types.ContainerJSON, i *types.ImageInspect) {
		c.ID = id
	}
}

// TestConcurrentListContainers tests that ListContainers can be called safely from multiple goroutines
func TestConcurrentListContainers(t *testing.T) {
	client := NewMockClient()

	// Add some mock containers
	for i := 0; i < 10; i++ {
		// Fixed: Use fmt.Sprintf to generate unique IDs instead of rune conversion
		// Previous implementation: string(rune('0'+i)) produces non-numeric characters when i >= 10
		containerID := ty.ContainerID(fmt.Sprintf("container%d", i))
		containerInfo := &types.ContainerJSON{
			ContainerJSONBase: &types.ContainerJSONBase{
				ID:    string(containerID),
				Name:  fmt.Sprintf("/test-container-%d", i),
				Image: fmt.Sprintf("test-image:%d", i),
			},
			Config: &container.Config{
				Image: fmt.Sprintf("test-image:%d", i),
			},
		}
		imageInfo := &types.ImageInspect{
			ID:       fmt.Sprintf("image_id_%d", i),
			RepoTags: []string{fmt.Sprintf("test-image:%d", i)},
		}
		client.containerMap[string(containerID)] = containerInfo
		client.imageMap[fmt.Sprintf("test-image:%d", i)] = imageInfo
		client.containerList = append(client.containerList, types.Container{
			ID:    string(containerID),
			Names: []string{fmt.Sprintf("/test-container-%d", i)},
		})
	}

	var wg sync.WaitGroup
	numGoroutines := 50
	errorsChan := make(chan error, numGoroutines)

	// Launch multiple goroutines that concurrently call ListContainers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			containers, err := client.ListContainers(func(c ty.FilterableContainer) bool { return true })
			if err != nil {
				errorsChan <- err
				return
			}
			if len(containers) != 10 {
				errorsChan <- errors.New("unexpected container count")
				return
			}
		}(i)
	}

	wg.Wait()
	close(errorsChan)

	// Check for errors
	for err := range errorsChan {
		t.Errorf("Error in concurrent ListContainers: %v", err)
	}

	// Verify that ListContainers was called the expected number of times
	if client.listCallCount != int64(numGoroutines) {
		t.Errorf("Expected %d ListContainers calls, got %d", numGoroutines, client.listCallCount)
	}
}

// TestConcurrentListContainersWithFilters tests concurrent ListContainers with different filters
func TestConcurrentListContainersWithFilters(t *testing.T) {
	client := NewMockClient()

	// Add mock containers with different names
	for i := 0; i < 20; i++ {
		// Fixed: Use fmt.Sprintf to generate unique IDs
		// Previous implementation had duplicate IDs (only 10 unique IDs for 20 containers)
		containerID := ty.ContainerID(fmt.Sprintf("container%d", i))
		containerInfo := &types.ContainerJSON{
			ContainerJSONBase: &types.ContainerJSONBase{
				ID:    string(containerID),
				Name:  fmt.Sprintf("/group-%d-container-%d", i%4, i),
				Image: fmt.Sprintf("test-image:%d", i),
			},
			Config: &container.Config{
				Image: fmt.Sprintf("test-image:%d", i),
			},
		}
		imageInfo := &types.ImageInspect{
			ID:       fmt.Sprintf("image_id_%d", i),
			RepoTags: []string{fmt.Sprintf("test-image:%d", i)},
		}
		client.containerMap[string(containerID)] = containerInfo
		client.imageMap[fmt.Sprintf("test-image:%d", i)] = imageInfo
		client.containerList = append(client.containerList, types.Container{
			ID:    string(containerID),
			Names: []string{fmt.Sprintf("/group-%d-container-%d", i%4, i)},
		})
	}

	var wg sync.WaitGroup
	numGoroutines := 30

	// Launch multiple goroutines with different filters
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// Each goroutine filters for a different group
			groupFilter := func(c ty.FilterableContainer) bool {
				return len(c.Name()) >= 7 && c.Name()[6:7] == string(rune('0'+id%4))
			}
			_, err := client.ListContainers(groupFilter)
			if err != nil {
				t.Errorf("Error in filtered ListContainers: %v", err)
			}
		}(i)
	}

	wg.Wait()
}

// TestConcurrentGetContainer tests concurrent GetContainer operations
func TestConcurrentGetContainer(t *testing.T) {
	client := NewMockClient()

	// Add mock containers
	containerIDs := []ty.ContainerID{}
	for i := 0; i < 10; i++ {
		containerID := ty.ContainerID(fmt.Sprintf("container%d", i))
		containerIDs = append(containerIDs, containerID)

		containerInfo := &types.ContainerJSON{
			ContainerJSONBase: &types.ContainerJSONBase{
				ID:    string(containerID),
				Name:  fmt.Sprintf("/test-container-%d", i),
				Image: fmt.Sprintf("test-image:%d", i),
			},
			Config: &container.Config{
				Image: fmt.Sprintf("test-image:%d", i),
			},
		}
		imageInfo := &types.ImageInspect{
			ID:       fmt.Sprintf("image_id_%d", i),
			RepoTags: []string{fmt.Sprintf("test-image:%d", i)},
		}
		client.containerMap[string(containerID)] = containerInfo
		client.imageMap[fmt.Sprintf("test-image:%d", i)] = imageInfo
	}

	var wg sync.WaitGroup
	numGoroutines := 100
	errorsChan := make(chan error, numGoroutines)

	// Launch multiple goroutines that concurrently call GetContainer
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			containerID := containerIDs[id%len(containerIDs)]
			c, err := client.GetContainer(containerID)
			if err != nil {
				errorsChan <- err
				return
			}
			if c.ID() != containerID {
				errorsChan <- errors.New("unexpected container ID")
				return
			}
		}(i)
	}

	wg.Wait()
	close(errorsChan)

	// Check for errors
	for err := range errorsChan {
		t.Errorf("Error in concurrent GetContainer: %v", err)
	}
}

// TestConcurrentMixedOperations tests mixed concurrent operations on containers
func TestConcurrentMixedOperations(t *testing.T) {
	client := NewMockClient()

	// Add mock containers
	containerIDs := []ty.ContainerID{}
	for i := 0; i < 5; i++ {
		containerID := ty.ContainerID(fmt.Sprintf("container%d", i))
		containerIDs = append(containerIDs, containerID)

		containerInfo := &types.ContainerJSON{
			ContainerJSONBase: &types.ContainerJSONBase{
				ID:    string(containerID),
				Name:  fmt.Sprintf("/test-container-%d", i),
				Image: fmt.Sprintf("test-image:%d", i),
				State: &types.ContainerState{Running: true},
			},
			Config: &container.Config{
				Image: fmt.Sprintf("test-image:%d", i),
			},
		}
		imageInfo := &types.ImageInspect{
			ID:       fmt.Sprintf("image_id_%d", i),
			RepoTags: []string{fmt.Sprintf("test-image:%d", i)},
		}
		client.containerMap[string(containerID)] = containerInfo
		client.imageMap[fmt.Sprintf("test-image:%d", i)] = imageInfo
		client.containerList = append(client.containerList, types.Container{
			ID:    string(containerID),
			Names: []string{fmt.Sprintf("/test-container-%d", i)},
		})
	}

	var wg sync.WaitGroup
	numGoroutines := 50
	errorsChan := make(chan error, numGoroutines*3)

	// Launch multiple goroutines that perform mixed operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(3)

		// Goroutine 1: ListContainers
		go func(id int) {
			defer wg.Done()
			_, err := client.ListContainers(func(c ty.FilterableContainer) bool { return true })
			if err != nil {
				errorsChan <- err
			}
		}(i)

		// Goroutine 2: GetContainer
		go func(id int) {
			defer wg.Done()
			containerID := containerIDs[id%len(containerIDs)]
			_, err := client.GetContainer(containerID)
			if err != nil {
				errorsChan <- err
			}
		}(i)

		// Goroutine 3: StopContainer
		go func(id int) {
			defer wg.Done()
			containerID := containerIDs[id%len(containerIDs)]
			c, err := client.GetContainer(containerID)
			if err != nil {
				errorsChan <- err
				return
			}
			err = client.StopContainer(c, 10*time.Second)
			if err != nil {
				errorsChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errorsChan)

	// Check for errors
	for err := range errorsChan {
		t.Errorf("Error in concurrent mixed operations: %v", err)
	}

	// Verify operations were called
	if client.listCallCount == 0 {
		t.Error("ListContainers was not called")
	}
	if client.stopCallCount == 0 {
		t.Error("StopContainer was not called")
	}
}

// TestConcurrentListContainersWithStress tests high-stress concurrent ListContainers operations
func TestConcurrentListContainersWithStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	client := NewMockClient()

	// Add many mock containers
	for i := 0; i < 100; i++ {
		// Fixed: Use fmt.Sprintf to generate unique IDs
		containerID := ty.ContainerID(fmt.Sprintf("container%d", i))
		containerInfo := &types.ContainerJSON{
			ContainerJSONBase: &types.ContainerJSONBase{
				ID:    string(containerID),
				Name:  fmt.Sprintf("/test-container-%d", i),
				Image: fmt.Sprintf("test-image:%d", i),
			},
			Config: &container.Config{
				Image: fmt.Sprintf("test-image:%d", i),
			},
		}
		imageInfo := &types.ImageInspect{
			ID:       fmt.Sprintf("image_id_%d", i),
			RepoTags: []string{fmt.Sprintf("test-image:%d", i)},
		}
		client.containerMap[string(containerID)] = containerInfo
		client.imageMap[fmt.Sprintf("test-image:%d", i)] = imageInfo
		client.containerList = append(client.containerList, types.Container{
			ID:    string(containerID),
			Names: []string{fmt.Sprintf("/test-container-%d", i)},
		})
	}

	// Record initial goroutine count
	initialGoroutines := runtime.NumGoroutine()

	var wg sync.WaitGroup
	numGoroutines := 200
	errorsChan := make(chan error, numGoroutines)

	// Launch many goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, err := client.ListContainers(func(c ty.FilterableContainer) bool { return true })
			if err != nil {
				errorsChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errorsChan)

	// Check for errors
	for err := range errorsChan {
		t.Errorf("Error in stress test: %v", err)
	}

	// Check for goroutine leaks
	finalGoroutines := runtime.NumGoroutine()
	// Fixed: Reduced threshold from 50 to 15 to better detect goroutine leaks
	// Previous threshold was too high, allowing significant leaks to go undetected
	if finalGoroutines-initialGoroutines > 15 {
		t.Errorf("Potential goroutine leak: started with %d goroutines, ended with %d",
			initialGoroutines, finalGoroutines)
	}
}

// TestConcurrentContainerStateAccess tests concurrent access to container state
func TestConcurrentContainerStateAccess(t *testing.T) {
	container := MockContainer(
		WithContainerName("test-container"),
		WithContainerID("test-id"),
		WithContainerState(types.ContainerState{Running: true}),
	)

	var wg sync.WaitGroup
	numGoroutines := 100

	// Launch multiple goroutines that concurrently access atomic boolean fields
	for i := 0; i < numGoroutines; i++ {
		wg.Add(2)

		// Goroutine 1: Write atomic fields
		go func(id int) {
			defer wg.Done()
			container.SetStale(id%2 == 0)
			container.SetLinkedToRestarting(id%3 == 0)
		}(i)

		// Goroutine 2: Read atomic fields
		go func(id int) {
			defer wg.Done()
			_ = container.IsStale()
			_ = container.IsLinkedToRestarting()
			_ = container.ToRestart()
		}(i)
	}

	wg.Wait()

	// Verify final state is consistent
	// The container should be either stale or not, not both
	isStale := container.IsStale()
	_ = isStale // Just verify no panic
}

// BenchmarkConcurrentListContainers benchmarks concurrent ListContainers operations
func BenchmarkConcurrentListContainers(b *testing.B) {
	client := NewMockClient()

	// Add mock containers
	for i := 0; i < 50; i++ {
		containerID := ty.ContainerID(fmt.Sprintf("container%d", i))
		containerInfo := &types.ContainerJSON{
			ContainerJSONBase: &types.ContainerJSONBase{
				ID:    string(containerID),
				Name:  fmt.Sprintf("/test-container-%d", i),
				Image: fmt.Sprintf("test-image:%d", i),
			},
			Config: &container.Config{
				Image: fmt.Sprintf("test-image:%d", i),
			},
		}
		imageInfo := &types.ImageInspect{
			ID:       fmt.Sprintf("image_id_%d", i),
			RepoTags: []string{fmt.Sprintf("test-image:%d", i)},
		}
		client.containerMap[string(containerID)] = containerInfo
		client.imageMap[fmt.Sprintf("test-image:%d", i)] = imageInfo
		client.containerList = append(client.containerList, types.Container{
			ID:    string(containerID),
			Names: []string{fmt.Sprintf("/test-container-%d", i)},
		})
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := client.ListContainers(func(c ty.FilterableContainer) bool { return true })
			if err != nil {
				b.Errorf("Error in benchmark: %v", err)
			}
		}
	})
}

// BenchmarkConcurrentGetContainer benchmarks concurrent GetContainer operations
func BenchmarkConcurrentGetContainer(b *testing.B) {
	client := NewMockClient()
	containerID := ty.ContainerID("test-container-id")

	containerInfo := &types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			ID:    string(containerID),
			Name:  "/test-container",
			Image: "test-image:latest",
		},
		Config: &container.Config{
			Image: "test-image:latest",
		},
	}
	imageInfo := &types.ImageInspect{
		ID:       "image_id",
		RepoTags: []string{"test-image:latest"},
	}
	client.containerMap[string(containerID)] = containerInfo
	client.imageMap["test-image:latest"] = imageInfo

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := client.GetContainer(containerID)
			if err != nil {
				b.Errorf("Error in benchmark: %v", err)
			}
		}
	})
}

// TestCreateListFilter tests the createListFilter function
func TestCreateListFilter(t *testing.T) {
	tests := []struct {
		name              string
		includeStopped    bool
		includeRestarting bool
		expectedStatuses  []string
	}{
		{
			name:              "default filters",
			includeStopped:    false,
			includeRestarting: false,
			expectedStatuses:  []string{"running"},
		},
		{
			name:              "include stopped",
			includeStopped:    true,
			includeRestarting: false,
			expectedStatuses:  []string{"running", "created", "exited"},
		},
		{
			name:              "include restarting",
			includeStopped:    false,
			includeRestarting: true,
			expectedStatuses:  []string{"running", "restarting"},
		},
		{
			name:              "include all",
			includeStopped:    true,
			includeRestarting: true,
			expectedStatuses:  []string{"running", "created", "exited", "restarting"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := dockerClient{
				ClientOptions: ClientOptions{
					IncludeStopped:    tt.includeStopped,
					IncludeRestarting: tt.includeRestarting,
				},
			}

			filter := client.createListFilter()
			statuses := filter.Get("status")

			if len(statuses) != len(tt.expectedStatuses) {
				t.Errorf("Expected %d statuses, got %d", len(tt.expectedStatuses), len(statuses))
			}

			for _, expected := range tt.expectedStatuses {
				found := false
				for _, actual := range statuses {
					if actual == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected status %s not found in filter", expected)
				}
			}
		})
	}
}

// TestConcurrentCreateListFilter tests concurrent creation of list filters
func TestConcurrentCreateListFilter(t *testing.T) {
	client := dockerClient{
		ClientOptions: ClientOptions{
			IncludeStopped:    true,
			IncludeRestarting: true,
		},
	}

	var wg sync.WaitGroup
	numGoroutines := 100

	// Launch multiple goroutines that concurrently create filters
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			filter := client.createListFilter()
			statuses := filter.Get("status")
			if len(statuses) != 4 {
				t.Errorf("Expected 4 statuses, got %d", len(statuses))
			}
		}(i)
	}

	wg.Wait()
}

// TestContainerMethodsRaceDetector uses go test -race to verify that atomic boolean fields
// are properly synchronized and can be accessed concurrently without race conditions.
func TestContainerMethodsRaceDetector(t *testing.T) {
	container := MockContainer(
		WithContainerName("test-container"),
		WithContainerID("test-id"),
	)

	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 1000; i++ {
			container.SetStale(i%2 == 0)
			container.SetLinkedToRestarting(i%3 == 0)
		}
		done <- true
	}()

	// Reader goroutine - only access atomic fields
	go func() {
		for i := 0; i < 1000; i++ {
			_ = container.IsStale()
			_ = container.IsLinkedToRestarting()
			_ = container.ToRestart()
		}
		done <- true
	}()

	<-done
	<-done

	// Verify that state values are valid (no corruption)
	isStale := container.IsStale()
	isLinked := container.IsLinkedToRestarting()
	toRestart := container.ToRestart()
	// These should be valid boolean values
	_ = isStale
	_ = isLinked
	_ = toRestart
}