package backend_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/garden"
	"github.com/concourse/concourse/worker/backend"
	"github.com/concourse/concourse/worker/backend/containerdadapter/containerdadapterfakes"
	"github.com/containerd/containerd"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ProcessSuite struct {
	suite.Suite
	*require.Assertions

	io                *containerdadapterfakes.FakeIO
	containerdProcess *containerdadapterfakes.FakeProcess
	ch                chan containerd.ExitStatus
	process           *backend.Process
}

func (s *ProcessSuite) SetupTest() {
	s.io = new(containerdadapterfakes.FakeIO)
	s.containerdProcess = new(containerdadapterfakes.FakeProcess)
	s.ch = make(chan containerd.ExitStatus, 1)

	s.process = backend.NewProcess(s.containerdProcess, s.ch)
}

func (s *ProcessSuite) TestID() {
	s.containerdProcess.IDReturns("id")
	id := s.process.ID()
	s.Equal("id", id)

	s.Equal(1, s.containerdProcess.IDCallCount())
}

func (s *ProcessSuite) TestWaitStatusErr() {
	expectedErr := errors.New("status-err")
	s.ch <- *containerd.NewExitStatus(0, time.Now(), expectedErr)

	_, err := s.process.Wait()
	s.True(errors.Is(err, expectedErr))
}

func (s *ProcessSuite) TestProcessWaitDeleteError() {
	s.ch <- *containerd.NewExitStatus(0, time.Now(), nil)

	expectedErr := errors.New("status-err")
	s.containerdProcess.DeleteReturns(nil, expectedErr)

	_, err := s.process.Wait()
	s.True(errors.Is(err, expectedErr))
}

func (s *ProcessSuite) TestProcessWaitBlocksUntilIOFinishes() {
	s.ch <- *containerd.NewExitStatus(0, time.Now(), nil)
	s.containerdProcess.IOReturns(s.io)

	_, err := s.process.Wait()
	s.NoError(err)

	s.Equal(1, s.containerdProcess.DeleteCallCount())
	s.Equal(1, s.containerdProcess.IOCallCount())
	s.Equal(1, s.io.WaitCallCount())
}

func (s *ProcessSuite) TestSetTTYWithNilWindowSize() {
	err := s.process.SetTTY(garden.TTYSpec{})
	s.NoError(err)
	s.Equal(0, s.containerdProcess.ResizeCallCount())
}

func (s *ProcessSuite) TestSetTTYResizeError() {
	expectedErr := errors.New("resize-err")
	s.containerdProcess.ResizeReturns(expectedErr)

	err := s.process.SetTTY(garden.TTYSpec{
		WindowSize: &garden.WindowSize{
			Columns: 123,
			Rows:    456,
		},
	})
	s.True(errors.Is(err, expectedErr))
}

func (s *ProcessSuite) TestSetTTYResize() {
	err := s.process.SetTTY(garden.TTYSpec{
		WindowSize: &garden.WindowSize{
			Columns: 123,
			Rows:    456,
		},
	})
	s.NoError(err)

	s.Equal(1, s.containerdProcess.ResizeCallCount())
	_, width, height := s.containerdProcess.ResizeArgsForCall(0)
	s.Equal(123, int(width))
	s.Equal(456, int(height))
}
